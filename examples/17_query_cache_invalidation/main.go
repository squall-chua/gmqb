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
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Stock struct {
	Symbol string  `bson:"symbol"`
	Price  float64 `bson:"price"`
}

func main() {
	// 1. Setup in-memory MongoDB with REPLICA SET (required for Change Streams)
	mongoServer, err := memongo.StartWithOptions(&memongo.Options{
		MongoVersion:     "8.2.5",
		ShouldUseReplica: true,
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

	// 2. Setup Cache
	goCacheClient := gocache.New(10*time.Minute, 10*time.Minute)
	cacheStore := gocachestore.NewGoCache(goCacheClient)
	cacheManager := cache.New[[]byte](cacheStore)

	// 3. Wrap Collection
	db := client.Database("finance")
	rawColl := db.Collection("stocks")
	coll := gmqb.WrapWithCache[Stock](rawColl, cacheManager, 10*time.Minute)
	ctx := context.Background()

	// 4. Start Change Stream Invalidator
	inv := gmqb.NewChangeStreamInvalidator(rawColl, cacheManager)
	watchCtx, cancelWatch := context.WithCancel(ctx)
	defer cancelWatch()

	go func() {
		fmt.Println("[Invalidator] Watching Change Stream...")
		if err := inv.Watch(watchCtx); err != nil && watchCtx.Err() == nil {
			log.Printf("[Invalidator] Watch stopped: %v", err)
		}
	}()

	// 5. Query and Cache
	_, _ = rawColl.InsertOne(ctx, Stock{Symbol: "AAPL", Price: 150.0})

	fmt.Println("\nStep 1: Initial query (loads into cache)")
	stock, _ := coll.FindOne(ctx, gmqb.Eq("symbol", "AAPL"))
	fmt.Printf("Price: %.2f\n", stock.Price)

	// 6. External write (direct to mongo-driver)
	fmt.Println("\nStep 2: External write occurring (bypassing gmqb wrapper)...")
	_, err = rawColl.UpdateOne(ctx, gmqb.Eq("symbol", "AAPL").BsonD(), gmqb.NewUpdate().Set("price", 155.5).BsonD())
	if err != nil {
		log.Fatal(err)
	}

	// 7. Wait for change stream to propagate invalidation
	fmt.Println("Waiting for Change Stream to trigger invalidation...")
	time.Sleep(2 * time.Second)

	// 8. Query again
	fmt.Println("\nStep 3: Querying again (should see updated price due to invalidation)")
	stock2, _ := coll.FindOne(ctx, gmqb.Eq("symbol", "AAPL"))
	fmt.Printf("Price: %.2f\n", stock2.Price)

	if stock2.Price == 155.5 {
		fmt.Println("\nSUCCESS: Cache was automatically invalidated!")
	} else {
		fmt.Println("\nFAILURE: Cache still returned stale data.")
	}
}
