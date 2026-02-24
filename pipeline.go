package gmqb

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Pipeline represents an immutable MongoDB aggregation pipeline.
// Each method appends a stage and returns a new Pipeline — the original is unchanged.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation-pipeline/
//
// Example:
//
//	pipeline := gmqb.NewPipeline().
//	    Match(gmqb.Eq("status", "active")).
//	    Group(gmqb.GroupSpec("$country", gmqb.GroupAcc("count", gmqb.AccSum(1)))).
//	    Sort(gmqb.Desc("count")).
//	    Limit(10)
type Pipeline struct {
	stages []bson.D
}

// NewPipeline creates an empty aggregation pipeline.
//
// Example:
//
//	p := gmqb.NewPipeline().Match(gmqb.Eq("status", "active"))
func NewPipeline() Pipeline {
	return Pipeline{}
}

// BsonD returns the pipeline as a []bson.D (mongo.Pipeline), suitable for passing
// to mongo.Collection.Aggregate().
func (p Pipeline) BsonD() []bson.D {
	return p.stages
}

// JSON returns the pipeline as a pretty-printed JSON array string.
func (p Pipeline) JSON() string {
	return pipelineToJSON(p.stages)
}

// CompactJSON returns the pipeline as a compact JSON array string.
func (p Pipeline) CompactJSON() string {
	return pipelineToCompactJSON(p.stages)
}

// IsEmpty returns true if the pipeline has no stages.
func (p Pipeline) IsEmpty() bool {
	return len(p.stages) == 0
}

// addStage appends a new stage and returns a new Pipeline.
func (p Pipeline) addStage(name string, value interface{}) Pipeline {
	newStages := make([]bson.D, len(p.stages), len(p.stages)+1)
	copy(newStages, p.stages)
	newStages = append(newStages, bson.D{{Key: name, Value: value}})
	return Pipeline{stages: newStages}
}

// --- Core Stages ---

// Match filters documents. Only documents matching the filter pass to the next stage.
//
// MongoDB equivalent:
//
//	{ $match: { <query> } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/match/
//
// Example:
//
//	p := gmqb.NewPipeline().Match(gmqb.Gte("age", 18))
func (p Pipeline) Match(filter Filter) Pipeline {
	return p.addStage("$match", filter.d)
}

// MatchRaw filters documents using a raw bson.D filter expression.
//
// Example:
//
//	p := gmqb.NewPipeline().MatchRaw(bson.D{{"status", "active"}})
func (p Pipeline) MatchRaw(filter bson.D) Pipeline {
	return p.addStage("$match", filter)
}

// Project reshapes each document, including, excluding, or computing new fields.
// The spec is a bson.D where field values are 1 (include), 0 (exclude), or an expression.
//
// MongoDB equivalent:
//
//	{ $project: { field1: 1, field2: 0, computed: <expression> } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/project/
//
// Example:
//
//	p := gmqb.NewPipeline().Project(gmqb.Include("name", "age"))
//	p := gmqb.NewPipeline().Project(gmqb.Exclude("password", "ssn"))
func (p Pipeline) Project(spec bson.D) Pipeline {
	return p.addStage("$project", spec)
}

// Group groups documents by a specified identifier and applies accumulator expressions.
//
// MongoDB equivalent:
//
//	{ $group: { _id: <expression>, <field1>: { <accumulator1>: <expr1> }, ... } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/group/
//
// Example:
//
//	// Single ID:
//	p := gmqb.NewPipeline().Group(gmqb.GroupSpec("$country",
//	    gmqb.GroupAcc("total", gmqb.AccSum(1)),
//	    gmqb.GroupAcc("avgAge", gmqb.AccAvg("$age")),
//	))
//
//	// Multiple fields ID:
//	p := gmqb.NewPipeline().Group(gmqb.GroupSpec(gmqb.GroupID("country", "city"),
//	    gmqb.GroupAcc("total", gmqb.AccSum(1)),
//	))
func (p Pipeline) Group(spec bson.D) Pipeline {
	return p.addStage("$group", spec)
}

// GroupAcc creates a field-expression pair (accumulator) for use in a $group stage.
//
// Example:
//
//	gmqb.GroupAcc("total", gmqb.AccSum(1))
func GroupAcc(field string, expr interface{}) bson.E {
	return bson.E{Key: field, Value: expr}
}

