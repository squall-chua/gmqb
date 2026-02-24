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

type Customer struct {
	Name       string  `bson:"name"`
	TotalSpent float64 `bson:"totalSpent"`
	Price      float64 `bson:"price"`
}

type DiscountResult struct {
	Name         string  `bson:"name"`
	TotalSpent   float64 `bson:"totalSpent"`
	Price        float64 `bson:"price"`
	DiscountRate float64 `bson:"discountRate"`
	FinalPrice   float64 `bson:"finalPrice"`
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
	coll := gmqb.Wrap[Customer](db.Collection("customers"))
	ctx := context.Background()

	// 3. Seed some data
	_, _ = coll.InsertMany(ctx, []Customer{
		{Name: "Alice", TotalSpent: 1200, Price: 100},
		{Name: "Bob", TotalSpent: 500, Price: 100},
	})

	// 4. Calculate a discount conditionally using aggregation expressions
	condExpr := gmqb.ExprCond(
		gmqb.ExprGte("$totalSpent", 1000),
		0.20, // 20% off
		0.05, // 5% off
	)

	pipeline := gmqb.NewPipeline().
		AddFields(gmqb.AddFieldsSpec(
			gmqb.AddField("discountRate", condExpr),
			gmqb.AddField("finalPrice", gmqb.ExprMultiply("$price", gmqb.ExprSubtract(1, "$discountRate"))),
		))

	fmt.Println("Expressions Aggregation JSON:")
	fmt.Println(pipeline.JSON())

	// 5. Execute aggregation
	results, err := gmqb.Aggregate[DiscountResult](coll, ctx, pipeline)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nAggregation Results:\n")
	for _, r := range results {
		fmt.Printf("- %s: Spent=%.2f, BasePrice=%.2f, Rate=%.2f, Final=%.2f\n", r.Name, r.TotalSpent, r.Price, r.DiscountRate, r.FinalPrice)
	}
}
