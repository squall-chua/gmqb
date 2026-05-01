package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/squall-chua/gmqb"
	"github.com/tryvium-travels/memongo"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// EmailJob represents a background task to send an email.
type EmailJob struct {
	To      string `bson:"to"`
	Subject string `bson:"subject"`
	Body    string `bson:"body"`
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// 1. Start memongo (zero external dependencies)
	mongoServer, err := memongo.StartWithOptions(&memongo.Options{
		MongoVersion: "8.2.5",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer mongoServer.Stop()

	// 2. Connect to MongoDB
	client, err := mongo.Connect(options.Client().ApplyURI(mongoServer.URI()))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	db := client.Database(memongo.RandomDatabase())

	// 3. Initialize Queue
	// We enable DLQ and set max retries to 3.
	q, err := gmqb.NewQueue[EmailJob](db, "emails", gmqb.QueueOpts{
		MaxAttempts:       3,
		VisibilityTimeout: 30 * time.Second,
		DLQEnabled:        true,
	})
	if err != nil {
		log.Fatal(err)
	}

	// 4. Ensure Indexes (important for performance)
	if err := q.EnsureIndexes(ctx); err != nil {
		log.Fatal(err)
	}

	// 5. Start a Worker (Consumer Group A)
	worker := gmqb.NewWorker(q, func(ctx context.Context, job EmailJob) error {
		fmt.Printf("Worker A: Sending email to %s: %s\n", job.To, job.Subject)
		// Simulate work
		time.Sleep(100 * time.Millisecond)
		return nil
	}, gmqb.WithConcurrency(2), gmqb.WithConsumerGroup("processor-a"))

	go func() {
		if err := worker.Run(ctx); err != nil {
			log.Printf("Worker error: %v", err)
		}
	}()

	// 6. Producer: Enqueue some jobs
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for i := 1; ; i++ {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				job := EmailJob{
					To:      fmt.Sprintf("user%d@example.com", i),
					Subject: "Welcome!",
					Body:    "Thanks for joining gmqb-mq.",
				}
				id, err := q.Enqueue(ctx, job)
				if err != nil {
					log.Printf("Enqueue error: %v", err)
				} else {
					fmt.Printf("Enqueued job %s\n", id.Hex())
				}
			}
		}
	}()

	fmt.Println("Message Queue example running. Press Ctrl+C to stop.")
	<-ctx.Done()
	fmt.Println("Shutting down...")
}
