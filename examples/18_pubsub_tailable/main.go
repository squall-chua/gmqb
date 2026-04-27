package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/squall-chua/gmqb"
	"github.com/tryvium-travels/memongo"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// OrderEvent represents a new order event in our system.
type OrderEvent struct {
	OrderID    string    `bson:"order_id"`
	CustomerID string    `bson:"customer_id"`
	Amount     float64   `bson:"amount"`
	CreatedAt  time.Time `bson:"created_at"`
}

func main() {
	// 1. Start a temporary MongoDB server using memongo for the demo.
	// This ensures the example works without needing a pre-installed MongoDB.
	mongoServer, err := memongo.StartWithOptions(&memongo.Options{
		MongoVersion: "8.2.5",
	})
	if err != nil {
		log.Fatalf("failed to start memongo: %v", err)
	}
	defer mongoServer.Stop()

	client, err := mongo.Connect(options.Client().ApplyURI(mongoServer.URI()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = client.Disconnect(context.Background()) }()

	db := client.Database("gmqb_examples")
	ctx := context.Background()

	// 2. Initialize the TailablePubSub bus for OrderEvent.
	// This idempotently creates a capped collection named "order_topic".
	// Capped collections are perfect for high-speed, durable event streams.
	bus, err := gmqb.NewTailablePubSub[OrderEvent](db, "order_topic", gmqb.CappedOpts{
		SizeBytes: 10 * 1024 * 1024, // 10 MB ring buffer
		MaxDocs:   5000,             // limit to 5000 events
	})
	if err != nil {
		log.Fatalf("failed to initialize bus: %v", err)
	}

	// 3. Start a Subscriber.
	// Subscriptions survive cursor death and collection drops via an internal
	// reconnect loop with exponential back-off.
	subCtx, subCancel := context.WithCancel(ctx)
	defer subCancel()

	events, stop := bus.Subscribe(subCtx)
	defer stop()

	go func() {
		fmt.Println(">>> Subscriber 1: Started listening for orders...")
		for event := range events {
			fmt.Printf(">>> Subscriber 1: [ORDER RECEIVED] ID: %s, Customer: %s, Amount: $%.2f\n",
				event.OrderID, event.CustomerID, event.Amount)
		}
		fmt.Println(">>> Subscriber 1: Stopped.")
	}()

	// 4. Publish some events.
	time.Sleep(500 * time.Millisecond) // wait for subscriber to warm up
	fmt.Println(">>> Publisher: Sending order events...")

	orders := []OrderEvent{
		{OrderID: "ORD-001", CustomerID: "CUST-A", Amount: 99.99, CreatedAt: time.Now()},
		{OrderID: "ORD-002", CustomerID: "CUST-B", Amount: 45.50, CreatedAt: time.Now()},
		{OrderID: "ORD-003", CustomerID: "CUST-C", Amount: 120.00, CreatedAt: time.Now()},
	}

	for _, order := range orders {
		if err := bus.Publish(ctx, order); err != nil {
			log.Printf("failed to publish: %v", err)
		}
		time.Sleep(200 * time.Millisecond)
	}

	// 5. Cleanup
	fmt.Println(">>> Publisher: All events sent. Waiting 1s before exit...")
	time.Sleep(1 * time.Second)
	subCancel() // Close subscriber
	time.Sleep(200 * time.Millisecond)
}
