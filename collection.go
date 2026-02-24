package gmqb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Collection is a type-safe wrapper around mongo.Collection that accepts
// gmqb builders (Filter, Updater, Pipeline) instead of raw bson.D.
// The type parameter T specifies the document struct type.
//
// See: https://www.mongodb.com/docs/drivers/go/current/fundamentals/crud/
//
// Example:
//
//	type User struct {
//	    Name string `bson:"name"`
//	    Age  int    `bson:"age"`
//	}
//	coll := gmqb.Wrap[User](db.Collection("users"))
//	users, err := coll.Find(ctx, gmqb.Gte("age", 18))
type Collection[T any] struct {
	coll *mongo.Collection
}

// Wrap creates a typed Collection wrapper around a mongo.Collection.
//
// Example:
//
//	coll := gmqb.Wrap[User](db.Collection("users"))
func Wrap[T any](coll *mongo.Collection) *Collection[T] {
	return &Collection[T]{coll: coll}
}

// Unwrap returns the underlying mongo.Collection for direct driver access.
func (c *Collection[T]) Unwrap() *mongo.Collection {
	return c.coll
}

// Find returns all documents matching the filter.
//
// MongoDB equivalent: db.collection.find(filter, opts)
//
// See: https://www.mongodb.com/docs/drivers/go/current/fundamentals/crud/read-operations/query-document/
//
// Example:
//
//	users, err := coll.Find(ctx, gmqb.Eq("active", true),
//	    gmqb.WithLimit(10),
//	    gmqb.WithSort(gmqb.Desc("createdAt")),
//	)
func (c *Collection[T]) Find(ctx context.Context, filter Filter, opts ...FindOpt) ([]T, error) {
	findOpts := buildFindOpts(opts)
	cursor, err := c.coll.Find(ctx, filter.BsonD(), findOpts)
	if err != nil {
		return nil, err
	}
	var results []T
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// FindOne returns a single document matching the filter.
// Returns mongo.ErrNoDocuments if no document matches.
//
// See: https://www.mongodb.com/docs/drivers/go/current/fundamentals/crud/read-operations/query-document/
//
// Example:
//
//	user, err := coll.FindOne(ctx, gmqb.Eq("email", "alice@example.com"))
func (c *Collection[T]) FindOne(ctx context.Context, filter Filter, opts ...FindOpt) (*T, error) {
	findOneOpts := options.FindOne()
	for _, opt := range opts {
		// Convert FindOpt to FindOneOptionsBuilder settings
		builder := options.Find()
		opt(builder)
		// Extract values from the builder via List()
		for _, fn := range builder.List() {
			var fo options.FindOptions
			_ = fn(&fo)
			if fo.Sort != nil {
				findOneOpts.SetSort(fo.Sort)
			}
			if fo.Projection != nil {
				findOneOpts.SetProjection(fo.Projection)
			}
			if fo.Skip != nil {
				findOneOpts.SetSkip(*fo.Skip)
			}
		}
	}

	var result T
	err := c.coll.FindOne(ctx, filter.BsonD(), findOneOpts).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// InsertOne inserts a single document.
//
// See: https://www.mongodb.com/docs/drivers/go/current/fundamentals/crud/write-operations/insert/
//
// Example:
//
//	result, err := coll.InsertOne(ctx, &User{Name: "Alice", Age: 30})
func (c *Collection[T]) InsertOne(ctx context.Context, doc *T) (*mongo.InsertOneResult, error) {
	return c.coll.InsertOne(ctx, doc)
}

// InsertMany inserts multiple documents.
//
// See: https://www.mongodb.com/docs/drivers/go/current/fundamentals/crud/write-operations/insert/
//
// Example:
//
//	result, err := coll.InsertMany(ctx, []User{
//	    {Name: "Alice", Age: 30},
//	    {Name: "Bob", Age: 25},
//	})
func (c *Collection[T]) InsertMany(ctx context.Context, docs []T) (*mongo.InsertManyResult, error) {
	ifaces := make([]interface{}, len(docs))
	for i := range docs {
		ifaces[i] = docs[i]
	}
	return c.coll.InsertMany(ctx, ifaces)
}

// BulkWrite performs multiple write operations in a single batch.
//
// See: https://www.mongodb.com/docs/drivers/go/current/fundamentals/crud/write-operations/bulk/
//
// Example:
//
//	models := []gmqb.WriteModel[User]{
//	    gmqb.NewInsertOneModel[User]().SetDocument(&User{Name: "Alice"}),
//	    gmqb.NewUpdateOneModel[User]().SetFilter(gmqb.Eq("name", "Bob")).SetUpdate(gmqb.NewUpdate().Set("age", 25)),
//	}
//	result, err := coll.BulkWrite(ctx, models)
func (c *Collection[T]) BulkWrite(ctx context.Context, models []WriteModel[T], opts ...BulkWriteOpt) (*mongo.BulkWriteResult, error) {
	if len(models) == 0 {
		return nil, nil // Return empty if no models specified
	}
	mongoModels := make([]mongo.WriteModel, len(models))
	for i, m := range models {
		mongoModels[i] = m.MongoWriteModel()
	}
	bwOpts := buildBulkWriteOpts(opts)
	return c.coll.BulkWrite(ctx, mongoModels, bwOpts)
}

// UpdateOne updates a single document matching the filter.
//
// See: https://www.mongodb.com/docs/drivers/go/current/fundamentals/crud/write-operations/modify/
//
// Example:
//
//	result, err := coll.UpdateOne(ctx,
//	    gmqb.Eq("name", "Alice"),
//	    gmqb.NewUpdate().Set("age", 31),
//	)
func (c *Collection[T]) UpdateOne(ctx context.Context, filter Filter, update Updater, opts ...UpdateOpt) (*mongo.UpdateResult, error) {
	if filter.IsEmpty() {
		return nil, fmt.Errorf("%w: UpdateOne requires a non-empty filter", ErrEmptyFilter)
	}
	if update.IsEmpty() {
		return nil, fmt.Errorf("%w: UpdateOne requires a non-empty update", ErrEmptyUpdate)
	}
	updateOpts := buildUpdateOneOpts(opts)
	return c.coll.UpdateOne(ctx, filter.BsonD(), update.BsonD(), updateOpts)
}

// UpdateMany updates all documents matching the filter.
//
// See: https://www.mongodb.com/docs/drivers/go/current/fundamentals/crud/write-operations/modify/
//
// Example:
//
//	result, err := coll.UpdateMany(ctx,
//	    gmqb.Lt("age", 18),
//	    gmqb.NewUpdate().Set("status", "minor"),
//	)
func (c *Collection[T]) UpdateMany(ctx context.Context, filter Filter, update Updater, opts ...UpdateManyOpt) (*mongo.UpdateResult, error) {
	if filter.IsEmpty() {
		return nil, fmt.Errorf("%w: UpdateMany requires a non-empty filter", ErrEmptyFilter)
	}
	if update.IsEmpty() {
		return nil, fmt.Errorf("%w: UpdateMany requires a non-empty update", ErrEmptyUpdate)
	}
	updateOpts := buildUpdateManyOpts(opts)
	return c.coll.UpdateMany(ctx, filter.BsonD(), update.BsonD(), updateOpts)
}

// DeleteOne deletes a single document matching the filter.
//
// See: https://www.mongodb.com/docs/drivers/go/current/fundamentals/crud/write-operations/delete/
//
// Example:
//
//	result, err := coll.DeleteOne(ctx, gmqb.Eq("name", "Alice"))
func (c *Collection[T]) DeleteOne(ctx context.Context, filter Filter) (*mongo.DeleteResult, error) {
	if filter.IsEmpty() {
		return nil, fmt.Errorf("%w: DeleteOne requires a non-empty filter", ErrEmptyFilter)
	}
	return c.coll.DeleteOne(ctx, filter.BsonD())
}

// DeleteMany deletes all documents matching the filter.
//
// See: https://www.mongodb.com/docs/drivers/go/current/fundamentals/crud/write-operations/delete/
//
// Example:
//
//	result, err := coll.DeleteMany(ctx, gmqb.Eq("status", "inactive"))
func (c *Collection[T]) DeleteMany(ctx context.Context, filter Filter) (*mongo.DeleteResult, error) {
	if filter.IsEmpty() {
		return nil, fmt.Errorf("%w: DeleteMany requires a non-empty filter", ErrEmptyFilter)
	}
	return c.coll.DeleteMany(ctx, filter.BsonD())
}

// CountDocuments returns the number of documents matching the filter.
//
// See: https://www.mongodb.com/docs/drivers/go/current/fundamentals/crud/read-operations/count/
//
// Example:
//
//	count, err := coll.CountDocuments(ctx, gmqb.Gte("age", 18))
func (c *Collection[T]) CountDocuments(ctx context.Context, filter Filter) (int64, error) {
	return c.coll.CountDocuments(ctx, filter.BsonD())
}

// Distinct returns the distinct values for a specified field.
//
// See: https://www.mongodb.com/docs/drivers/go/current/fundamentals/crud/read-operations/distinct/
//
// Example:
//
//	result := coll.Distinct(ctx, "country", gmqb.Exists("country", true))
func (c *Collection[T]) Distinct(ctx context.Context, field string, filter Filter) *mongo.DistinctResult {
	return c.coll.Distinct(ctx, field, filter.BsonD())
}

// Aggregate runs an aggregation pipeline on the collection and returns typed results.
// The type parameter R can differ from the collection's T when the pipeline
// reshapes documents.
//
// See: https://www.mongodb.com/docs/drivers/go/current/fundamentals/crud/read-operations/aggregate/
//
// Example:
//
//	type Stats struct {
//	    Country string  `bson:"_id"`
//	    Count   int     `bson:"count"`
//	}
//	stats, err := gmqb.Aggregate[Stats](coll, ctx,
//	    gmqb.NewPipeline().
//	        Group(gmqb.GroupSpec("$country", gmqb.GroupAcc("count", gmqb.AccSum(1)))),
//	)
func Aggregate[R any, T any](c *Collection[T], ctx context.Context, pipeline Pipeline) ([]R, error) {
	if pipeline.IsEmpty() {
		return nil, fmt.Errorf("%w: Aggregate requires a non-empty pipeline", ErrEmptyPipeline)
	}
	cursor, err := c.coll.Aggregate(ctx, pipeline.BsonD())
	if err != nil {
		return nil, err
	}
	var results []R
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// --- Sort Helpers ---

// SortField represents a single sort field with direction.
type SortField struct {
	Field string
	Order int // 1 for ascending, -1 for descending
}

// SortRule is a convenience helper to create a SortField without struct syntax.
//
// Example:
//
//	gmqb.SortRule("createdAt", -1)
func SortRule(field string, order int) SortField {
	return SortField{Field: field, Order: order}
}

// SortSpec converts SortField values to a bson.D sort specification.
//
// Example:
//
//	spec := gmqb.SortSpec(
//	    gmqb.SortRule("age", -1),
//	    gmqb.SortRule("name", 1),
//	)
func SortSpec(fields ...SortField) bson.D {
	d := make(bson.D, len(fields))
	for i, f := range fields {
		d[i] = bson.E{Key: f.Field, Value: f.Order}
	}
	return d
}

// Include creates a projection that includes the specified fields (value 1).
//
// Example:
//
//	spec := gmqb.Include("name", "email")
func Include(fields ...string) bson.D {
	d := make(bson.D, len(fields))
	for i, f := range fields {
		d[i] = bson.E{Key: f, Value: 1}
	}
	return d
}

// Exclude creates a projection that excludes the specified fields (value 0).
//
// Example:
//
//	spec := gmqb.Exclude("password", "ssn")
func Exclude(fields ...string) bson.D {
	d := make(bson.D, len(fields))
	for i, f := range fields {
		d[i] = bson.E{Key: f, Value: 0}
	}
	return d
}
