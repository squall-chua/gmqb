package gmqb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestPipeline_Coverage(t *testing.T) {
	p := NewPipeline()

	// These tests just verify the builder methods return a pipeline with the newly appended stage key

	tests := []struct {
		name      string
		build     func() Pipeline
		expectKey string
	}{
		{"MatchRaw", func() Pipeline { return p.MatchRaw(bson.D{{Key: "x", Value: 1}}) }, "$match"},
		{"UnwindWithOpts", func() Pipeline {
			return p.UnwindWithOpts(UnwindOpts{Path: "path", IncludeArrayIndex: "i", PreserveNullAndEmptyArrays: true})
		}, "$unwind"},
		{"LookupPipeline", func() Pipeline {
			return p.LookupPipeline(LookupPipelineOpts{From: "from", Let: []bson.E{{Key: "var1", Value: "$v1"}}, Pipeline: NewPipeline(), As: "as"})
		}, "$lookup"},
		{"SetFields", func() Pipeline { return p.SetFields(bson.D{{Key: "x", Value: 1}}) }, "$set"},
		{"Unset", func() Pipeline { return p.Unset("a", "b") }, "$unset"},
		{"BucketAuto", func() Pipeline {
			return p.BucketAuto(BucketAutoOpts{GroupBy: "$group", Buckets: 5, Granularity: "R5", Output: bson.D{{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}}}})
		}, "$bucketAuto"},
		{"Bucket", func() Pipeline {
			return p.Bucket(BucketOpts{GroupBy: "$group", Boundaries: []interface{}{0, 10}, Default: "other"})
		}, "$bucket"},
		{"ReplaceWith", func() Pipeline { return p.ReplaceWith("$newRoot") }, "$replaceWith"},
		{"Redact", func() Pipeline { return p.Redact("$cond") }, "$redact"},
		{"UnionWith", func() Pipeline { return p.UnionWith("coll", nil) }, "$unionWith"},
		{"OutToDb", func() Pipeline { return p.OutToDb("db", "coll") }, "$out"},
		{"Merge", func() Pipeline { return p.Merge(MergeOpts{Into: "coll", On: "id"}) }, "$merge"},
		{"GraphLookup", func() Pipeline {
			return p.GraphLookup(GraphLookupOpts{From: "from", StartWith: "$start", ConnectFromField: "connFrom", ConnectToField: "connTo", As: "as"})
		}, "$graphLookup"},
		{"GeoNear", func() Pipeline {
			return p.GeoNear(GeoNearOpts{Near: "$near", DistanceField: "dist", Spherical: true})
		}, "$geoNear"},
		{"Fill", func() Pipeline { return p.Fill(FillSpec(FillOutput("x", FillMethod("linear")))) }, "$fill"},
		{"Densify", func() Pipeline { return p.Densify(DensifySpec("x", DensifyRange(1, "hour", "full"))) }, "$densify"},
		{"SetWindowFields", func() Pipeline { return p.SetWindowFields(SetWindowFieldsSpec("", nil)) }, "$setWindowFields"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.build().BsonD()
			assert.Equal(t, tc.expectKey, res[0][0].Key)
		})
	}

	// Test GroupID
	gid := GroupID("a", "b")
	assert.Equal(t, "a", gid[0].Key)
	assert.Equal(t, "$a", gid[0].Value)

	// Test FillValue
	fv := FillValue("val")
	assert.Equal(t, "value", fv.Key)

	// Test Window
	w := Window("documents", "unbounded", "current")
	assert.Equal(t, "window", w.Key)

	// CompactJSON
	j := p.MatchRaw(bson.D{{Key: "x", Value: 1}}).CompactJSON()
	assert.Contains(t, j, "$match")

	j2 := p.MatchRaw(bson.D{{Key: "x", Value: 1}}).JSON()
	assert.Contains(t, j2, "$match")
}
