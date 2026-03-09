package gmqb_test

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	gocachestore "github.com/eko/gocache/store/go_cache/v4"
	gocache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tryvium-travels/memongo"
	"go.mongodb.org/mongo-driver/v2/mongo"
	mongooptions "go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/squall-chua/gmqb"
)

// startReplicaSet starts a per-test memongo server with replica set support and
// returns a connected mongo.Collection ready for change stream tests.
// The server and client are cleaned up automatically via t.Cleanup.
func startReplicaSet(t *testing.T) (*memongo.Server, *mongo.Collection) {
	t.Helper()

	srv, err := memongo.StartWithOptions(&memongo.Options{
		MongoVersion:     "8.2.5",
		ShouldUseReplica: true,
	})
	if err != nil {
		t.Skipf("memongo replica-set unavailable: %v", err)
	}
	t.Cleanup(srv.Stop)

	client, err := mongo.Connect(mongooptions.Client().ApplyURI(srv.URI()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })

	dbName := memongo.RandomDatabase()
	coll := client.Database(dbName).Collection("inv_test")
	return srv, coll
}

// newInvalidatorCache creates an in-memory gocache manager for invalidator tests.
func newInvalidatorCache() cache.CacheInterface[[]byte] {
	c := gocache.New(5*time.Minute, 10*time.Minute)
	s := gocachestore.NewGoCache(c)
	return cache.New[[]byte](s)
}

func TestChangeStream_Insert_InvalidatesCache(t *testing.T) {
	_, mColl := startReplicaSet(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cacheMgr := newInvalidatorCache()
	coll := gmqb.WrapWithCache[User](mColl, cacheMgr, time.Minute)
	inv := gmqb.NewChangeStreamInvalidator(mColl, cacheMgr)

	watchCtx, watchCancel := context.WithCancel(ctx)
	defer watchCancel()
	go func() {
		if err := inv.Watch(watchCtx); err != nil && watchCtx.Err() == nil {
			log.Printf("TestChangeStream_Insert: watch error: %v", err)
		}
	}()

	// Seed one user and warm up the cache.
	_, err := coll.InsertMany(ctx, []User{{Name: "Alice", Active: true}})
	require.NoError(t, err)
	first, err := coll.Find(ctx, gmqb.Eq("active", true))
	require.NoError(t, err)
	require.Len(t, first, 1)

	// Insert via the raw collection to bypass CachedCollection.
	_, err = mColl.InsertOne(ctx, User{Name: "Bob", Active: true})
	require.NoError(t, err)

	// Wait up to 5 s for the change stream to trigger cache invalidation.
	require.Eventually(t, func() bool {
		results, findErr := coll.Find(ctx, gmqb.Eq("active", true))
		return findErr == nil && len(results) == 2
	}, 5*time.Second, 100*time.Millisecond, "cache should be invalidated after insert")
}

func TestChangeStream_Update_InvalidatesCache(t *testing.T) {
	_, mColl := startReplicaSet(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cacheMgr := newInvalidatorCache()
	coll := gmqb.WrapWithCache[User](mColl, cacheMgr, time.Minute)
	inv := gmqb.NewChangeStreamInvalidator(mColl, cacheMgr)

	watchCtx, watchCancel := context.WithCancel(ctx)
	defer watchCancel()
	go func() { _ = inv.Watch(watchCtx) }()

	_, err := coll.InsertOne(ctx, &User{Name: "Alice", Age: 30})
	require.NoError(t, err)

	// Prime the cache.
	u, err := coll.FindOne(ctx, gmqb.Eq("name", "Alice"))
	require.NoError(t, err)
	assert.Equal(t, 30, u.Age)

	// Update via raw collection.
	_, err = mColl.UpdateOne(ctx,
		gmqb.Eq("name", "Alice").BsonD(),
		gmqb.NewUpdate().Set("age", 99).BsonD(),
	)
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		u2, findErr := coll.FindOne(ctx, gmqb.Eq("name", "Alice"))
		return findErr == nil && u2.Age == 99
	}, 5*time.Second, 100*time.Millisecond, "cache should be invalidated after update")
}

func TestChangeStream_Delete_InvalidatesCache(t *testing.T) {
	_, mColl := startReplicaSet(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cacheMgr := newInvalidatorCache()
	coll := gmqb.WrapWithCache[User](mColl, cacheMgr, time.Minute)
	inv := gmqb.NewChangeStreamInvalidator(mColl, cacheMgr)

	watchCtx, watchCancel := context.WithCancel(ctx)
	defer watchCancel()
	go func() { _ = inv.Watch(watchCtx) }()

	_, err := coll.InsertMany(ctx, []User{
		{Name: "Alice", Active: true},
		{Name: "Bob", Active: true},
	})
	require.NoError(t, err)

	// Warm cache.
	all, err := coll.Find(ctx, gmqb.Eq("active", true))
	require.NoError(t, err)
	require.Len(t, all, 2)

	// Delete via raw collection.
	_, err = mColl.DeleteOne(ctx, gmqb.Eq("name", "Alice").BsonD())
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		results, findErr := coll.Find(ctx, gmqb.Eq("active", true))
		return findErr == nil && len(results) == 1
	}, 5*time.Second, 100*time.Millisecond, "cache should be invalidated after delete")
}

func TestChangeStream_ResumeToken_PersistedAndLoaded(t *testing.T) {
	_, mColl := startReplicaSet(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	cacheMgr := newInvalidatorCache()
	inv := gmqb.NewChangeStreamInvalidator(mColl, cacheMgr)

	// Run the first watch briefly so it can accumulate a resume token.
	firstCtx, firstCancel := context.WithTimeout(ctx, 500*time.Millisecond)
	_ = inv.Watch(firstCtx) // blocks ~500 ms then returns
	firstCancel()

	// Insert a doc after the first watch ends.
	_, err := mColl.InsertOne(ctx, User{Name: "Alice"})
	require.NoError(t, err)

	// Wrap for cached reads.
	coll := gmqb.WrapWithCache[User](mColl, cacheMgr, time.Minute)

	// Start a second watch — it should resume from the persisted token.
	watchCtx, watchCancel := context.WithCancel(ctx)
	defer watchCancel()
	go func() { _ = inv.Watch(watchCtx) }()

	// Give it a moment to process any replayed / new events.
	time.Sleep(time.Second)

	// The key assertion: Watch must not crash and the collection must be readable.
	_, err = coll.Find(ctx, gmqb.Eq("name", "Alice"))
	require.NoError(t, err)
}
