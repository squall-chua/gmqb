package gmqb_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/codec"
	gocachestore "github.com/eko/gocache/store/go_cache/v4"
	gocache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/squall-chua/gmqb"
)

// newInMemCache creates a fresh in-memory gocache manager for each test.
func newInMemCache() cache.CacheInterface[[]byte] {
	client := gocache.New(5*time.Minute, 10*time.Minute)
	store := gocachestore.NewGoCache(client)
	return cache.New[[]byte](store)
}

// cachedUsers returns a CachedCollection[User] backed by an in-memory cache,
// using the shared testDB populated in TestMain (integration_test.go).
func cachedUsers(t *testing.T, ttl time.Duration) *gmqb.CachedCollection[User] {
	t.Helper()
	coll := testDB.Collection(t.Name())
	_ = coll.Drop(context.Background())
	return gmqb.WrapWithCache[User](coll, newInMemCache(), ttl)
}

// --- Find ---

func TestCache_Find_CacheMiss_ThenHit(t *testing.T) {
	coll := cachedUsers(t, time.Minute)
	ctx := context.Background()

	// Seed some docs
	_, err := coll.InsertMany(ctx, []User{
		{Name: "Alice", Age: 30, Active: true},
		{Name: "Bob", Age: 25, Active: false},
	})
	require.NoError(t, err)

	// First call — cache miss, goes to MongoDB
	results1, err := coll.Find(ctx, gmqb.Eq("active", true))
	require.NoError(t, err)
	require.Len(t, results1, 1)
	assert.Equal(t, "Alice", results1[0].Name)

	// Insert another active user — without invalidation the cache should still return the old value
	_, err = coll.InsertOne(ctx, &User{Name: "Carol", Age: 22, Active: true})
	require.NoError(t, err)

	// Second call with same filter — should come from cache (1 result, not 2)
	results2, err := coll.Find(ctx, gmqb.Eq("active", true))
	require.NoError(t, err)
	assert.Len(t, results2, 1, "cache hit should return the stale result")
}

func TestCache_Find_DifferentFilters_DifferentKeys(t *testing.T) {
	coll := cachedUsers(t, time.Minute)
	ctx := context.Background()

	_, err := coll.InsertMany(ctx, []User{
		{Name: "Alice", Age: 30, Active: true},
		{Name: "Bob", Age: 25, Active: false},
	})
	require.NoError(t, err)

	active, err := coll.Find(ctx, gmqb.Eq("active", true))
	require.NoError(t, err)
	require.Len(t, active, 1)

	inactive, err := coll.Find(ctx, gmqb.Eq("active", false))
	require.NoError(t, err)
	require.Len(t, inactive, 1)

	// Each filter must produce independent results
	assert.Equal(t, "Alice", active[0].Name)
	assert.Equal(t, "Bob", inactive[0].Name)
}

func TestCache_Find_DifferentOptions_DifferentKeys(t *testing.T) {
	coll := cachedUsers(t, time.Minute)
	ctx := context.Background()

	_, err := coll.InsertMany(ctx, []User{
		{Name: "Alice", Age: 30, Active: true},
		{Name: "Bob", Age: 22, Active: true},
		{Name: "Carol", Age: 28, Active: true},
	})
	require.NoError(t, err)

	// Limit 1 sorted ascending
	one, err := coll.Find(ctx, gmqb.Eq("active", true),
		gmqb.WithSort(gmqb.Asc("age")), gmqb.WithLimit(1))
	require.NoError(t, err)
	require.Len(t, one, 1)
	assert.Equal(t, "Bob", one[0].Name) // youngest

	// No limit — different cache key, different result count
	all, err := coll.Find(ctx, gmqb.Eq("active", true),
		gmqb.WithSort(gmqb.Asc("age")))
	require.NoError(t, err)
	assert.Len(t, all, 3)
}

// --- FindOne ---

func TestCache_FindOne_CacheMiss_ThenHit(t *testing.T) {
	coll := cachedUsers(t, time.Minute)
	ctx := context.Background()

	_, err := coll.InsertOne(ctx, &User{Name: "Alice", Age: 30})
	require.NoError(t, err)

	u1, err := coll.FindOne(ctx, gmqb.Eq("name", "Alice"))
	require.NoError(t, err)
	assert.Equal(t, 30, u1.Age)

	// Update the doc — cache should still return old age
	_, err = coll.UpdateOne(ctx, gmqb.Eq("name", "Alice"), gmqb.NewUpdate().Set("age", 99))
	require.NoError(t, err)

	u2, err := coll.FindOne(ctx, gmqb.Eq("name", "Alice"))
	require.NoError(t, err)
	assert.Equal(t, 30, u2.Age, "cache hit expected — stale age")
}

// --- CountDocuments ---

func TestCache_CountDocuments_CacheMiss_ThenHit(t *testing.T) {
	coll := cachedUsers(t, time.Minute)
	ctx := context.Background()

	_, err := coll.InsertMany(ctx, []User{
		{Name: "A", Active: true},
		{Name: "B", Active: true},
		{Name: "C", Active: false},
	})
	require.NoError(t, err)

	count1, err := coll.CountDocuments(ctx, gmqb.Eq("active", true))
	require.NoError(t, err)
	assert.Equal(t, int64(2), count1)

	// Insert another active user — cache should still return 2
	_, err = coll.InsertOne(ctx, &User{Name: "D", Active: true})
	require.NoError(t, err)

	count2, err := coll.CountDocuments(ctx, gmqb.Eq("active", true))
	require.NoError(t, err)
	assert.Equal(t, int64(2), count2, "cache hit expected")
}

