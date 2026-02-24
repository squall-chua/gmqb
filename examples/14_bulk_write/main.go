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

type User struct {
	Name    string   `bson:"name"`
	Age     int      `bson:"age"`
	Country string   `bson:"country"`
	Active  bool     `bson:"active"`
	Tags    []string `bson:"tags,omitempty"`
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

	clientOpt := options.Client().ApplyURI(mongoServer.URI())
	client, err := mongo.Connect(clientOpt)
	if err != nil {
		log.Fatalf("mongo connect error: %v", err)
	}
	defer client.Disconnect(context.Background())

	// 2. Wrap the collection with our generic User type
	coll := gmqb.Wrap[User](client.Database("gmqb_examples").Collection("users"))
	ctx := context.Background()

	// Clear collection for repeatable runs
	_, _ = coll.DeleteMany(ctx, gmqb.Raw(nil))

	// 3. Define the Bulk Write operations
	// Let's perform a mix of insert, update, replace and delete models
	fmt.Println("--- Bulk Write Operations ---")

	models := []gmqb.WriteModel[User]{
		// Insert a new user
		gmqb.NewInsertOneModel[User]().SetDocument(&User{
			Name:    "Alice",
			Age:     30,
			Country: "US",
			Active:  true,
		}),

		// Insert another user
		gmqb.NewInsertOneModel[User]().SetDocument(&User{
			Name:    "Bob",
			Age:     25,
			Country: "UK",
			Active:  true,
		}),

		// Insert a user to be deleted shortly
		gmqb.NewInsertOneModel[User]().SetDocument(&User{
			Name:    "Charlie",
			Age:     35,
			Country: "CA",
			Active:  false,
		}),

		// Update "Alice" and increment her age by 1
		gmqb.NewUpdateOneModel[User]().
			SetFilter(gmqb.Eq("name", "Alice")).
			SetUpdate(gmqb.NewUpdate().Inc("age", 1)),

		// Delete Charlie
		gmqb.NewDeleteOneModel[User]().
			SetFilter(gmqb.Eq("name", "Charlie")),

		// Replace Bob with a completely new document
		gmqb.NewReplaceOneModel[User]().
			SetFilter(gmqb.Eq("name", "Bob")).
			SetReplacement(&User{
				Name:    "Bob Replaced",
				Age:     99,
				Country: "NZ",
				Active:  true,
			}),
	}

	// 4. Execute Bulk Write
	// You can pass gmqb.WithOrdered(false) if order is not important
	bulkRes, err := coll.BulkWrite(ctx, models, gmqb.WithOrdered(true))
	if err != nil {
		log.Fatalf("bulk write error: %v", err)
	}

	fmt.Printf("Inserted Count: %d\n", bulkRes.InsertedCount)
	fmt.Printf("Modified Count: %d\n", bulkRes.ModifiedCount)
	fmt.Printf("Deleted  Count: %d\n", bulkRes.DeletedCount)

	// 5. Verify the remaining data
	fmt.Println("\n--- Find Results ---")
	users, err := coll.Find(ctx, gmqb.Raw(nil), gmqb.WithSort(gmqb.Asc("name")))
	if err != nil {
		log.Fatalf("find error: %v", err)
	}

	for _, u := range users {
		fmt.Printf("User: %s | Age: %d | Country: %s\n", u.Name, u.Age, u.Country)
	}
}
