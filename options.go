package gmqb

import (
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// --- Find Options ---

// FindOpt is a functional option for configuring find operations.
type FindOpt func(*options.FindOptionsBuilder)

// WithSort sets the sort specification for a find operation.
//
// Example:
//
//	coll.Find(ctx, filter, gmqb.WithSort(gmqb.Desc("createdAt")))
func WithSort(sort interface{}) FindOpt {
	return func(o *options.FindOptionsBuilder) {
		o.SetSort(sort)
	}
}

// WithProjection sets the projection for a find operation.
//
// Example:
//
//	coll.Find(ctx, filter, gmqb.WithProjection(gmqb.Include("name", "email")))
func WithProjection(projection interface{}) FindOpt {
	return func(o *options.FindOptionsBuilder) {
		o.SetProjection(projection)
	}
}

// WithSkip sets the number of documents to skip.
//
// Example:
//
//	coll.Find(ctx, filter, gmqb.WithSkip(20))
func WithSkip(n int64) FindOpt {
	return func(o *options.FindOptionsBuilder) {
		o.SetSkip(n)
	}
}

// WithLimit sets the maximum number of documents to return.
//
// Example:
//
//	coll.Find(ctx, filter, gmqb.WithLimit(10))
func WithLimit(n int64) FindOpt {
	return func(o *options.FindOptionsBuilder) {
		o.SetLimit(n)
	}
}

// buildFindOpts applies functional options to a FindOptionsBuilder.
func buildFindOpts(opts []FindOpt) *options.FindOptionsBuilder {
	o := options.Find()
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// --- Update Options ---

// UpdateOpt is a functional option for configuring update operations.
type UpdateOpt func(*options.UpdateOneOptionsBuilder)

// WithUpsert sets the upsert flag. When true, creates a new document if no match.
//
// Example:
//
//	coll.UpdateOne(ctx, filter, update, gmqb.WithUpsert(true))
func WithUpsert(upsert bool) UpdateOpt {
	return func(o *options.UpdateOneOptionsBuilder) {
		o.SetUpsert(upsert)
	}
}

// buildUpdateOneOpts applies functional options to an UpdateOneOptionsBuilder.
func buildUpdateOneOpts(opts []UpdateOpt) *options.UpdateOneOptionsBuilder {
	o := options.UpdateOne()
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// UpdateManyOpt is a functional option for configuring UpdateMany operations.
type UpdateManyOpt func(*options.UpdateManyOptionsBuilder)

// WithUpsertMany sets the upsert flag for UpdateMany operations.
func WithUpsertMany(upsert bool) UpdateManyOpt {
	return func(o *options.UpdateManyOptionsBuilder) {
		o.SetUpsert(upsert)
	}
}

// buildUpdateManyOpts applies functional options to an UpdateManyOptionsBuilder.
func buildUpdateManyOpts(opts []UpdateManyOpt) *options.UpdateManyOptionsBuilder {
	o := options.UpdateMany()
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// --- Bulk Write Options ---

// BulkWriteOpt is a functional option for configuring bulk write operations.
type BulkWriteOpt func(*options.BulkWriteOptionsBuilder)

// WithOrdered sets the ordered flag. When true, operations are executed serially.
//
// Example:
//
//	coll.BulkWrite(ctx, models, gmqb.WithOrdered(false))
func WithOrdered(ordered bool) BulkWriteOpt {
	return func(o *options.BulkWriteOptionsBuilder) {
		o.SetOrdered(ordered)
	}
}

// buildBulkWriteOpts applies functional options to a BulkWriteOptionsBuilder.
func buildBulkWriteOpts(opts []BulkWriteOpt) *options.BulkWriteOptionsBuilder {
	o := options.BulkWrite()
	for _, opt := range opts {
		opt(o)
	}
	return o
}