// --- CachedAggregate ---

func TestCache_CachedAggregate_CacheMiss_ThenHit(t *testing.T) {
	coll := cachedUsers(t, time.Minute)
	ctx := context.Background()

	_, err := coll.InsertMany(ctx, []User{
		{Name: "Alice", Country: "US"},
		{Name: "Bob", Country: "US"},
		{Name: "Carol", Country: "UK"},
	})
	require.NoError(t, err)

	type CountryStats struct {
		Country string `bson:"_id"`
		Count   int    `bson:"count"`
	}

	pipeline := gmqb.NewPipeline().
		Group(gmqb.GroupSpec("$country", gmqb.GroupAcc("count", gmqb.AccSum(1)))).
		Sort(gmqb.Desc("count"))

	stats1, err := gmqb.CachedAggregate[CountryStats](coll, ctx, pipeline)
	require.NoError(t, err)
	require.Len(t, stats1, 2)
	assert.Equal(t, 2, stats1[0].Count)
	assert.Equal(t, "US", stats1[0].Country)

	// Insert more — cache must return old result
	_, err = coll.InsertOne(ctx, &User{Name: "Dave", Country: "DE"})
	require.NoError(t, err)

	stats2, err := gmqb.CachedAggregate[CountryStats](coll, ctx, pipeline)
	require.NoError(t, err)
	assert.Len(t, stats2, 2, "cache hit: stale result expected")
}

// --- InvalidateCache ---

func TestCache_InvalidateCache_ClearsAll(t *testing.T) {
	coll := cachedUsers(t, time.Minute)
	ctx := context.Background()

	_, err := coll.InsertMany(ctx, []User{
		{Name: "Alice", Active: true},
		{Name: "Bob", Active: false},
	})
	require.NoError(t, err)

	// Warm up two cache entries
	_, err = coll.Find(ctx, gmqb.Eq("active", true))
	require.NoError(t, err)
	count1, err := coll.CountDocuments(ctx, gmqb.Eq("active", false))
	require.NoError(t, err)
	assert.Equal(t, int64(1), count1)

	// Add more docs then invalidate
	_, err = coll.InsertMany(ctx, []User{
		{Name: "Carol", Active: true},
		{Name: "Dave", Active: false},
	})
	require.NoError(t, err)

	err = coll.InvalidateCache(ctx)
	require.NoError(t, err)

	// Re-fetch should now see fresh data
	active, err := coll.Find(ctx, gmqb.Eq("active", true))
	require.NoError(t, err)
	assert.Len(t, active, 2, "cache invalidated: fresh result expected")

	count2, err := coll.CountDocuments(ctx, gmqb.Eq("active", false))
	require.NoError(t, err)
	assert.Equal(t, int64(2), count2, "count should reflect fresh data")
}

// --- TTL expiry ---

func TestCache_Find_CacheExpires_AfterTTL(t *testing.T) {
	coll := cachedUsers(t, 200*time.Millisecond)
	ctx := context.Background()

	_, err := coll.InsertOne(ctx, &User{Name: "Alice", Active: true})
	require.NoError(t, err)

	// Prime cache
	r1, err := coll.Find(ctx, gmqb.Eq("active", true))
	require.NoError(t, err)
	assert.Len(t, r1, 1)

	// Insert another while still in TTL window — should get cached result
	_, err = coll.InsertOne(ctx, &User{Name: "Bob", Active: true})
	require.NoError(t, err)
	r2, err := coll.Find(ctx, gmqb.Eq("active", true))
	require.NoError(t, err)
	assert.Len(t, r2, 1, "cache still hot")

	// Wait for TTL to expire
	time.Sleep(400 * time.Millisecond)

	// Should now get fresh data from MongoDB
	r3, err := coll.Find(ctx, gmqb.Eq("active", true))
	require.NoError(t, err)
	assert.Len(t, r3, 2, "cache expired: fresh result expected")
}

// --- Write passthroughs ---

func TestCache_WritePassthrough_InsertAndDelete(t *testing.T) {
	coll := cachedUsers(t, time.Minute)
	ctx := context.Background()

	res, err := coll.InsertOne(ctx, &User{Name: "Alice", Age: 30})
	require.NoError(t, err)
	assert.NotNil(t, res.InsertedID)

	del, err := coll.DeleteOne(ctx, gmqb.Eq("name", "Alice"))
	require.NoError(t, err)
	assert.Equal(t, int64(1), del.DeletedCount)
}

// --- CacheMetrics ---

func TestCache_WrapWithCacheAndMetrics_ReturnsProvider(t *testing.T) {
	client := gocache.New(5*time.Minute, 10*time.Minute)
	st := gocachestore.NewGoCache(client)
	mgr := cache.New[[]byte](st)

	var hitCount atomic.Int64
	mp := &stubMetrics{onHit: func() { hitCount.Add(1) }}

	coll := gmqb.WrapWithCacheAndMetrics[User](
		testDB.Collection(t.Name()),
		mgr,
		time.Minute,
		mp,
	)

	assert.NotNil(t, coll.CacheMetrics(), "metrics provider should be accessible")
}

// stubMetrics is a minimal metrics.MetricsInterface for testing without Prometheus.
type stubMetrics struct {
	onHit func()
}

func (s *stubMetrics) RecordFromCodec(_ codec.CodecInterface) {
	if s.onHit != nil {
		s.onHit()
	}
}
