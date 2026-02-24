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
	Name  string   `bson:"name"`
	Price float64  `bson:"price"`
	Tags  []string `bson:"tags"`
}

type TagCount struct {
	Tag   string `bson:"_id"`
	Count int    `bson:"count"`
}

type PriceBucket struct {
	ID struct {
		Min float64 `bson:"min"`
		Max float64 `bson:"max"`
	} `bson:"_id"`
	Count int `bson:"count"`
}

type FacetResult struct {
	CategorizedByTags  []TagCount    `bson:"categorizedByTags"`
	CategorizedByPrice []PriceBucket `bson:"categorizedByPrice"`
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
		{Name: "Laptop", Price: 1200, Tags: []string{"electronics", "computers"}},
		{Name: "Mouse", Price: 25, Tags: []string{"electronics", "accessories"}},
		{Name: "Keyboard", Price: 75, Tags: []string{"electronics", "accessories"}},
		{Name: "Desk", Price: 350, Tags: []string{"furniture", "office"}},
		{Name: "Chair", Price: 150, Tags: []string{"furniture", "office"}},
	})

	// 4. Build Pipeline
	// Process documents through multiple sub-pipelines simultaneously
	pipeline := gmqb.NewPipeline().Facet(map[string]gmqb.Pipeline{
		"categorizedByTags": gmqb.NewPipeline().
			Unwind("$tags").
			Group(gmqb.GroupSpec(gmqb.GroupID("tags"), gmqb.GroupAcc("count", gmqb.AccSum(1)))).
			Sort(gmqb.SortSpec(gmqb.SortRule("count", -1))),

		"categorizedByPrice": gmqb.NewPipeline().
			Match(gmqb.Exists("price", true)).
			BucketAuto(gmqb.BucketAutoOpts{GroupBy: "$price", Buckets: 3}),
	})

	fmt.Println("Facet Aggregation JSON:")
	fmt.Println(pipeline.JSON())

	// 5. Execute aggregation
	results, err := gmqb.Aggregate[FacetResult](coll, ctx, pipeline)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nFacet Aggregation Results:\n")
	for _, r := range results {
		fmt.Printf("Tags breakdown:\n")
		for _, tc := range r.CategorizedByTags {
			fmt.Printf("  - %s: %d\n", tc.Tag, tc.Count)
		}
		fmt.Printf("Price buckets:\n")
		for _, pb := range r.CategorizedByPrice {
			fmt.Printf("  - $%.0f to $%.0f: %d items\n", pb.ID.Min, pb.ID.Max, pb.Count)
		}
	}
}
