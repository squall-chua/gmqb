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

type Inventory struct {
	SKU string `bson:"sku"`
	Qty int    `bson:"qty"`
}

type Order struct {
	Item     string `bson:"item"`
	Price    int    `bson:"price"`
	Quantity int    `bson:"quantity"`
}

type AggResult struct {
	Item          string      `bson:"item"`
	Price         int         `bson:"price"`
	Quantity      int         `bson:"quantity"`
	InventoryDocs []Inventory `bson:"inventory_docs,omitempty"` // before unwind
	InventoryDoc  Inventory   `bson:"inventory_doc,omitempty"`  // after unwind (optional if unwound)
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

	// 2. Wrap collections
	db := client.Database("testdb")
	collOrders := gmqb.Wrap[Order](db.Collection("orders"))
	collInventory := gmqb.Wrap[Inventory](db.Collection("inventory"))
	ctx := context.Background()

	// 3. Seed some data
	_, _ = collOrders.InsertMany(ctx, []Order{
		{Item: "almonds", Price: 12, Quantity: 2},
		{Item: "pecans", Price: 20, Quantity: 1},
	})
	_, _ = collInventory.InsertMany(ctx, []Inventory{
		{SKU: "almonds", Qty: 120},
		{SKU: "pecans", Qty: 80},
	})

	// 4. Build Pipeline
	pipeline := gmqb.NewPipeline().
		Lookup(gmqb.LookupOpts{From: "inventory", LocalField: "item", ForeignField: "sku", As: "inventory_docs"}).
		Unwind("$inventory_docs") // NOTE: Added $ for Unwind syntax

	fmt.Println("Lookup and Unwind Aggregation JSON:")
	fmt.Println(pipeline.JSON())

	// 5. Execute aggregation
	results, err := gmqb.Aggregate[AggResult](collOrders, ctx, pipeline)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nAggregation Results:\n")
	for _, r := range results {
		// Output the unwound inventory doc which is populated into the same name due to unwind path unless renamed
		fmt.Printf("- Item: %s, Price: %d, Qty: %d, Inventory: %+v\n", r.Item, r.Price, r.Quantity, r.InventoryDocs)
	}
}
