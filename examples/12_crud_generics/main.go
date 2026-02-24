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
	Name   string `bson:"name"`
	Status string `bson:"status"`
}

type NameOnly struct {
	Name string `bson:"name"`
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

	// 2. Wrap a literal mongo.Collection into a gmqb.Collection
	db := client.Database("testdb")
	coll := gmqb.Wrap[User](db.Collection("users"))
	ctx := context.Background()

	// Seed some data
	_, _ = coll.InsertMany(ctx, []User{
		{Name: "Alice", Status: "active"},
		{Name: "Bob", Status: "inactive"},
	})

	// 3. Use gmqb's Find with strong typing
	filter := gmqb.Eq("status", "active")
	// users is typed as []User
	users, err := coll.Find(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Active Users:\n")
	for _, u := range users {
		fmt.Printf("- %+v\n", u)
	}

	// 4. For custom projections:
	pipeline := gmqb.NewPipeline().Match(gmqb.Eq("status", "active"))
	names, err := gmqb.Aggregate[NameOnly](coll, ctx, pipeline)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nProjected User Names:\n")
	for _, n := range names {
		fmt.Printf("- %s\n", n.Name)
	}
}
