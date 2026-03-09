package gmqb

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/metrics"
	"github.com/eko/gocache/lib/v4/store"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// CachedCollection wraps Collection[T] and adds a transparent read-through
// cache for Find, FindOne, CountDocuments, and CachedAggregate calls.
// Write operations (InsertOne, UpdateOne, etc.) pass straight through to
// the underlying collection — cache invalidation is either TTL-based or
// driven by a ChangeStreamInvalidator.
//
// Example (in-memory, TTL-only):
//
//	store    := gocachestore.NewGoCache(gocache.New(5*time.Minute, 10*time.Minute))
//	cacheMgr := cache.New[[]byte](store)
//	coll     := gmqb.WrapWithCache[User](db.Collection("users"), cacheMgr, 5*time.Minute)
type CachedCollection[T any] struct {
	inner   *Collection[T]
	cache   cache.CacheInterface[[]byte]
	ttl     time.Duration
	metricP metrics.MetricsInterface // nil when metrics are not enabled
}

// WrapWithCache creates a CachedCollection around a raw mongo.Collection.
//
// Example:
//
//	coll := gmqb.WrapWithCache[User](rawColl, cacheMgr, 5*time.Minute)
func WrapWithCache[T any](
	coll *mongo.Collection,
	cacheManager cache.CacheInterface[[]byte],
	ttl time.Duration,
) *CachedCollection[T] {
	return &CachedCollection[T]{
		inner: Wrap[T](coll),
		cache: cacheManager,
		ttl:   ttl,
	}
}

// WrapWithCacheAndMetrics is like WrapWithCache but wraps the cache manager
// with gocache's metric layer so hit/miss/set/delete counters are recorded.
//
// Example (Prometheus):
//
//	promMetrics := metrics.NewPrometheus("gmqb")
//	coll := gmqb.WrapWithCacheAndMetrics[User](db.Collection("users"), cacheMgr, 5*time.Minute, promMetrics)
func WrapWithCacheAndMetrics[T any](
	coll *mongo.Collection,
	cacheManager cache.CacheInterface[[]byte],
	ttl time.Duration,
	metricsProvider metrics.MetricsInterface,
) *CachedCollection[T] {
	metricCache := cache.NewMetric[[]byte](metricsProvider, cacheManager)
	return &CachedCollection[T]{
		inner:   Wrap[T](coll),
		cache:   metricCache,
		ttl:     ttl,
		metricP: metricsProvider,
	}
}

// CacheMetrics returns the metrics provider supplied at construction time,
// or nil if metrics were not enabled.
func (c *CachedCollection[T]) CacheMetrics() metrics.MetricsInterface {
	return c.metricP
}

// Unwrap returns the underlying typed Collection[T].
func (c *CachedCollection[T]) Unwrap() *Collection[T] {
	return c.inner
}

// collectionTag returns the gocache tag used for all entries belonging to this
// collection so InvalidateByTags can flush them all at once.
func (c *CachedCollection[T]) collectionTag() string {
	col := c.inner.Unwrap()
	return col.Database().Name() + "." + col.Name()
}

