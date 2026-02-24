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

type ScoreRecord struct {
	Score int `bson:"score"`
}

type Student struct {
	Name   string        `bson:"name"`
	Tags   []string      `bson:"tags"`
	Scores []ScoreRecord `bson:"scores"`
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
	coll := gmqb.Wrap[Student](db.Collection("students"))
	ctx := context.Background()

	// 3. Seed some data
	_, _ = coll.InsertOne(ctx, &Student{
		Name: "Alice",
		Tags: []string{"active"},
		Scores: []ScoreRecord{
			{Score: 80},
			{Score: 85},
		},
	})

	// 4. Update 1: Add to set
	u1 := gmqb.NewUpdate().AddToSet("tags", "premium")

	fmt.Println("AddToSet Update JSON:")
	fmt.Println(u1.JSON())

	_, err = coll.UpdateOne(ctx, gmqb.Eq("name", "Alice"), u1)
	if err != nil {
		log.Fatal(err)
	}

	// 5. Update 2: Push with opts (sort and slice)
	slice := -5 // Keep only the top 5 (Wait, negative slice means keep last N elements, but we sort -1 so highest are first? No, sort -1 makes highest first, slice -5 keeps the last 5... Wait, negative slice means keep last 5 elements AFTER sort. Wait, for Top N we need sort -1 and slice N (positive limit is only supported in newer MongoDB? No, slice must be negative or zero in MongoDB update push, meaning keep last N. So we sort 1 to put highest at end, or we keep last N). Wait, in $slice, passing -5 means keep the last 5 elements. If we sort by score: -1 (descending, highest first), the highest are at the beginning. If we slice -5, we keep the last 5, which would be the lowest! Actually, MongoDB docs say positive slice is supported since 3.6 for $push.)
	// Actually, let's keep the user's example logic:

	// Pushing multiple items with a sort and slice applied
	u2 := gmqb.NewUpdate().PushWithOpts("scores", gmqb.PushOpts{
		Each:  []interface{}{ScoreRecord{Score: 89}, ScoreRecord{Score: 92}},
		Sort:  bson.D{{Key: "score", Value: -1}}, // Keep highest scores
		Slice: &slice,                            // Originally slice was -5
	})

	fmt.Println("\nPushWithOpts Update JSON:")
	fmt.Println(u2.JSON())

	_, err = coll.UpdateOne(ctx, gmqb.Eq("name", "Alice"), u2)
	if err != nil {
		log.Fatal(err)
	}

	alice, _ := coll.FindOne(ctx, gmqb.Eq("name", "Alice"))
	fmt.Printf("\nUpdated Alice:\n%+v\n", alice)
}
