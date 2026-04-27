package gmqb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// CappedOpts configures the capped collection that backs a topic.
type CappedOpts struct {
	SizeBytes int64 // ring-buffer byte limit (required; MongoDB min is 4096)
	MaxDocs   int64 // optional document count cap (0 = no limit)
}

// DefaultCappedOpts is a sensible default: 5 MB ring buffer, no doc limit.
var DefaultCappedOpts = CappedOpts{SizeBytes: 5 * 1024 * 1024}

// TailablePubSub provides a pub/sub mechanism using a MongoDB capped collection
// and tailable cursors. It is suitable for standalone MongoDB deployments where
// change streams are not available.
type TailablePubSub[T any] struct {
	coll *mongo.Collection
	opts CappedOpts
}

// NewTailablePubSub creates a TailablePubSub that uses a capped collection named topic.
// It idempotently creates the collection with the provided options if it doesn't exist.
//
// Example:
//
//	bus, err := gmqb.NewTailablePubSub[MyEvent](db, "my_topic", gmqb.DefaultCappedOpts)
func NewTailablePubSub[T any](db *mongo.Database, topic string, opts CappedOpts) (*TailablePubSub[T], error) {
	if opts.SizeBytes < 4096 {
		return nil, fmt.Errorf("gmqb pubsub: SizeBytes must be at least 4096")
	}

	createOpts := options.CreateCollection().
		SetCapped(true).
		SetSizeInBytes(opts.SizeBytes)
	if opts.MaxDocs > 0 {
		createOpts.SetMaxDocuments(opts.MaxDocs)
	}

	err := db.CreateCollection(context.Background(), topic, createOpts)
	if err != nil {
		// Ignore if collection already exists
		if !mongo.IsDuplicateKeyError(err) {
			// In some driver versions/server versions, it might not be a duplicate key error
			// but a CommandError with code 48 (NamespaceExists).
			if ce, ok := err.(mongo.CommandError); !ok || ce.Code != 48 {
				return nil, fmt.Errorf("gmqb pubsub: create capped collection: %w", err)
			}
		}
	}

	return &TailablePubSub[T]{
		coll: db.Collection(topic),
		opts: opts,
	}, nil
}

// envelope wraps the user payload so we can extend metadata later without
// breaking the BSON layout of existing documents.
type envelope[T any] struct {
	Payload T `bson:"p"`
}

// Publish inserts a typed document into the capped collection.
func (ps *TailablePubSub[T]) Publish(ctx context.Context, event T) error {
	_, err := ps.coll.InsertOne(ctx, envelope[T]{Payload: event})
	return err
}

// Subscribe returns a channel that receives events from the topic, and a
// cancel function to stop the subscription.
//
// Reconnects automatically with exponential back-off (max 30 s) if the
// tailable cursor dies.
//
// Important: always call the returned CancelFunc to avoid goroutine leaks.
func (ps *TailablePubSub[T]) Subscribe(ctx context.Context) (<-chan T, context.CancelFunc) {
	subCtx, cancel := context.WithCancel(ctx)
	ch := make(chan T, 64) // buffered to absorb bursts

	go func() {
		defer close(ch)
		ps.tail(subCtx, ch)
	}()

	return ch, cancel
}

func (ps *TailablePubSub[T]) tail(ctx context.Context, ch chan<- T) {
	backoff := time.Second
	const maxBackoff = 30 * time.Second

	for {
		if ctx.Err() != nil {
			return
		}

		ps.tailOnce(ctx, ch)

		if ctx.Err() != nil {
			// Context cancelled — clean exit.
			return
		}

		// Transient error or cursor exhaustion — back off and retry.
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}

		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}

func (ps *TailablePubSub[T]) tailOnce(ctx context.Context, ch chan<- T) {
	findOpts := options.Find().
		SetCursorType(options.TailableAwait).
		SetNoCursorTimeout(true).
		SetMaxAwaitTime(2 * time.Second)

	cursor, err := ps.coll.Find(ctx, bson.D{}, findOpts)
	if err != nil {
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var env envelope[T]
		if err := cursor.Decode(&env); err != nil {
			// Non-fatal: skip malformed docs and continue
			continue
		}

		select {
		case ch <- env.Payload:
		case <-ctx.Done():
			return
		}
	}
}
