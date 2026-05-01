package gmqb

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// DeliveryMode defines the delivery guarantee for the worker.
type DeliveryMode int

const (
	// AtMostOnce guarantees that a message is delivered at most once.
	// It is "fire and forget" — if the worker crashes during processing,
	// the message is lost.
	AtMostOnce DeliveryMode = iota
	// AtLeastOnce guarantees that a message is delivered at least once.
	// If a worker crashes or fails, the message will be retried after the
	// visibility timeout.
	AtLeastOnce
	// ExactlyOnce guarantees that a message is processed exactly once
	// per consumer group by using a deduplication collection.
	ExactlyOnce
)

// WorkerMode defines how workers distribute messages.
type WorkerMode int

const (
	// LoadBalance mode distributes messages among workers in a competing
	// consumer pattern. Each message is processed by only one worker.
	LoadBalance WorkerMode = iota
	// FanOut mode delivers every message to every consumer group.
	FanOut
)

// QueueOpts configures the queue behavior.
type QueueOpts struct {
	// MaxAttempts is the maximum number of times a message will be retried
	// before being moved to the dead-letter queue. Default is 3.
	MaxAttempts int
	// VisibilityTimeout is the duration a message remains hidden from other
	// workers after being claimed. Default is 30s.
	VisibilityTimeout time.Duration
	// DLQEnabled determines if dead-letter queueing is enabled. Default is true.
	DLQEnabled bool
	// DedupTTL is the retention period for the deduplication collection. Default is 7 days.
	DedupTTL time.Duration
	// DLQTTL is the retention period for the dead-letter queue. Default is 30 days.
	DLQTTL time.Duration
	// RetentionTTL is the retention period for messages in the primary queue.
	// This is especially important for Fan-Out mode. Default is 7 days.
	RetentionTTL time.Duration
}

// DefaultQueueOpts provides sensible defaults for a Queue.
var DefaultQueueOpts = QueueOpts{
	MaxAttempts:       3,
	VisibilityTimeout: 30 * time.Second,
	DLQEnabled:        true,
	DedupTTL:          7 * 24 * time.Hour,
	DLQTTL:            30 * 24 * time.Hour,
	RetentionTTL:      7 * 24 * time.Hour,
}

// WorkerConfig holds the configuration for a Worker.
type WorkerConfig struct {
	Delivery          DeliveryMode
	Concurrency       int
	ConsumerGroup     string
	VisibilityTimeout time.Duration
	ChangeStreamWake  bool
	DedupTTL          time.Duration
	RetentionTTL      time.Duration
}

// WorkerOpt is a functional option for configuring a Worker.
type WorkerOpt func(*WorkerConfig)

// WithDelivery sets the delivery guarantee for the worker.
func WithDelivery(mode DeliveryMode) WorkerOpt {
	return func(c *WorkerConfig) {
		c.Delivery = mode
	}
}

// WithConcurrency sets the number of concurrent goroutines for the worker.
func WithConcurrency(n int) WorkerOpt {
	return func(c *WorkerConfig) {
		if n > 0 {
			c.Concurrency = n
		}
	}
}

// WithConsumerGroup sets the consumer group name, enabling FanOut mode.
func WithConsumerGroup(name string) WorkerOpt {
	return func(c *WorkerConfig) {
		c.ConsumerGroup = name
	}
}

// WithWorkerVisibilityTimeout overrides the queue's visibility timeout for this worker.
func WithWorkerVisibilityTimeout(d time.Duration) WorkerOpt {
	return func(c *WorkerConfig) {
		c.VisibilityTimeout = d
	}
}

// WithChangeStreamWakeUp enables push-based wake-up via MongoDB change streams.
func WithChangeStreamWakeUp() WorkerOpt {
	return func(c *WorkerConfig) {
		c.ChangeStreamWake = true
	}
}

// WithDedupTTL sets the retention period for the deduplication collection.
func WithDedupTTL(d time.Duration) WorkerOpt {
	return func(c *WorkerConfig) {
		c.DedupTTL = d
	}
}

// WithFanOutRetentionTTL sets the retention period for messages in fan-out mode.
func WithFanOutRetentionTTL(d time.Duration) WorkerOpt {
	return func(c *WorkerConfig) {
		c.RetentionTTL = d
	}
}

// EnqueueConfig holds configuration for a single Enqueue operation.
type EnqueueConfig struct {
	ID         bson.ObjectID
	Idempotent bool
}

// EnqueueOpt is a functional option for Enqueue.
type EnqueueOpt func(*EnqueueConfig)

// WithID sets a custom ID for the message, enabling producer-side idempotency.
func WithID(id bson.ObjectID) EnqueueOpt {
	return func(c *EnqueueConfig) {
		c.ID = id
	}
}

// WithIdempotent ensures that if a message with the same ID already exists,
// Enqueue returns success instead of a duplicate key error.
func WithIdempotent() EnqueueOpt {
	return func(c *EnqueueConfig) {
		c.Idempotent = true
	}
}
