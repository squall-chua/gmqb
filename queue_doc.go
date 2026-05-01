/*
Package gmqb provides a durable, typed message queue feature backed by MongoDB.

Overview

The message queue feature supports competing consumers (load balancing), consumer groups (fan-out),
and multiple delivery guarantees.

Basic Usage

	db := client.Database("mydb")
	q, _ := gmqb.NewQueue[MyJob](db, "my_queue", gmqb.DefaultQueueOpts)
	_ = q.EnsureIndexes(ctx)

	// Producer
	q.Enqueue(ctx, MyJob{ID: 1})

	// Consumer
	worker := gmqb.NewWorker(q, func(ctx context.Context, job MyJob) error {
		fmt.Println("Processing:", job.ID)
		return nil
	})
	worker.Run(ctx)

Delivery Guarantees

- AtMostOnce: Fire and forget. Message is marked as done immediately upon being claimed.
- AtLeastOnce: Default mode. Message is retried if handler returns an error or worker crashes.
- ExactlyOnce: Uses a deduplication collection to ensure a message is processed only once per group.

Worker Models

- Load Balancing: If no ConsumerGroup is specified, workers compete for messages in the primary queue.
- Fan-Out: If a ConsumerGroup is specified, every message is delivered to every unique group.

Performance

The queue uses adaptive polling with backoff (50ms to 500ms). When a message is enqueued within the
same process, a "wake" signal is sent to local workers for near-instant (0-5ms) response.
For cross-process near-real-time delivery, workers can opt-in to MongoDB Change Streams wake-up.

Dead Letter Queue (DLQ)

Messages that exceed MaxAttempts are automatically moved to a "{name}_dlq" collection.
You can access and requeue them using q.DLQ().
*/
package gmqb