// GroupID is a convenience helper to group by multiple document fields
// by mapping each field name to its corresponding "$field" reference.
//
// Example:
//
//	gmqb.GroupID("country", "city") // produces: { country: "$country", city: "$city" }
func GroupID(fields ...string) bson.D {
	d := make(bson.D, len(fields))
	for i, f := range fields {
		d[i] = bson.E{Key: f, Value: "$" + f}
	}
	return d
}

// GroupSpec builds a specification for the $group stage.
// The id parameter specifies the _id expression to group by. It can be a simple
// string (e.g., "$country"), a bson.D for complex compound keys, or generated
// via GroupID("country", "city").
//
// The accumulators are built using the GroupAcc helper.
//
// Example:
//
//	spec := gmqb.GroupSpec(gmqb.GroupID("country", "city"),
//	    gmqb.GroupAcc("total", gmqb.AccSum(1)),
//	    gmqb.GroupAcc("avgAge", gmqb.AccAvg("$age")),
//	)
func GroupSpec(id interface{}, accumulators ...bson.E) bson.D {
	d := make(bson.D, len(accumulators)+1)
	d[0] = bson.E{Key: "_id", Value: id}
	for i, acc := range accumulators {
		d[i+1] = acc
	}
	return d
}

// Sort reorders documents by the specified sort key(s). Use 1 for ascending, -1 for descending.
//
// MongoDB equivalent:
//
//	{ $sort: { field1: 1, field2: -1 } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/sort/
//
// Example:
//
//	// Simple sorting:
//	p := gmqb.NewPipeline().Sort(gmqb.Asc("name", "age"))
//
//	// Mixed-order sorting:
//	p := gmqb.NewPipeline().Sort(gmqb.SortSpec(
//	    gmqb.SortRule("age", -1),
//	    gmqb.SortRule("name", 1),
//	))
func (p Pipeline) Sort(spec bson.D) Pipeline {
	return p.addStage("$sort", spec)
}

// Asc is a convenience helper that creates an ascending sort spec for one or more fields.
//
// Example:
//
//	p := gmqb.NewPipeline().Sort(gmqb.Asc("name", "age"))
func Asc(fields ...string) bson.D {
	d := make(bson.D, len(fields))
	for i, f := range fields {
		d[i] = bson.E{Key: f, Value: 1}
	}
	return d
}

// Desc is a convenience helper that creates a descending sort spec for one or more fields.
//
// Example:
//
//	p := gmqb.NewPipeline().Sort(gmqb.Desc("createdAt", "score"))
func Desc(fields ...string) bson.D {
	d := make(bson.D, len(fields))
	for i, f := range fields {
		d[i] = bson.E{Key: f, Value: -1}
	}
	return d
}

// Limit passes only the first n documents to the next stage.
//
// MongoDB equivalent:
//
//	{ $limit: n }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/limit/
//
// Example:
//
//	p := gmqb.NewPipeline().Sort(gmqb.Desc("score")).Limit(10)
func (p Pipeline) Limit(n int64) Pipeline {
	return p.addStage("$limit", n)
}

// Skip skips the first n documents.
//
// MongoDB equivalent:
//
//	{ $skip: n }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/skip/
//
// Example:
//
//	p := gmqb.NewPipeline().Skip(20).Limit(10) // page 3
func (p Pipeline) Skip(n int64) Pipeline {
	return p.addStage("$skip", n)
}

// Unwind deconstructs an array field, outputting one document per array element.
//
// MongoDB equivalent (simple):
//
//	{ $unwind: "$field" }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/unwind/
//
// Example:
//
//	p := gmqb.NewPipeline().Unwind("$tags")
func (p Pipeline) Unwind(path string) Pipeline {
	return p.addStage("$unwind", path)
}

// UnwindOpts provides options for the $unwind stage.
type UnwindOpts struct {
	// Path is the array field path (e.g. "$tags"). Required.
	Path string
	// IncludeArrayIndex is the name of a new field to hold the array index.
	IncludeArrayIndex string
	// PreserveNullAndEmptyArrays keeps documents with null/empty/missing arrays.
	PreserveNullAndEmptyArrays bool
}

