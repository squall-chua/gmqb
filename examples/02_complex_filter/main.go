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

type Product struct {
	Sku      string   `bson:"sku"`
	Tags     []string `bson:"tags"`
	Metadata struct {
		Weight int `bson:"weight"`
	} `bson:"metadata"`
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
	coll := gmqb.Wrap[Product](db.Collection("products"))
	ctx := context.Background()

	// 3. Seed some data
	_, _ = coll.InsertMany(ctx, []Product{
		{Sku: "ABC-123", Tags: []string{"electronics"}, Metadata: struct {
			Weight int `bson:"weight"`
		}{Weight: 10}}, // Matches regex
		{Sku: "XYZ-999", Tags: []string{"sale", "clearance"}, Metadata: struct {
			Weight int `bson:"weight"`
		}{Weight: 60}}, // Matches AND condition
		{Sku: "XYZ-888", Tags: []string{"sale"}, Metadata: struct {
			Weight int `bson:"weight"`
		}{Weight: 40}}, // No match
	})

	// 4. Build Filter
	f := gmqb.Field[Product]
	filter := gmqb.Or(
		gmqb.Regex(f("Sku"), "^ABC", "i"),
		gmqb.And(
			gmqb.In(f("Tags"), "sale", "clearance"),
			gmqb.Exists(f("Metadata"), true),
			gmqb.Gt(f("Metadata.Weight"), 50),
		),
	)

	fmt.Println("Complex Filter JSON:")
	fmt.Println(filter.JSON())

	// 5. Execute query
	products, err := coll.Find(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nFound %d product(s) matching complex filter:\n", len(products))
	for _, p := range products {
		fmt.Printf("- %+v\n", p)
	}
}