// cacheKey builds a deterministic SHA-256 cache key from a human-readable
// prefix and an arbitrary set of values (marshalled to JSON).
func cacheKey(prefix string, parts ...any) (string, error) {
	h := sha256.New()
	h.Write([]byte(prefix))
	for _, p := range parts {
		b, err := json.Marshal(p)
		if err != nil {
			return "", fmt.Errorf("gmqb cache key: marshal %T: %w", p, err)
		}
		h.Write(b)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// filterKey serialises a Filter to a stable string via Extended JSON.
func filterKey(f Filter) (string, error) {
	raw, err := bson.MarshalExtJSON(f.BsonD(), false, false)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

// pipelineKey serialises a Pipeline into a stable Extended JSON string.
func pipelineKey(p Pipeline) (string, error) {
	// Wrap the pipeline in a root document because bson.MarshalExtJSON
	// cannot directly marshal an array at the top level.
	doc := bson.D{{Key: "pipeline", Value: p.BsonD()}}
	b, err := bson.MarshalExtJSON(doc, true, false)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// findOptsKey serialises FindOpts to a string for inclusion in the cache key.
func findOptsKey(opts []FindOpt) (string, error) {
	fo := buildFindOpts(opts)
	b, err := json.Marshal(fo)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// --- Read operations ---

// Find returns all documents matching the filter, serving from cache on a hit.
func (c *CachedCollection[T]) Find(ctx context.Context, filter Filter, opts ...FindOpt) ([]T, error) {
	fk, err := filterKey(filter)
	if err != nil {
		return c.inner.Find(ctx, filter, opts...)
	}
	ok, err := findOptsKey(opts)
	if err != nil {
		return c.inner.Find(ctx, filter, opts...)
	}
	key, err := cacheKey("find:"+c.collectionTag(), fk, ok)
	if err != nil {
		return c.inner.Find(ctx, filter, opts...)
	}

	if raw, cErr := c.cache.Get(ctx, key); cErr == nil {
		var results []T
		if json.Unmarshal(raw, &results) == nil {
			return results, nil
		}
	}

	results, err := c.inner.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}

	if raw, mErr := json.Marshal(results); mErr == nil {
		_ = c.cache.Set(ctx, key, raw,
			store.WithExpiration(c.ttl),
			store.WithTags([]string{c.collectionTag()}),
		)
	}
	return results, nil
}

// FindOne returns a single document matching the filter, serving from cache on a hit.
// Returns mongo.ErrNoDocuments if no document matches.
func (c *CachedCollection[T]) FindOne(ctx context.Context, filter Filter, opts ...FindOpt) (*T, error) {
	fk, err := filterKey(filter)
	if err != nil {
		return c.inner.FindOne(ctx, filter, opts...)
	}
	ok, err2 := findOptsKey(opts)
	if err2 != nil {
		return c.inner.FindOne(ctx, filter, opts...)
	}
	key, err3 := cacheKey("findone:"+c.collectionTag(), fk, string(ok))
	if err3 != nil {
		return c.inner.FindOne(ctx, filter, opts...)
	}

	if raw, cErr := c.cache.Get(ctx, key); cErr == nil {
		// stored as a 1-element slice to distinguish "not found" from cache miss
		var results []*T
		if json.Unmarshal(raw, &results) == nil && len(results) == 1 {
			return results[0], nil
		}
	}

	result, err := c.inner.FindOne(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}

	if raw, mErr := json.Marshal([]*T{result}); mErr == nil {
		_ = c.cache.Set(ctx, key, raw,
			store.WithExpiration(c.ttl),
			store.WithTags([]string{c.collectionTag()}),
		)
	}
	return result, nil
}

// CountDocuments returns the number of documents matching the filter, serving from cache on a hit.
func (c *CachedCollection[T]) CountDocuments(ctx context.Context, filter Filter) (int64, error) {
	fk, err := filterKey(filter)
	if err != nil {
		return c.inner.CountDocuments(ctx, filter)
	}
	key, err2 := cacheKey("count:"+c.collectionTag(), fk)
	if err2 != nil {
		return c.inner.CountDocuments(ctx, filter)
	}

	if raw, cErr := c.cache.Get(ctx, key); cErr == nil && len(raw) == 8 {
		return int64(binary.BigEndian.Uint64(raw)), nil
	}

	count, err := c.inner.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(count))
	_ = c.cache.Set(ctx, key, buf,
		store.WithExpiration(c.ttl),
		store.WithTags([]string{c.collectionTag()}),
	)
	return count, nil
}

// CachedAggregate runs an aggregation pipeline with caching.
// It mirrors the package-level Aggregate[R, T] function.
//
// Example:
//
//	stats, err := gmqb.CachedAggregate[Stats](coll, ctx,
//	    gmqb.NewPipeline().Group(gmqb.GroupSpec("$country", gmqb.GroupAcc("count", gmqb.AccSum(1)))),
//	)
func CachedAggregate[R any, T any](c *CachedCollection[T], ctx context.Context, pipeline Pipeline) ([]R, error) {
	pk, err := pipelineKey(pipeline)
	if err != nil {
		return Aggregate[R](c.inner, ctx, pipeline)
	}
	key, err2 := cacheKey("aggregate:"+c.collectionTag(), pk)
	if err2 != nil {
		return Aggregate[R](c.inner, ctx, pipeline)
	}

	if raw, cErr := c.cache.Get(ctx, key); cErr == nil {
		var results []R
		if json.Unmarshal(raw, &results) == nil {
			return results, nil
		}
	}

	results, err := Aggregate[R](c.inner, ctx, pipeline)
	if err != nil {
		return nil, err
	}

	if raw, mErr := json.Marshal(results); mErr == nil {
		_ = c.cache.Set(ctx, key, raw,
			store.WithExpiration(c.ttl),
			store.WithTags([]string{c.collectionTag()}),
		)
	}
	return results, nil
}

// --- Write passthrough ---

// InsertOne inserts a single document (no caching).
func (c *CachedCollection[T]) InsertOne(ctx context.Context, doc *T) (*mongo.InsertOneResult, error) {
	return c.inner.InsertOne(ctx, doc)
}

// InsertMany inserts multiple documents (no caching).
func (c *CachedCollection[T]) InsertMany(ctx context.Context, docs []T) (*mongo.InsertManyResult, error) {
	return c.inner.InsertMany(ctx, docs)
}

// UpdateOne updates a single matching document (no caching).
func (c *CachedCollection[T]) UpdateOne(ctx context.Context, filter Filter, update Updater, opts ...UpdateOpt) (*mongo.UpdateResult, error) {
	return c.inner.UpdateOne(ctx, filter, update, opts...)
}

// UpdateMany updates all matching documents (no caching).
func (c *CachedCollection[T]) UpdateMany(ctx context.Context, filter Filter, update Updater, opts ...UpdateManyOpt) (*mongo.UpdateResult, error) {
	return c.inner.UpdateMany(ctx, filter, update, opts...)
}

// DeleteOne deletes a single matching document (no caching).
func (c *CachedCollection[T]) DeleteOne(ctx context.Context, filter Filter) (*mongo.DeleteResult, error) {
	return c.inner.DeleteOne(ctx, filter)
}

// DeleteMany deletes all matching documents (no caching).
func (c *CachedCollection[T]) DeleteMany(ctx context.Context, filter Filter) (*mongo.DeleteResult, error) {
	return c.inner.DeleteMany(ctx, filter)
}

// FindOneAndDelete deletes a single matching document and returns it (no caching).
func (c *CachedCollection[T]) FindOneAndDelete(ctx context.Context, filter Filter, opts ...FindOneAndDeleteOpt) (*T, error) {
	return c.inner.FindOneAndDelete(ctx, filter, opts...)
}

// FindOneAndUpdate updates a single matching document and returns it (no caching).
func (c *CachedCollection[T]) FindOneAndUpdate(ctx context.Context, filter Filter, update Updater, opts ...FindOneAndUpdateOpt) (*T, error) {
	return c.inner.FindOneAndUpdate(ctx, filter, update, opts...)
}

// FindOneAndReplace replaces a single matching document and returns it (no caching).
func (c *CachedCollection[T]) FindOneAndReplace(ctx context.Context, filter Filter, replacement *T, opts ...FindOneAndReplaceOpt) (*T, error) {
	return c.inner.FindOneAndReplace(ctx, filter, replacement, opts...)
}

// BulkWrite performs multiple write operations (no caching).
func (c *CachedCollection[T]) BulkWrite(ctx context.Context, models []WriteModel[T], opts ...BulkWriteOpt) (*mongo.BulkWriteResult, error) {
	return c.inner.BulkWrite(ctx, models, opts...)
}

// Distinct returns distinct values for a field (no caching).
func (c *CachedCollection[T]) Distinct(ctx context.Context, field string, filter Filter) *mongo.DistinctResult {
	return c.inner.Distinct(ctx, field, filter)
}

// InvalidateCache flushes all cached entries for this collection.
// Useful for manual invalidation in tests or after bulk operations.
func (c *CachedCollection[T]) InvalidateCache(ctx context.Context) error {
	return c.cache.Invalidate(ctx, store.WithInvalidateTags([]string{c.collectionTag()}))
}