// UnwindWithOpts deconstructs an array field with additional options.
//
// MongoDB equivalent:
//
//	{ $unwind: { path: "$field", includeArrayIndex: "idx", preserveNullAndEmptyArrays: true } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/unwind/
//
// Example:
//
//	p := gmqb.NewPipeline().UnwindWithOpts(gmqb.UnwindOpts{
//	    Path:                       "$items",
//	    IncludeArrayIndex:          "itemIndex",
//	    PreserveNullAndEmptyArrays: true,
//	})
func (p Pipeline) UnwindWithOpts(opts UnwindOpts) Pipeline {
	doc := bson.D{{Key: "path", Value: opts.Path}}
	if opts.IncludeArrayIndex != "" {
		doc = append(doc, bson.E{Key: "includeArrayIndex", Value: opts.IncludeArrayIndex})
	}
	if opts.PreserveNullAndEmptyArrays {
		doc = append(doc, bson.E{Key: "preserveNullAndEmptyArrays", Value: true})
	}
	return p.addStage("$unwind", doc)
}

// LookupOpts configures the $lookup stage for left outer joins.
type LookupOpts struct {
	// From is the foreign collection name.
	From string
	// LocalField is the field from the input documents.
	LocalField string
	// ForeignField is the field from the "from" collection.
	ForeignField string
	// As is the output array field name.
	As string
}

// Lookup performs a left outer join to another collection.
//
// MongoDB equivalent:
//
//	{ $lookup: { from: "coll", localField: "f1", foreignField: "f2", as: "output" } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/lookup/
//
// Example:
//
//	p := gmqb.NewPipeline().Lookup(gmqb.LookupOpts{
//	    From:         "orders",
//	    LocalField:   "_id",
//	    ForeignField: "userId",
//	    As:           "userOrders",
//	})
func (p Pipeline) Lookup(opts LookupOpts) Pipeline {
	doc := bson.D{
		{Key: "from", Value: opts.From},
		{Key: "localField", Value: opts.LocalField},
		{Key: "foreignField", Value: opts.ForeignField},
		{Key: "as", Value: opts.As},
	}
	return p.addStage("$lookup", doc)
}

// LookupPipelineOpts configures a $lookup with a sub-pipeline.
type LookupPipelineOpts struct {
	From     string
	Let      bson.D   // variables to pass to the pipeline
	Pipeline Pipeline // sub-pipeline
	As       string
}

// LookupPipeline performs a join with a sub-pipeline for more complex conditions.
//
// MongoDB equivalent:
//
//	{ $lookup: { from: "coll", let: { ... }, pipeline: [ ... ], as: "output" } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/lookup/
//
// Example:
//
//	subPipeline := gmqb.NewPipeline().
//	    Match(gmqb.Expr(bson.D{{"$gt", bson.A{"$amount", "$$minAmount"}}}))
//	p := gmqb.NewPipeline().LookupPipeline(gmqb.LookupPipelineOpts{
//	    From: "orders",
//	    Let:  bson.D{{"minAmount", "$minOrderAmount"}},
//	    Pipeline: subPipeline,
//	    As:   "qualifiedOrders",
//	})
func (p Pipeline) LookupPipeline(opts LookupPipelineOpts) Pipeline {
	doc := bson.D{
		{Key: "from", Value: opts.From},
	}
	if len(opts.Let) > 0 {
		doc = append(doc, bson.E{Key: "let", Value: opts.Let})
	}
	doc = append(doc, bson.E{Key: "pipeline", Value: opts.Pipeline.stages})
	doc = append(doc, bson.E{Key: "as", Value: opts.As})
	return p.addStage("$lookup", doc)
}

// AddFields adds new fields to documents. Alias: SetFields.
//
// MongoDB equivalent:
//
//	{ $addFields: { field1: <expression>, field2: <expression> } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/addFields/
//
// Example:
//
//	p := gmqb.NewPipeline().AddFields(gmqb.AddFieldsSpec(
//	    gmqb.AddField("fullName", gmqb.ExprConcat("$firstName", " ", "$lastName")),
//	    gmqb.AddField("isAdult", gmqb.ExprGte("$age", 18)),
//	))
func (p Pipeline) AddFields(fields bson.D) Pipeline {
	return p.addStage("$addFields", fields)
}

