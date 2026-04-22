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

// WithCollation sets the collation for an update operation.
func WithCollation(collation *options.Collation) UpdateOpt {
	return func(o *options.UpdateOneOptionsBuilder) {
		o.SetCollation(collation)
	}
}

// WithHint sets the hint for an update operation.
func WithHint(hint interface{}) UpdateOpt {
	return func(o *options.UpdateOneOptionsBuilder) {
		o.SetHint(hint)
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

// --- Replace Options ---

// ReplaceOpt is a functional option for configuring replace operations.
type ReplaceOpt func(*options.ReplaceOptionsBuilder)

// WithUpsertReplace sets the upsert flag for replace operations.
func WithUpsertReplace(upsert bool) ReplaceOpt {
	return func(o *options.ReplaceOptionsBuilder) {
		o.SetUpsert(upsert)
	}
}

// buildReplaceOpts applies functional options to a ReplaceOptionsBuilder.
func buildReplaceOpts(opts []ReplaceOpt) *options.ReplaceOptionsBuilder {
	o := options.Replace()
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

// --- FindOneAndDelete Options ---

// FindOneAndDeleteOpt is a functional option for configuring FindOneAndDelete operations.
type FindOneAndDeleteOpt func(*options.FindOneAndDeleteOptionsBuilder)

// WithSortFindAndDelete sets the sort specification for FindOneAndDelete.
func WithSortFindAndDelete(sort interface{}) FindOneAndDeleteOpt {
	return func(o *options.FindOneAndDeleteOptionsBuilder) {
		o.SetSort(sort)
	}
}

// WithProjectionFindAndDelete sets the projection for FindOneAndDelete.
func WithProjectionFindAndDelete(projection interface{}) FindOneAndDeleteOpt {
	return func(o *options.FindOneAndDeleteOptionsBuilder) {
		o.SetProjection(projection)
	}
}

// WithHintFindAndDelete sets the hint for FindOneAndDelete.
func WithHintFindAndDelete(hint interface{}) FindOneAndDeleteOpt {
	return func(o *options.FindOneAndDeleteOptionsBuilder) {
		o.SetHint(hint)
	}
}

// WithReturnDocumentDelete is a no-op for FindOneAndDelete as it always returns
// the document before deletion. Provided for API parity.
func WithReturnDocumentDelete(rd options.ReturnDocument) FindOneAndDeleteOpt {
	return func(o *options.FindOneAndDeleteOptionsBuilder) {}
}

// buildFindOneAndDeleteOpts applies functional options to a FindOneAndDeleteOptionsBuilder.
func buildFindOneAndDeleteOpts(opts []FindOneAndDeleteOpt) *options.FindOneAndDeleteOptionsBuilder {
	o := options.FindOneAndDelete()
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// --- FindOneAndUpdate Options ---

// FindOneAndUpdateOpt is a functional option for configuring FindOneAndUpdate operations.
type FindOneAndUpdateOpt func(*options.FindOneAndUpdateOptionsBuilder)

// WithSortFindAndUpdate sets the sort specification for FindOneAndUpdate.
func WithSortFindAndUpdate(sort interface{}) FindOneAndUpdateOpt {
	return func(o *options.FindOneAndUpdateOptionsBuilder) {
		o.SetSort(sort)
	}
}

// WithProjectionFindAndUpdate sets the projection for FindOneAndUpdate.
func WithProjectionFindAndUpdate(projection interface{}) FindOneAndUpdateOpt {
	return func(o *options.FindOneAndUpdateOptionsBuilder) {
		o.SetProjection(projection)
	}
}

// WithHintFindAndUpdate sets the hint for FindOneAndUpdate.
func WithHintFindAndUpdate(hint interface{}) FindOneAndUpdateOpt {
	return func(o *options.FindOneAndUpdateOptionsBuilder) {
		o.SetHint(hint)
	}
}

// WithReturnDocument specifies whether to return the original or the updated document.
//
// Example:
//
//	coll.FindOneAndUpdate(ctx, filter, update, gmqb.WithReturnDocument(options.After))
func WithReturnDocument(rd options.ReturnDocument) FindOneAndUpdateOpt {
	return func(o *options.FindOneAndUpdateOptionsBuilder) {
		o.SetReturnDocument(rd)
	}
}

// WithUpsertFindAndUpdate sets the upsert flag for FindOneAndUpdate.
func WithUpsertFindAndUpdate(upsert bool) FindOneAndUpdateOpt {
	return func(o *options.FindOneAndUpdateOptionsBuilder) {
		o.SetUpsert(upsert)
	}
}

// buildFindOneAndUpdateOpts applies functional options to a FindOneAndUpdateOptionsBuilder.
func buildFindOneAndUpdateOpts(opts []FindOneAndUpdateOpt) *options.FindOneAndUpdateOptionsBuilder {
	o := options.FindOneAndUpdate()
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// --- FindOneAndReplace Options ---

// FindOneAndReplaceOpt is a functional option for configuring FindOneAndReplace operations.
type FindOneAndReplaceOpt func(*options.FindOneAndReplaceOptionsBuilder)

// WithSortFindAndReplace sets the sort specification for FindOneAndReplace.
func WithSortFindAndReplace(sort interface{}) FindOneAndReplaceOpt {
	return func(o *options.FindOneAndReplaceOptionsBuilder) {
		o.SetSort(sort)
	}
}

// WithProjectionFindAndReplace sets the projection for FindOneAndReplace.
func WithProjectionFindAndReplace(projection interface{}) FindOneAndReplaceOpt {
	return func(o *options.FindOneAndReplaceOptionsBuilder) {
		o.SetProjection(projection)
	}
}

// WithHintFindAndReplace sets the hint for FindOneAndReplace.
func WithHintFindAndReplace(hint interface{}) FindOneAndReplaceOpt {
	return func(o *options.FindOneAndReplaceOptionsBuilder) {
		o.SetHint(hint)
	}
}

// WithReturnDocumentReplace specifies whether to return the original or the replaced document.
//
// Example:
//
//	coll.FindOneAndReplace(ctx, filter, replace, gmqb.WithReturnDocumentReplace(options.After))
func WithReturnDocumentReplace(rd options.ReturnDocument) FindOneAndReplaceOpt {
	return func(o *options.FindOneAndReplaceOptionsBuilder) {
		o.SetReturnDocument(rd)
	}
}

// WithUpsertFindAndReplace sets the upsert flag for FindOneAndReplace.
func WithUpsertFindAndReplace(upsert bool) FindOneAndReplaceOpt {
	return func(o *options.FindOneAndReplaceOptionsBuilder) {
		o.SetUpsert(upsert)
	}
}

// buildFindOneAndReplaceOpts applies functional options to a FindOneAndReplaceOptionsBuilder.
func buildFindOneAndReplaceOpts(opts []FindOneAndReplaceOpt) *options.FindOneAndReplaceOptionsBuilder {
	o := options.FindOneAndReplace()
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// --- Count Options ---

// CountOpt is a functional option for configuring count operations.
type CountOpt func(*options.CountOptionsBuilder)

// WithLimitCount sets the maximum number of documents to count.
func WithLimitCount(n int64) CountOpt {
	return func(o *options.CountOptionsBuilder) {
		o.SetLimit(n)
	}
}

// WithSkipCount sets the number of documents to skip before counting.
func WithSkipCount(n int64) CountOpt {
	return func(o *options.CountOptionsBuilder) {
		o.SetSkip(n)
	}
}

// buildCountOpts applies functional options to a CountOptionsBuilder.
func buildCountOpts(opts []CountOpt) *options.CountOptionsBuilder {
	o := options.Count()
	for _, opt := range opts {
		opt(o)
	}
	return o
}

