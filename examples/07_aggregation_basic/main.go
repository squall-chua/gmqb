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

type Order struct {
	Status string  `bson:"status"`
	State  string  `bson:"state"`
	City   string  `bson:"city"`
	Amount float64 `bson:"amount"`
	Score  float64 `bson:"score"`
}

type AggResult struct {
	ID struct {
		State string `bson:"state"`
		City  string `bson:"city"`
	} `bson:"_id"`
	TotalValue float64 `bson:"totalValue"`
	AvgScore   float64 `bson:"avgScore"`
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
	coll := gmqb.Wrap[Order](db.Collection("orders"))
	ctx := context.Background()

	// 3. Seed some data
	_, _ = coll.InsertMany(ctx, []Order{
		{Status: "active", State: "NY", City: "NYC", Amount: 100, Score: 4.5},
		{Status: "active", State: "NY", City: "NYC", Amount: 150, Score: 4.0},
		{Status: "inactive", State: "NY", City: "NYC", Amount: 200, Score: 5.0},
		{Status: "active", State: "CA", City: "LA", Amount: 300, Score: 4.8},
	})

	// 4. Build Pipeline
	pipeline := gmqb.NewPipeline().
		Match(gmqb.Eq("status", "active")).
		Group(gmqb.GroupSpec(
			gmqb.GroupID("state", "city"),
			gmqb.GroupAcc("totalValue", gmqb.AccSum("$amount")),
			gmqb.GroupAcc("avgScore", gmqb.AccAvg("$score")),
		)).
		Sort(gmqb.SortSpec(gmqb.SortRule("totalValue", -1))).
		Limit(10)

	fmt.Println("Basic Aggregation JSON:")
	fmt.Println(pipeline.JSON())

	// 5. Execute aggregation
	results, err := gmqb.Aggregate[AggResult](coll, ctx, pipeline)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nAggregation Results:\n")
	for _, r := range results {
		fmt.Printf("- State: %s, City: %s | Total: %.2f | Avg. Score: %.2f\n", r.ID.State, r.ID.City, r.TotalValue, r.AvgScore)
	}
}