// AddField creates a field-expression pair for use in an $addFields or $set stage.
//
// Example:
//
//	gmqb.AddField("isAdult", gmqb.ExprGte("$age", 18))
func AddField(field string, expr interface{}) bson.E {
	return bson.E{Key: field, Value: expr}
}

// AddFieldsSpec builds a specification for the $addFields or $set stage.
//
// Example:
//
//	spec := gmqb.AddFieldsSpec(
//	    gmqb.AddField("fullName", gmqb.ExprConcat("$firstName", " ", "$lastName")),
//	    gmqb.AddField("isAdult", gmqb.ExprGte("$age", 18)),
//	)
func AddFieldsSpec(fields ...bson.E) bson.D {
	d := make(bson.D, len(fields))
	copy(d, fields)
	return d
}

// SetFields is an alias for AddFields. Corresponds to the $set aggregation stage.
// Use the AddFieldsSpec and AddField helpers to construct its specification.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/set/
//
// Example:
//
//	p := gmqb.NewPipeline().SetFields(gmqb.AddFieldsSpec(
//	    gmqb.AddField("status", "active"),
//	))
func (p Pipeline) SetFields(fields bson.D) Pipeline {
	return p.addStage("$set", fields)
}

// Unset removes fields from documents.
//
// MongoDB equivalent:
//
//	{ $unset: ["field1", "field2"] }   // multiple
//	{ $unset: "field" }                // single
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/unset/
//
// Example:
//
//	p := gmqb.NewPipeline().Unset("password", "ssn")
func (p Pipeline) Unset(fields ...string) Pipeline {
	if len(fields) == 1 {
		return p.addStage("$unset", fields[0])
	}
	return p.addStage("$unset", fields)
}

// Count inserts a document with a count of the number of documents at this stage.
//
// MongoDB equivalent:
//
//	{ $count: "fieldName" }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/count/
//
// Example:
//
//	p := gmqb.NewPipeline().Match(gmqb.Gt("age", 18)).Count("adultCount")
func (p Pipeline) Count(field string) Pipeline {
	return p.addStage("$count", field)
}

// Facet processes multiple aggregation pipelines within a single stage.
// Each key in the map is a facet name, and each value is a Pipeline.
//
// MongoDB equivalent:
//
//	{ $facet: { facet1: [ stage1, ... ], facet2: [ stage1, ... ] } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/facet/
//
// Example:
//
//	// Count users by age and country in separate facets
//	p := gmqb.NewPipeline().Facet(map[string]gmqb.Pipeline{
//	    "byAge": gmqb.NewPipeline().Group(gmqb.GroupSpec("$ageRange", gmqb.GroupAcc("count", gmqb.AccSum(1)))),
//	    "byCountry": gmqb.NewPipeline().Group(gmqb.GroupSpec("$country", gmqb.GroupAcc("count", gmqb.AccSum(1)))),
//	})
func (p Pipeline) Facet(facets map[string]Pipeline) Pipeline {
	doc := make(bson.D, 0, len(facets))
	for name, sub := range facets {
		doc = append(doc, bson.E{Key: name, Value: sub.stages})
	}
	return p.addStage("$facet", doc)
}

// BucketOpts configures the $bucket aggregation stage.
type BucketOpts struct {
	GroupBy    interface{}   // expression to group by
	Boundaries []interface{} // array of boundary values
	Default    interface{}   // value for documents outside boundaries
	Output     bson.D        // output document specification
}

// Bucket categorizes documents into groups (buckets) based on boundaries.
//
// MongoDB equivalent:
//
//	{ $bucket: { groupBy: <expr>, boundaries: [...], default: <val>, output: { ... } } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/bucket/
//
// Example:
//
//	p := gmqb.NewPipeline().Bucket(gmqb.BucketOpts{
//	    GroupBy:    "$age",
//	    Boundaries: []interface{}{0, 18, 30, 50, 100},
//	    Default:    "other",
//	    Output:     bson.D{{"count", bson.D{{"$sum", 1}}}},
//	})
func (p Pipeline) Bucket(opts BucketOpts) Pipeline {
	doc := bson.D{
		{Key: "groupBy", Value: opts.GroupBy},
		{Key: "boundaries", Value: opts.Boundaries},
	}
	if opts.Default != nil {
		doc = append(doc, bson.E{Key: "default", Value: opts.Default})
	}
	if len(opts.Output) > 0 {
		doc = append(doc, bson.E{Key: "output", Value: opts.Output})
	}
	return p.addStage("$bucket", doc)
}

