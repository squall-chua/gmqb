package main

import (
	"context"
	"fmt"
	"log"

	"github.com/squall-chua/gmqb"
	"github.com/tryvium-travels/memongo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Place struct {
	Name     string `bson:"name"`
	Location bson.D `bson:"location"`
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
	coll := gmqb.Wrap[Place](db.Collection("places"))
	ctx := context.Background()

	// 2.5 Ensure Index for 2dsphere
	_, err = coll.Unwrap().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "location", Value: "2dsphere"}},
	})
	if err != nil {
		log.Fatalf("Failed to create index: %v", err)
	}

	// 3. Seed some data
	_, _ = coll.InsertMany(ctx, []Place{
		{Name: "Central Park", Location: gmqb.Point(-73.9667, 40.78)},        // Target point
		{Name: "Times Square", Location: gmqb.Point(-73.9851, 40.7589)},      // ~2.8km away
		{Name: "Statue of Liberty", Location: gmqb.Point(-74.0445, 40.6892)}, // ~12km away
	})

	// 4. Find locations near a GeoJSON point
	point := gmqb.Point(-73.9667, 40.78)
	filter := gmqb.Near("location", point, 5000, 100) // max 5000m, min 100m (excludes target point)

	fmt.Println("Geospatial Near Filter JSON:")
	fmt.Println(filter.JSON())

	// 5. Execute query
	places, err := coll.Find(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nFound %d place(s) near Central Park:\n", len(places))
	for _, p := range places {
		fmt.Printf("- %+v\n", p)
	}
}
