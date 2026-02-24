package main

import (
	"context"
	"fmt"
	"log"

	"github.com/squall-chua/gmqb"
	"github.com/tryvium-travels/memongo"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Session struct {
	Status string `bson:"status"`
}

func main() {
	// 1. Setup in-memory MongoDB
	mongoServer, err := memongo.StartWithOptions(&memongo.Options{
		MongoVersion: "8.2.5",
	})
	if err != nil {
		log.Fatalf("Failed to start memongo: %v", err)
	}
	defer mongoServer.Stop()

	client, err := mongo.Connect(options.Client().ApplyURI(mongoServer.URI()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect(context.Background())

	// 2. Wrap collection
	db := client.Database("testdb")
	coll := gmqb.Wrap[Session](db.Collection("sessions"))
	ctx := context.Background()

	// 3. Seed some data
	_, _ = coll.InsertMany(ctx, []Session{
		{Status: "active"},
		{Status: "inactive"},
	})

	filter := gmqb.Eq("status", "active")

	// Serializes identical to what the MongoDB shell expects
	jsonStr := filter.JSON()
	compactJson := filter.CompactJSON()

	fmt.Println("Pretty JSON:")
	fmt.Println(jsonStr)

	fmt.Println("\nCompact JSON:")
	fmt.Println(compactJson)

	// Verify query executes successfully
	activeSessions, _ := coll.Find(ctx, filter)
	fmt.Printf("\nFound %d active session(s)\n", len(activeSessions))
}
