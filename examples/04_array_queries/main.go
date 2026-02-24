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

type Result struct {
	Score int `bson:"score"`
}

type Student struct {
	Name    string   `bson:"name"`
	Results []Result `bson:"results"`
	Tags    []string `bson:"tags"`
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
	_, _ = coll.InsertMany(ctx, []Student{
		{Name: "Alice", Results: []Result{{Score: 85}, {Score: 92}}, Tags: []string{"math", "science", "arts"}}, // Matches both ElemMatch and Size
		{Name: "Bob", Results: []Result{{Score: 70}, {Score: 95}}, Tags: []string{"math", "science"}},           // Matches neither
		{Name: "Charlie", Results: []Result{{Score: 88}}, Tags: []string{"sports"}},                             // Matches ElemMatch, but not Size
	})

	// 4. Match documents where 'results' array contains at least one element that matches BOTH conditions
	filter := gmqb.ElemMatch("results", gmqb.And(
		gmqb.Gte("score", 80),
		gmqb.Lt("score", 90),
	))

	// Also filter by array size
	filter2 := gmqb.Size("tags", 3)

	fmt.Println("ElemMatch Filter JSON:")
	fmt.Println(filter.JSON())

	studentsElemMatch, _ := coll.Find(ctx, filter)
	fmt.Printf("\nFound %d student(s) matching ElemMatch:\n", len(studentsElemMatch))
	for _, s := range studentsElemMatch {
		fmt.Printf("- %+v\n", s)
	}

	fmt.Println("\nSize Filter JSON:")
	fmt.Println(filter2.JSON())

	studentsSize, _ := coll.Find(ctx, filter2)
	fmt.Printf("\nFound %d student(s) matching Size:\n", len(studentsSize))
	for _, s := range studentsSize {
		fmt.Printf("- %+v\n", s)
	}
}
