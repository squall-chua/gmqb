package gmqb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Queue is a durable message queue backed by MongoDB.
type Queue[T any] struct {
	db   *mongo.Database
	name string
	opts QueueOpts

	coll    *mongo.Collection
	dlqColl *mongo.Collection
	dedup   *mongo.Collection
	offsets *mongo.Collection
	claims  *mongo.Collection

	wake chan struct{}
}

// NewQueue creates a new Queue[T] with the given database and name.
// It initializes the collections but does not create indexes (use EnsureIndexes for that).
func NewQueue[T any](db *mongo.Database, name string, opts QueueOpts) (*Queue[T], error) {
	if opts.MaxAttempts <= 0 {
		opts.MaxAttempts = DefaultQueueOpts.MaxAttempts
	}
	if opts.VisibilityTimeout <= 0 {
		opts.VisibilityTimeout = DefaultQueueOpts.VisibilityTimeout
	}
	if opts.DedupTTL <= 0 {
		opts.DedupTTL = DefaultQueueOpts.DedupTTL
	}
	if opts.DLQTTL <= 0 {
		opts.DLQTTL = DefaultQueueOpts.DLQTTL
	}
	if opts.RetentionTTL <= 0 {
		opts.RetentionTTL = DefaultQueueOpts.RetentionTTL
	}

	q := &Queue[T]{
		db:      db,
		name:    name,
		opts:    opts,
		coll:    db.Collection(primaryCollName(name)),
		dlqColl: db.Collection(dlqCollName(name)),
		dedup:   db.Collection(dedupCollName(name)),
		offsets: db.Collection(offsetCollName(name)),
		claims:  db.Collection(claimsCollName(name)),
		wake:    make(chan struct{}, 1),
	}

	return q, nil
}

// Enqueue adds a new message to the queue.
func (q *Queue[T]) Enqueue(ctx context.Context, payload T, opts ...EnqueueOpt) (bson.ObjectID, error) {
	cfg := EnqueueConfig{
		ID: bson.NewObjectID(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	doc := queueDoc[T]{
		ID:          cfg.ID,
		Payload:     payload,
		Status:      statusPending,
		Attempts:    0,
		MaxAttempts: q.opts.MaxAttempts,
		CreatedAt:   time.Now(),
	}

	_, err := q.coll.InsertOne(ctx, doc)
	if err != nil {
		if cfg.Idempotent && mongo.IsDuplicateKeyError(err) {
			return cfg.ID, nil
		}
		return bson.NilObjectID, fmt.Errorf("gmqb queue: enqueue: %w", err)
	}

	// Trigger wake channel for in-process workers
	select {
	case q.wake <- struct{}{}:
	default:
	}

	return doc.ID, nil
}

// EnsureIndexes creates the necessary indexes for the queue to operate efficiently.
// This should be called once during application startup.
func (q *Queue[T]) EnsureIndexes(ctx context.Context) error {
	// Primary collection index for claiming messages
	// Filter: status="pending" or (status="processing" and claimedAt < now - visibilityTimeout)
	// Compound index on {status: 1, createdAt: 1} or {status: 1, _id: 1}
	_, err := q.coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "s", Value: 1},
			{Key: "_id", Value: 1},
		},
		Options: options.Index().SetName("gmqb_queue_claim"),
	})
	if err != nil {
		return fmt.Errorf("gmqb queue: ensure primary indexes: %w", err)
	}

	// Primary collection TTL index
	_, err = q.coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "t", Value: 1}},
		Options: options.Index().
			SetExpireAfterSeconds(int32(q.opts.RetentionTTL.Seconds())).
			SetName("gmqb_queue_ttl"),
	})
	if err != nil {
		return fmt.Errorf("gmqb queue: ensure primary ttl index: %w", err)
	}

	// Deduplication collection TTL index
	_, err = q.dedup.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "t", Value: 1}},
		Options: options.Index().
			SetExpireAfterSeconds(int32(q.opts.DedupTTL.Seconds())).
			SetName("gmqb_queue_dedup_ttl"),
	})
	if err != nil {
		return fmt.Errorf("gmqb queue: ensure dedup indexes: %w", err)
	}

	// DLQ TTL index (if enabled)
	if q.opts.DLQEnabled {
		_, err = q.dlqColl.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{{Key: "t", Value: 1}},
			Options: options.Index().
				SetExpireAfterSeconds(int32(q.opts.DLQTTL.Seconds())).
				SetName("gmqb_queue_dlq_ttl"),
		})
		if err != nil {
			return fmt.Errorf("gmqb queue: ensure dlq indexes: %w", err)
		}
	}

	// Group claims index for reaping stale claims
	_, err = q.claims.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "s", Value: 1},
				{Key: "ca", Value: 1},
			},
			Options: options.Index().SetName("gmqb_queue_claims_reap"),
		},
		{
			Keys: bson.D{{Key: "t", Value: 1}},
			Options: options.Index().
				SetExpireAfterSeconds(int32(q.opts.RetentionTTL.Seconds())).
				SetName("gmqb_queue_claims_ttl"),
		},
	})
	if err != nil {
		return fmt.Errorf("gmqb queue: ensure claims indexes: %w", err)
	}

	return nil
}

// DLQ returns a accessor for the dead-letter queue.
func (q *Queue[T]) DLQ() *DLQ[T] {
	return &DLQ[T]{
		q:    q,
		coll: q.dlqColl,
	}
}
