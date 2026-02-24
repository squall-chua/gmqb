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

type User struct {
	Name string `bson:"name"`
	Age  int    `bson:"age"`
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
	coll := gmqb.Wrap[User](db.Collection("users"))
	ctx := context.Background()

	// 3. Seed some data
	_, _ = coll.InsertMany(ctx, []User{
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 25},
		{Name: "Alice", Age: 17}, // Too young
	})

	// 4. Build Filter
	f := gmqb.Field[User]
	filter := gmqb.And(
		gmqb.Eq(f("Name"), "Alice"),
		gmqb.Gte(f("Age"), 18),
	)

	fmt.Println("Basic Find Filter JSON:")
	fmt.Println(filter.JSON())

	// 5. Execute query
	users, err := coll.Find(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nFound %d user(s) matching Alice >= 18:\n", len(users))
	for _, u := range users {
		fmt.Printf("- %+v\n", u)
	}
}