// BucketAutoOpts configures the $bucketAuto aggregation stage.
type BucketAutoOpts struct {
	GroupBy     interface{} // expression to group by
	Buckets     int         // number of buckets
	Output      bson.D      // output document specification
	Granularity string      // preferred number series (e.g. "R5", "R10", "1-2-5")
}

// BucketAuto distributes documents into auto-determined buckets.
//
// MongoDB equivalent:
//
//	{ $bucketAuto: { groupBy: <expr>, buckets: n, output: { ... }, granularity: "R5" } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/bucketAuto/
//
// Example:
//
//	p := gmqb.NewPipeline().BucketAuto(gmqb.BucketAutoOpts{
//	    GroupBy: "$price",
//	    Buckets: 5,
//	})
func (p Pipeline) BucketAuto(opts BucketAutoOpts) Pipeline {
	doc := bson.D{
		{Key: "groupBy", Value: opts.GroupBy},
		{Key: "buckets", Value: opts.Buckets},
	}
	if len(opts.Output) > 0 {
		doc = append(doc, bson.E{Key: "output", Value: opts.Output})
	}
	if opts.Granularity != "" {
		doc = append(doc, bson.E{Key: "granularity", Value: opts.Granularity})
	}
	return p.addStage("$bucketAuto", doc)
}

// ReplaceRoot replaces each document with the specified embedded document.
//
// MongoDB equivalent:
//
//	{ $replaceRoot: { newRoot: <expression> } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/replaceRoot/
//
// Example:
//
//	p := gmqb.NewPipeline().ReplaceRoot("$address")
func (p Pipeline) ReplaceRoot(newRoot interface{}) Pipeline {
	return p.addStage("$replaceRoot", bson.D{{Key: "newRoot", Value: newRoot}})
}

// ReplaceWith is an alias for ReplaceRoot.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/replaceWith/
func (p Pipeline) ReplaceWith(newRoot interface{}) Pipeline {
	return p.addStage("$replaceWith", newRoot)
}

// Redact restricts document content based on information in the documents themselves.
//
// MongoDB equivalent:
//
//	{ $redact: <expression> }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/redact/
//
// Example:
//
//	p := gmqb.NewPipeline().Redact(bson.D{
//	    {"$cond", bson.D{
//	        {"if", bson.D{{"$eq", bson.A{"$level", "public"}}}},
//	        {"then", "$$DESCEND"},
//	        {"else", "$$PRUNE"},
//	    }},
//	})
func (p Pipeline) Redact(expression interface{}) Pipeline {
	return p.addStage("$redact", expression)
}

// Sample randomly selects the specified number of documents.
//
// MongoDB equivalent:
//
//	{ $sample: { size: n } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/sample/
//
// Example:
//
//	p := gmqb.NewPipeline().Sample(5)
func (p Pipeline) Sample(size int64) Pipeline {
	return p.addStage("$sample", bson.D{{Key: "size", Value: size}})
}

// SortByCount groups documents by an expression and sorts by count descending.
//
// MongoDB equivalent:
//
//	{ $sortByCount: <expression> }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/sortByCount/
//
// Example:
//
//	p := gmqb.NewPipeline().SortByCount("$status")
func (p Pipeline) SortByCount(expression interface{}) Pipeline {
	return p.addStage("$sortByCount", expression)
}

// UnionWith combines the pipeline results from another collection.
//
// MongoDB equivalent:
//
//	{ $unionWith: { coll: "otherColl", pipeline: [...] } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/unionWith/
//
// Example:
//
//	p := gmqb.NewPipeline().UnionWith("archivedUsers", nil)
func (p Pipeline) UnionWith(coll string, subPipeline *Pipeline) Pipeline {
	doc := bson.D{{Key: "coll", Value: coll}}
	if subPipeline != nil && len(subPipeline.stages) > 0 {
		doc = append(doc, bson.E{Key: "pipeline", Value: subPipeline.stages})
	}
	return p.addStage("$unionWith", doc)
}

