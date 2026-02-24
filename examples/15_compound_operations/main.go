package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/squall-chua/gmqb"
	"github.com/tryvium-travels/memongo"
)

// User represents a document in the users collection.
type User struct {
	Name   string `bson:"name"`
	Age    int    `bson:"age"`
	Status string `bson:"status"`
}

func main() {
	// 1. Setup in-memory MongoDB for the example
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
	coll := gmqb.Wrap[User](db.Collection("users"))
	ctx := context.Background()

	// 3. Seed some data
	_, _ = coll.InsertMany(ctx, []User{
		{Name: "Alice", Age: 30, Status: "active"},
		{Name: "Bob", Age: 25, Status: "inactive"},
		{Name: "Charlie", Age: 35, Status: "active"},
	})
	fmt.Println("Seeded users: Alice, Bob, Charlie")

	// 4. FindOneAndUpdate: Update Bob's status to active and return the updated document.
	// By default, it returns the document BEFORE the update. We use WithReturnDocument(options.After)
	// to get the document AFTER the update.
	fmt.Println("\n--- FindOneAndUpdate ---")
	updatedBob, err := coll.FindOneAndUpdate(ctx,
		gmqb.Eq("name", "Bob"),
		gmqb.NewUpdate().Set("status", "active"),
		gmqb.WithReturnDocument(options.After),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Updated Bob: %+v\n", updatedBob)

	// 5. FindOneAndReplace: Replace Charlie entirely with a new document.
	fmt.Println("\n--- FindOneAndReplace ---")
	replacedCharlie, err := coll.FindOneAndReplace(ctx,
		gmqb.Eq("name", "Charlie"),
		&User{Name: "Charlie", Age: 36, Status: "super-active"},
		gmqb.WithReturnDocumentReplace(options.After),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Replaced Charlie: %+v\n", replacedCharlie)

	// 6. FindOneAndDelete: Delete Alice and return her document.
	fmt.Println("\n--- FindOneAndDelete ---")
	deletedAlice, err := coll.FindOneAndDelete(ctx, gmqb.Eq("name", "Alice"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Deleted Alice: %+v\n", deletedAlice)

	// 7. Verify final state
	fmt.Println("\n--- Final Collection State ---")
	remainingUsers, _ := coll.Find(ctx, gmqb.Raw(nil))
	for _, u := range remainingUsers {
		fmt.Printf("- %+v\n", u)
	}
}
