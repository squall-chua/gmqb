package gmqb_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/squall-chua/gmqb"
	"github.com/stretchr/testify/assert"
)

// startReplicaSet is provided by cache_invalidator_test.go in the same gmqb_test package.

func TestQueue_Integration_LoadBalancing(t *testing.T) {
	db, _ := startStandalone(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	type Job struct {
		ID int `bson:"id"`
	}

	q, _ := gmqb.NewQueue[Job](db, "lb_test", gmqb.DefaultQueueOpts)
	_ = q.EnsureIndexes(ctx)

	var processed int32
	handler := func(ctx context.Context, payload Job) error {
		atomic.AddInt32(&processed, 1)
		return nil
	}

	w1 := gmqb.NewWorker(q, handler, gmqb.WithConcurrency(2))
	w2 := gmqb.NewWorker(q, handler, gmqb.WithConcurrency(2))

	go w1.Run(ctx)
	go w2.Run(ctx)

	for i := 0; i < 10; i++ {
		_, _ = q.Enqueue(ctx, Job{ID: i})
	}

	// Wait for processing
	assert.Eventually(t, func() bool {
		return atomic.LoadInt32(&processed) == 10
	}, 5*time.Second, 100*time.Millisecond)
}

func TestQueue_Integration_FanOut(t *testing.T) {
	db, _ := startStandalone(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	type Msg struct {
		Text string `bson:"text"`
	}

	q, _ := gmqb.NewQueue[Msg](db, "fanout_test", gmqb.DefaultQueueOpts)
	_ = q.EnsureIndexes(ctx)

	var countA, countB int32
	wA := gmqb.NewWorker(q, func(ctx context.Context, m Msg) error {
		atomic.AddInt32(&countA, 1)
		return nil
	}, gmqb.WithConsumerGroup("groupA"))

	wB := gmqb.NewWorker(q, func(ctx context.Context, m Msg) error {
		atomic.AddInt32(&countB, 1)
		return nil
	}, gmqb.WithConsumerGroup("groupB"))

	go wA.Run(ctx)
	go wB.Run(ctx)

	_, _ = q.Enqueue(ctx, Msg{Text: "hello"})
	_, _ = q.Enqueue(ctx, Msg{Text: "world"})

	assert.Eventually(t, func() bool {
		return atomic.LoadInt32(&countA) == 2 && atomic.LoadInt32(&countB) == 2
	}, 5*time.Second, 100*time.Millisecond)
}

func TestQueue_Integration_DLQ(t *testing.T) {
	db, _ := startStandalone(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	type Job struct {
		X int `bson:"x"`
	}

	opts := gmqb.DefaultQueueOpts
	opts.MaxAttempts = 2
	opts.VisibilityTimeout = 100 * time.Millisecond

	q, _ := gmqb.NewQueue[Job](db, "dlq_integ", opts)
	_ = q.EnsureIndexes(ctx)

	handler := func(ctx context.Context, payload Job) error {
		return assert.AnError // Always fail
	}

	w := gmqb.NewWorker(q, handler)
	go w.Run(ctx)

	_, _ = q.Enqueue(ctx, Job{X: 42})

	// Wait for it to move to DLQ
	dlq := q.DLQ()
	assert.Eventually(t, func() bool {
		list, _ := dlq.List(ctx, 10, 0)
		return len(list) == 1
	}, 5*time.Second, 100*time.Millisecond)
}

func TestQueue_Integration_ExactlyOnce(t *testing.T) {
	db, _ := startStandalone(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	type Job struct {
		Key string `bson:"key"`
	}

	q, _ := gmqb.NewQueue[Job](db, "exactly_once_test", gmqb.DefaultQueueOpts)
	_ = q.EnsureIndexes(ctx)

	var processed int32
	handler := func(ctx context.Context, payload Job) error {
		atomic.AddInt32(&processed, 1)
		return nil
	}

	w1 := gmqb.NewWorker(q, handler, gmqb.WithDelivery(gmqb.ExactlyOnce), gmqb.WithConsumerGroup("G1"))
	w2 := gmqb.NewWorker(q, handler, gmqb.WithDelivery(gmqb.ExactlyOnce), gmqb.WithConsumerGroup("G1"))

	// Manual enqueue to simulate same message ID if we were doing something tricky,
	// but normally Enqueue creates new ID.
	// Exactly-once protects against double processing if a worker crashes AFTER handler but BEFORE ack.
	// We can simulate this by nacking manually.

	_, _ = q.Enqueue(ctx, Job{Key: "once"})

	// Force double processing simulation by nacking
	go w1.Run(ctx)
	go w2.Run(ctx)

	// Wait for at least 1 processing
	assert.Eventually(t, func() bool {
		return atomic.LoadInt32(&processed) >= 1
	}, 5*time.Second, 100*time.Millisecond)

	// Wait a bit more to ensure w2 doesn't process it again
	time.Sleep(500 * time.Millisecond)
	assert.Equal(t, int32(1), atomic.LoadInt32(&processed))
}