// Out writes the pipeline output to a collection. Must be the last stage.
//
// MongoDB equivalent:
//
//	{ $out: "collectionName" }
//	{ $out: { db: "dbName", coll: "collName" } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/out/
//
// Example:
//
//	p := gmqb.NewPipeline().Match(gmqb.Eq("active", true)).Out("activeUsers")
func (p Pipeline) Out(collection string) Pipeline {
	return p.addStage("$out", collection)
}

// OutToDb writes the pipeline output to a collection in a specific database. Must be
// the last stage.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/out/
func (p Pipeline) OutToDb(db, collection string) Pipeline {
	return p.addStage("$out", bson.D{{Key: "db", Value: db}, {Key: "coll", Value: collection}})
}

// MergeOpts configures the $merge aggregation stage.
type MergeOpts struct {
	Into           interface{} // string or bson.D{db, coll}
	On             interface{} // field or array of fields
	Let            bson.D      // variables
	WhenMatched    interface{} // "replace", "keepExisting", "merge", "fail", or pipeline
	WhenNotMatched string      // "insert" or "discard" or "fail"
}

// Merge writes pipeline output into a collection with merge behavior. Must be the last stage.
//
// MongoDB equivalent:
//
//	{ $merge: { into: "coll", on: "_id", whenMatched: "merge", whenNotMatched: "insert" } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/merge/
//
// Example:
//
//	p := gmqb.NewPipeline().Merge(gmqb.MergeOpts{
//	    Into:           "monthlyTotals",
//	    On:             "_id",
//	    WhenMatched:    "merge",
//	    WhenNotMatched: "insert",
//	})
func (p Pipeline) Merge(opts MergeOpts) Pipeline {
	doc := bson.D{{Key: "into", Value: opts.Into}}
	if opts.On != nil {
		doc = append(doc, bson.E{Key: "on", Value: opts.On})
	}
	if len(opts.Let) > 0 {
		doc = append(doc, bson.E{Key: "let", Value: opts.Let})
	}
	if opts.WhenMatched != nil {
		doc = append(doc, bson.E{Key: "whenMatched", Value: opts.WhenMatched})
	}
	if opts.WhenNotMatched != "" {
		doc = append(doc, bson.E{Key: "whenNotMatched", Value: opts.WhenNotMatched})
	}
	return p.addStage("$merge", doc)
}

// GraphLookupOpts configures the $graphLookup stage.
type GraphLookupOpts struct {
	From                    string
	StartWith               interface{}
	ConnectFromField        string
	ConnectToField          string
	As                      string
	MaxDepth                *int
	DepthField              string
	RestrictSearchWithMatch Filter
}

// GraphLookup performs a recursive search on a collection.
//
// MongoDB equivalent:
//
//	{ $graphLookup: { from: "coll", startWith: "$field", connectFromField: "f1",
//	  connectToField: "f2", as: "output", maxDepth: n, depthField: "depth" } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/graphLookup/
//
// Example:
//
//	p := gmqb.NewPipeline().GraphLookup(gmqb.GraphLookupOpts{
//	    From:             "employees",
//	    StartWith:        "$reportsTo",
//	    ConnectFromField: "reportsTo",
//	    ConnectToField:   "name",
//	    As:               "reportingHierarchy",
//	})
func (p Pipeline) GraphLookup(opts GraphLookupOpts) Pipeline {
	doc := bson.D{
		{Key: "from", Value: opts.From},
		{Key: "startWith", Value: opts.StartWith},
		{Key: "connectFromField", Value: opts.ConnectFromField},
		{Key: "connectToField", Value: opts.ConnectToField},
		{Key: "as", Value: opts.As},
	}
	if opts.MaxDepth != nil {
		doc = append(doc, bson.E{Key: "maxDepth", Value: *opts.MaxDepth})
	}
	if opts.DepthField != "" {
		doc = append(doc, bson.E{Key: "depthField", Value: opts.DepthField})
	}
	if !opts.RestrictSearchWithMatch.IsEmpty() {
		doc = append(doc, bson.E{Key: "restrictSearchWithMatch", Value: opts.RestrictSearchWithMatch.d})
	}
	return p.addStage("$graphLookup", doc)
}

