package gmqb

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ChangeStreamInvalidator watches a MongoDB change stream and invalidates the
// gocache tag for a collection whenever an insert, update, replace, or delete
// operation is detected.
//
// This is optional — only use it when your MongoDB deployment is a replica set
// or Atlas cluster. For standalone mongod, rely on TTL-based cache expiry instead.
//
// The resume token is stored in the same cache backend under a reserved key with
// no expiry, so the stream can resume after a restart without an extra state store.
//
// Example:
//
//	inv := gmqb.NewChangeStreamInvalidator(db.Collection("users"), cacheMgr)
//	go func() {
//	    if err := inv.Watch(ctx); err != nil && ctx.Err() == nil {
//	        log.Printf("change stream stopped: %v", err)
//	    }
//	}()
type ChangeStreamInvalidator struct {
	coll           *mongo.Collection
	cache          cache.CacheInterface[[]byte]
	tag            string // e.g. "mydb.users"
	resumeTokenKey string // cache key for persisting the resume token
}

// NewChangeStreamInvalidator creates a ChangeStreamInvalidator for coll.
func NewChangeStreamInvalidator(coll *mongo.Collection, cacheManager cache.CacheInterface[[]byte]) *ChangeStreamInvalidator {
	tag := coll.Database().Name() + "." + coll.Name()
	return &ChangeStreamInvalidator{
		coll:           coll,
		cache:          cacheManager,
		tag:            tag,
		resumeTokenKey: "gmqb::resumetoken::" + tag,
	}
}

// Watch opens and monitors the change stream. It blocks until ctx is cancelled.
// Reconnects automatically with exponential back-off (max 30 s) on stream errors.
// Returns nil when ctx is cancelled, or the last non-recoverable error.
func (inv *ChangeStreamInvalidator) Watch(ctx context.Context) error {
	backoff := time.Second
	const maxBackoff = 30 * time.Second

	for {
		if err := ctx.Err(); err != nil {
			return nil
		}

		err := inv.watch(ctx)
		if ctx.Err() != nil {
			// Context cancelled — clean exit.
			return nil
		}

		// Transient error — back off and retry.
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(backoff):
		}
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
		_ = err // err is only informational; we always retry
	}
}

// watch runs a single change stream session.
func (inv *ChangeStreamInvalidator) watch(ctx context.Context) error {
	pipeline := mongo.Pipeline{
		{{
			Key: "$match",
			Value: bson.D{{
				Key:   "operationType",
				Value: bson.D{{Key: "$in", Value: bson.A{"insert", "update", "replace", "delete"}}},
			}},
		}},
	}

	watchOpts := options.ChangeStream().SetFullDocument(options.UpdateLookup)

	// Attempt to resume from a stored token.
	if token := inv.loadResumeToken(ctx); token != nil {
		watchOpts.SetResumeAfter(token)
	}

	cs, err := inv.coll.Watch(ctx, pipeline, watchOpts)
	if err != nil {
		return fmt.Errorf("gmqb invalidator: open change stream: %w", err)
	}
	defer cs.Close(ctx)

	for cs.Next(ctx) {
		// Persist the resume token so we can pick up where we left off.
		inv.saveResumeToken(ctx, cs.ResumeToken())

		// Invalidate ALL cached queries for this collection.
		if err := inv.cache.Invalidate(ctx, store.WithInvalidateTags([]string{inv.tag})); err != nil {
			// Non-fatal: log and continue.
			_ = err
		}
	}

	if err := cs.Err(); err != nil && ctx.Err() == nil {
		return err
	}
	return nil
}

// saveResumeToken persists the resume token to the cache with no expiry.
func (inv *ChangeStreamInvalidator) saveResumeToken(ctx context.Context, token bson.Raw) {
	if token == nil {
		return
	}
	b, err := json.Marshal(token)
	if err != nil {
		return
	}
	_ = inv.cache.Set(ctx, inv.resumeTokenKey, b, store.WithExpiration(0))
}

// loadResumeToken retrieves a previously persisted resume token, or nil.
func (inv *ChangeStreamInvalidator) loadResumeToken(ctx context.Context) bson.Raw {
	raw, err := inv.cache.Get(ctx, inv.resumeTokenKey)
	if err != nil || len(raw) == 0 {
		return nil
	}
	var token bson.Raw
	if err := json.Unmarshal(raw, &token); err != nil {
		return nil
	}
	return token
}
