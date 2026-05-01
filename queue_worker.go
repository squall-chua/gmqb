package gmqb

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Handler is a function that processes a message payload.
type Handler[T any] func(ctx context.Context, payload T) error

// Worker consumes messages from a Queue.
type Worker[T any] struct {
	q       *Queue[T]
	handler Handler[T]
	config  WorkerConfig
	id      string
}

// NewWorker creates a new Worker[T] for the given Queue and handler.
func NewWorker[T any](q *Queue[T], handler Handler[T], opts ...WorkerOpt) *Worker[T] {
	cfg := WorkerConfig{
		Delivery:          AtLeastOnce,
		Concurrency:       1,
		VisibilityTimeout: q.opts.VisibilityTimeout,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	return &Worker[T]{
		q:       q,
		handler: handler,
		config:  cfg,
		id:      bson.NewObjectID().Hex(),
	}
}

// Run starts the worker and blocks until the context is cancelled.
func (w *Worker[T]) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	// Start reaper goroutine for AtLeastOnce and ExactlyOnce modes
	if w.config.Delivery != AtMostOnce {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w.reapLoop(ctx)
		}()
	}

	// Start change stream wake-up if enabled
	if w.config.ChangeStreamWake {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w.watchLoop(ctx)
		}()
	}

	// Start worker goroutines
	for i := 0; i < w.config.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w.pollLoop(ctx)
		}()
	}

	wg.Wait()
	return nil
}

func (w *Worker[T]) watchLoop(ctx context.Context) {
	// Simple watch for inserts to wake up workers
	opts := options.ChangeStream().SetFullDocument(options.UpdateLookup)
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"operationType": "insert"}}},
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		stream, err := w.q.coll.Watch(ctx, pipeline, opts)
		if err != nil {
			time.Sleep(time.Second) // Simple backoff for watch errors
			continue
		}

		for stream.Next(ctx) {
			select {
			case w.q.wake <- struct{}{}:
			default:
			}
		}

		_ = stream.Close(ctx)
	}
}

func (w *Worker[T]) pollLoop(ctx context.Context) {
	interval := 50 * time.Millisecond
	const maxInterval = 500 * time.Millisecond

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		doc, err := w.claim(ctx)
		if err != nil {
			// No messages or error
			select {
			case <-ctx.Done():
				return
			case <-w.q.wake:
				interval = 50 * time.Millisecond
			case <-time.After(interval):
				interval *= 2
				if interval > maxInterval {
					interval = maxInterval
				}
			}
			continue
		}

		// Reset interval on success
		interval = 50 * time.Millisecond

		w.process(ctx, doc)
	}
}

func (w *Worker[T]) claim(ctx context.Context) (*queueDoc[T], error) {
	if w.config.ConsumerGroup != "" {
		return claimFanOut[T](ctx, w.q.coll, w.q.offsets, w.q.claims, w.config.ConsumerGroup, w.id, w.config.VisibilityTimeout)
	}

	targetStatus := statusProcessing
	if w.config.Delivery == AtMostOnce {
		targetStatus = statusDone
	}

	return claimOne[T](ctx, w.q.coll, w.id, bson.NilObjectID, w.config.VisibilityTimeout, targetStatus)
}

func (w *Worker[T]) process(ctx context.Context, doc *queueDoc[T]) {
	if w.config.Delivery == ExactlyOnce {
		group := w.config.ConsumerGroup
		if group == "" {
			group = "default"
		}
		dedupKey := fmt.Sprintf("%s:%s", doc.ID.Hex(), group)

		_, err := w.q.dedup.InsertOne(ctx, dedupDoc{
			ID:        dedupKey,
			CreatedAt: time.Now(),
		})

		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				// Already processed by another worker or previously
				if w.config.ConsumerGroup != "" {
					_ = ackFanOut(ctx, w.q.claims, w.config.ConsumerGroup, doc.ID)
				} else {
					_ = ackOne(ctx, w.q.coll, doc.ID)
				}
				return
			}
			// Transient error inserting dedup key, nack and retry
			_ = nackOne(ctx, w.q.coll, doc.ID)
			return
		}
	}

	err := w.handler(ctx, doc.Payload)

	if w.config.Delivery == AtMostOnce {
		return // Already done via claimOne targetStatus
	}

	if err != nil {
		if doc.Attempts >= doc.MaxAttempts {
			if w.config.ConsumerGroup != "" {
				_ = moveToDLQFanOut[T](ctx, w.q.claims, w.q.dlqColl, w.config.ConsumerGroup, doc)
			} else {
				_ = moveToDLQ(ctx, w.q.coll, w.q.dlqColl, doc)
			}
		} else {
			if w.config.ConsumerGroup != "" {
				_ = nackFanOut(ctx, w.q.claims, w.config.ConsumerGroup, doc.ID)
			} else {
				_ = nackOne(ctx, w.q.coll, doc.ID)
			}
		}
		return
	}

	if w.config.ConsumerGroup != "" {
		_ = ackFanOut(ctx, w.q.claims, w.config.ConsumerGroup, doc.ID)
	} else {
		_ = ackOne(ctx, w.q.coll, doc.ID)
	}
}

func (w *Worker[T]) reapLoop(ctx context.Context) {
	ticker := time.NewTicker(w.config.VisibilityTimeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.reap(ctx)
		}
	}
}

func (w *Worker[T]) reap(ctx context.Context) {
	now := time.Now()
	staleThreshold := now.Add(-w.config.VisibilityTimeout)

	// 1. Reap Load-Balanced messages (primary collection)
	if w.config.ConsumerGroup == "" {
		filter := bson.M{
			"s":  statusProcessing,
			"ca": bson.M{"$lt": staleThreshold},
			"a":  bson.M{"$gte": w.q.opts.MaxAttempts},
		}

		cursor, err := w.q.coll.Find(ctx, filter)
		if err == nil {
			defer cursor.Close(ctx)
			for cursor.Next(ctx) {
				var doc queueDoc[T]
				if err := cursor.Decode(&doc); err == nil {
					_ = moveToDLQ[T](ctx, w.q.coll, w.q.dlqColl, &doc)
				}
			}
		}
	} else {
		// 2. Reap Fan-Out group claims (claims collection)
		filter := bson.M{
			"_id.g": w.config.ConsumerGroup,
			"s":     statusProcessing,
			"ca":    bson.M{"$lt": staleThreshold},
			"a":     bson.M{"$gte": w.q.opts.MaxAttempts},
		}

		cursor, err := w.q.claims.Find(ctx, filter)
		if err == nil {
			defer cursor.Close(ctx)
			for cursor.Next(ctx) {
				var claim groupClaimDoc
				if err := cursor.Decode(&claim); err == nil {
					// Fetch message to move to DLQ
					var doc queueDoc[T]
					if err := w.q.coll.FindOne(ctx, bson.M{"_id": claim.ID.MsgID}).Decode(&doc); err == nil {
						_ = moveToDLQFanOut[T](ctx, w.q.claims, w.q.dlqColl, w.config.ConsumerGroup, &doc)
					}
				}
			}
		}
	}
}