// GeoNearOpts configures the $geoNear aggregation stage.
type GeoNearOpts struct {
	Near          interface{} // GeoJSON point or legacy coordinates
	DistanceField string
	Spherical     bool
	MaxDistance   *float64
	MinDistance   *float64
	Query         Filter
	IncludeLocs   string
	Key           string
}

// GeoNear returns documents sorted by proximity to a geospatial point.
// Must be the first stage in a pipeline.
//
// MongoDB equivalent:
//
//	{ $geoNear: { near: point, distanceField: "dist", spherical: true, ... } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/geoNear/
//
// Example:
//
//	p := gmqb.NewPipeline().GeoNear(gmqb.GeoNearOpts{
//	    Near:          bson.D{{"type", "Point"}, {"coordinates", bson.A{-73.99, 40.73}}},
//	    DistanceField: "distance",
//	    Spherical:     true,
//	})
func (p Pipeline) GeoNear(opts GeoNearOpts) Pipeline {
	doc := bson.D{
		{Key: "near", Value: opts.Near},
		{Key: "distanceField", Value: opts.DistanceField},
		{Key: "spherical", Value: opts.Spherical},
	}
	if opts.MaxDistance != nil {
		doc = append(doc, bson.E{Key: "maxDistance", Value: *opts.MaxDistance})
	}
	if opts.MinDistance != nil {
		doc = append(doc, bson.E{Key: "minDistance", Value: *opts.MinDistance})
	}
	if !opts.Query.IsEmpty() {
		doc = append(doc, bson.E{Key: "query", Value: opts.Query.d})
	}
	if opts.IncludeLocs != "" {
		doc = append(doc, bson.E{Key: "includeLocs", Value: opts.IncludeLocs})
	}
	if opts.Key != "" {
		doc = append(doc, bson.E{Key: "key", Value: opts.Key})
	}
	return p.addStage("$geoNear", doc)
}

// Fill populates null and missing field values within documents.
//
// MongoDB equivalent:
//
//	{ $fill: { output: { field: { method: "linear" } } } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/fill/
//
// Example:
//
//	p := gmqb.NewPipeline().Fill(gmqb.FillSpec(
//	    gmqb.FillOutput("score", gmqb.FillMethod("linear")),
//	    gmqb.FillOutput("grade", gmqb.FillValue("Incomplete")),
//	))
func (p Pipeline) Fill(spec bson.D) Pipeline {
	return p.addStage("$fill", spec)
}

// FillOutput creates an output field specification for the $fill stage.
//
// Example:
//
//	gmqb.FillOutput("score", gmqb.FillMethod("linear"))
//	gmqb.FillOutput("grade", gmqb.FillValue("Incomplete"))
func FillOutput(field string, spec bson.E) bson.E {
	return bson.E{Key: field, Value: bson.D{spec}}
}

// FillMethod specifies a fill method for FillOutput ("linear" or "locf").
func FillMethod(method string) bson.E {
	return bson.E{Key: "method", Value: method}
}

// FillValue specifies a static fill value for FillOutput.
func FillValue(value interface{}) bson.E {
	return bson.E{Key: "value", Value: value}
}

// FillSpec builds a specification for the $fill stage.
// For more complex usage involving partitions or sorting, construct a raw bson.D.
//
// Example:
//
//	spec := gmqb.FillSpec(
//	    gmqb.FillOutput("score", gmqb.FillMethod("linear")),
//	)
func FillSpec(outputs ...bson.E) bson.D {
	d := make(bson.D, len(outputs))
	copy(d, outputs)
	return bson.D{{Key: "output", Value: d}}
}

// Densify creates new documents in a sequence where values are missing.
//
// MongoDB equivalent:
//
//	{ $densify: { field: "timestamp", range: { step: 1, unit: "hour", bounds: "full" } } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/densify/
//
// Example:
//
//	p := gmqb.NewPipeline().Densify(
//	    gmqb.DensifySpec("timestamp", gmqb.DensifyRange(1, "hour", "full")),
//	)
func (p Pipeline) Densify(spec bson.D) Pipeline {
	return p.addStage("$densify", spec)
}

// DensifyRange specifies the range of documents to create in a $densify stage.
// For bounds, use "full", "partition", or an array of [lower, upper].
//
// Example:
//
//	gmqb.DensifyRange(1, "hour", "full")
func DensifyRange(step interface{}, unit string, bounds interface{}) bson.D {
	d := bson.D{{Key: "step", Value: step}}
	if unit != "" {
		d = append(d, bson.E{Key: "unit", Value: unit})
	}
	d = append(d, bson.E{Key: "bounds", Value: bounds})
	return d
}

