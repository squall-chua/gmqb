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

type SalesRecord struct {
	State     string `bson:"state"`
	OrderDate string `bson:"orderDate"` // Simplified as string for example
	Quantity  int    `bson:"quantity"`
}

type WindowResult struct {
	State              string `bson:"state"`
	OrderDate          string `bson:"orderDate"`
	Quantity           int    `bson:"quantity"`
	CumulativeQuantity int    `bson:"cumulativeQuantity"`
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
	coll := gmqb.Wrap[SalesRecord](db.Collection("sales"))
	ctx := context.Background()

	// 3. Seed some data
	_, _ = coll.InsertMany(ctx, []SalesRecord{
		{State: "CA", OrderDate: "2023-01-01", Quantity: 10},
		{State: "CA", OrderDate: "2023-01-05", Quantity: 20},
		{State: "CA", OrderDate: "2023-01-10", Quantity: 5},
		{State: "NY", OrderDate: "2023-01-02", Quantity: 15},
		{State: "NY", OrderDate: "2023-01-08", Quantity: 25},
	})

	// 4. Build Pipeline
	pipeline := gmqb.NewPipeline().SetWindowFields(gmqb.SetWindowFieldsSpec(
		"$state", // partitionBy
		gmqb.SortSpec(gmqb.SortRule("orderDate", 1)),
		gmqb.WindowOutput(
			"cumulativeQuantity",
			gmqb.AccSum("$quantity"),
			gmqb.Window("documents", "unbounded", "current"),
		),
	))

	fmt.Println("SetWindowFields Aggregation JSON:")
	fmt.Println(pipeline.JSON())

	// 5. Execute aggregation
	results, err := gmqb.Aggregate[WindowResult](coll, ctx, pipeline)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nAggregation Results:\n")
	for _, r := range results {
		fmt.Printf("- State: %s, Date: %s, Qty: %d, Running Total: %d\n", r.State, r.OrderDate, r.Quantity, r.CumulativeQuantity)
	}
}
