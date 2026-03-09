package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	gocachestore "github.com/eko/gocache/store/go_cache/v4"
	gocache "github.com/patrickmn/go-cache"
	"github.com/squall-chua/gmqb"
	"github.com/tryvium-travels/memongo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Product struct {
	ID    bson.ObjectID `bson:"_id,omitempty"`
	Name  string        `bson:"name"`
	Price int           `bson:"price"`
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

	// 2. Setup Cache Manager (eko/gocache with go-cache store)
	goCacheClient := gocache.New(5*time.Minute, 10*time.Minute)
	cacheStore := gocachestore.NewGoCache(goCacheClient)
	cacheManager := cache.New[[]byte](cacheStore)

	// 3. Wrap collection with caching
	db := client.Database("shop")
	rawColl := db.Collection("products")
	// Cache results for 1 minute
	coll := gmqb.WrapWithCache[Product](rawColl, cacheManager, time.Minute)
	ctx := context.Background()

	// 4. Seed data
	_, _ = coll.InsertMany(ctx, []Product{
		{Name: "Laptop", Price: 1200},
		{Name: "Mouse", Price: 25},
		{Name: "Keyboard", Price: 75},
	})

	filter := gmqb.Gte("price", 50)

	// 5. First query (Cache Miss)
	start := time.Now()
	products1, err := coll.Find(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Query 1 (Miss) took: %v, Results: %d\n", time.Since(start), len(products1))

	// 6. Same query (Cache Hit)
	start = time.Now()
	products2, err := coll.Find(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Query 2 (Hit)  took: %v, Results: %d\n", time.Since(start), len(products2))

	// 7. Manual Invalidation
	fmt.Println("\nInvalidating cache manually...")
	_ = coll.InvalidateCache(ctx)

	// 8. Third query (Cache Miss again)
	start = time.Now()
	products3, err := coll.Find(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Query 3 (Miss) took: %v, Results: %d\n", time.Since(start), len(products3))
}