// DensifySpec builds a specification for the $densify stage.
//
// Example:
//
//	spec := gmqb.DensifySpec("timestamp", gmqb.DensifyRange(1, "hour", "full"))
func DensifySpec(field string, rangeSpec bson.D) bson.D {
	return bson.D{
		{Key: "field", Value: field},
		{Key: "range", Value: rangeSpec},
	}
}

// SetWindowFields groups documents and applies window operators.
//
// MongoDB equivalent:
//
//	{ $setWindowFields: { partitionBy: "$field", sortBy: { ... }, output: { ... } } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/setWindowFields/
//
// Example:
//
//	p := gmqb.NewPipeline().SetWindowFields(gmqb.SetWindowFieldsSpec(
//	    "$state",
//	    gmqb.SortSpec(gmqb.SortRule("orderDate", 1)),
//	    gmqb.WindowOutput("cumulativeQuantity", gmqb.AccSum("$quantity"), gmqb.Window("documents", "unbounded", "current")),
//	))
func (p Pipeline) SetWindowFields(spec bson.D) Pipeline {
	return p.addStage("$setWindowFields", spec)
}

// WindowOutput creates an output field specification for the $setWindowFields stage.
//
// Example:
//
//	gmqb.WindowOutput("cumulativeQuantity", gmqb.AccSum("$quantity"), gmqb.Window("documents", "unbounded", "current"))
func WindowOutput(field string, expr interface{}, window bson.E) bson.E {
	var outDoc bson.D
	if extD, ok := expr.(bson.D); ok {
		outDoc = make(bson.D, len(extD), len(extD)+1)
		copy(outDoc, extD)
		if window.Key != "" {
			outDoc = append(outDoc, window)
		}
	} else {
		// Fallback if the user didn't pass a proper accumulator
		outDoc = bson.D{{Key: "$expr", Value: expr}}
		if window.Key != "" {
			outDoc = append(outDoc, window)
		}
	}

	return bson.E{Key: field, Value: outDoc}
}

// Window creates a window specification for a WindowOutput accumulator.
// The boundsType must be "documents" or "range".
// Bounds can be numbers (e.g. -1, 1) or strings like "unbounded" or "current".
//
// Example:
//
//	gmqb.Window("documents", "unbounded", "current")
func Window(boundsType string, lowerBound, upperBound interface{}) bson.E {
	return bson.E{Key: "window", Value: bson.D{
		{Key: boundsType, Value: bson.A{lowerBound, upperBound}},
	}}
}

// SetWindowFieldsSpec builds a specification for the $setWindowFields stage.
//
// Example:
//
//	spec := gmqb.SetWindowFieldsSpec(
//	    "$state",
//	    gmqb.SortSpec(gmqb.SortRule("orderDate", 1)),
//	    gmqb.WindowOutput("cumulativeQuantity", gmqb.AccSum("$quantity"), gmqb.Window("documents", "unbounded", "current")),
//	)
func SetWindowFieldsSpec(partitionBy interface{}, sortBy bson.D, outputs ...bson.E) bson.D {
	d := make(bson.D, 0, 3)
	if partitionBy != nil && partitionBy != "" {
		d = append(d, bson.E{Key: "partitionBy", Value: partitionBy})
	}
	if len(sortBy) > 0 {
		d = append(d, bson.E{Key: "sortBy", Value: sortBy})
	}

	if len(outputs) > 0 {
		outDoc := make(bson.D, len(outputs))
		copy(outDoc, outputs)
		d = append(d, bson.E{Key: "output", Value: outDoc})
	}

	return d
}

// RawStage appends a raw aggregation stage. Use for stages not yet supported by
// the builder, or for custom/Atlas-specific stages.
//
// Example:
//
//	p := gmqb.NewPipeline().RawStage("$search", bson.D{
//	    {"text", bson.D{{"query", "coffee"}, {"path", "description"}}},
//	})
func (p Pipeline) RawStage(name string, value interface{}) Pipeline {
	return p.addStage(name, value)
}
