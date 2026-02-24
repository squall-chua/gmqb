package gmqb

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestNewPipeline_Empty(t *testing.T) {
	assert.True(t, NewPipeline().IsEmpty())
}

func TestPipeline_Match(t *testing.T) {
	stages := NewPipeline().Match(Eq("status", "active")).BsonD()
	require.Len(t, stages, 1)
	assert.Equal(t, "$match", stages[0][0].Key)
}

func TestPipeline_Project(t *testing.T) {
	stages := NewPipeline().Project(bson.D{{"name", 1}, {"_id", 0}}).BsonD()
	require.Len(t, stages, 1)
	assert.Equal(t, "$project", stages[0][0].Key)
}

func TestPipeline_Group(t *testing.T) {
	stages := NewPipeline().Group(GroupSpec(
		"$country",
		GroupAcc("total", AccSum(1)),
	)).BsonD()
	assert.Equal(t, "$group", stages[0][0].Key)
}

func TestPipeline_Sort(t *testing.T) {
	stages := NewPipeline().Sort(Desc("age")).BsonD()
	assert.Equal(t, "$sort", stages[0][0].Key)
}

func TestPipeline_LimitSkip(t *testing.T) {
	stages := NewPipeline().Skip(20).Limit(10).BsonD()
	require.Len(t, stages, 2)
	assert.Equal(t, "$skip", stages[0][0].Key)
	assert.Equal(t, "$limit", stages[1][0].Key)
}

func TestPipeline_Unwind(t *testing.T) {
	stages := NewPipeline().Unwind("$tags").BsonD()
	assert.Equal(t, "$unwind", stages[0][0].Key)
}

func TestPipeline_Lookup(t *testing.T) {
	stages := NewPipeline().Lookup(LookupOpts{
		From:         "orders",
		LocalField:   "userId",
		ForeignField: "_id",
		As:           "userOrders",
	}).BsonD()
	assert.Equal(t, "$lookup", stages[0][0].Key)
}

func TestPipeline_AddFields(t *testing.T) {
	stages := NewPipeline().AddFields(AddFieldsSpec(
		AddField("isAdult", bson.D{{"$gte", bson.A{"$age", 18}}}),
	)).BsonD()
	assert.Equal(t, "$addFields", stages[0][0].Key)
}

func TestPipeline_Unset(t *testing.T) {
	stages := NewPipeline().Unset("password", "ssn").BsonD()
	assert.Equal(t, "$unset", stages[0][0].Key)
}

func TestPipeline_Count(t *testing.T) {
	stages := NewPipeline().Count("total").BsonD()
	assert.Equal(t, "$count", stages[0][0].Key)
}

func TestPipeline_Sample(t *testing.T) {
	stages := NewPipeline().Sample(5).BsonD()
	assert.Equal(t, "$sample", stages[0][0].Key)
}

func TestPipeline_SortByCount(t *testing.T) {
	stages := NewPipeline().SortByCount("$status").BsonD()
	assert.Equal(t, "$sortByCount", stages[0][0].Key)
}

func TestPipeline_Out(t *testing.T) {
	stages := NewPipeline().Out("archive").BsonD()
	assert.Equal(t, "$out", stages[0][0].Key)
}

func TestPipeline_ReplaceRoot(t *testing.T) {
	stages := NewPipeline().ReplaceRoot("$address").BsonD()
	assert.Equal(t, "$replaceRoot", stages[0][0].Key)
}

func TestPipeline_RawStage(t *testing.T) {
	stages := NewPipeline().RawStage("$search", bson.D{{"text", bson.D{{"query", "test"}}}}).BsonD()
	assert.Equal(t, "$search", stages[0][0].Key)
}

func TestPipeline_Immutable(t *testing.T) {
	p1 := NewPipeline().Match(Eq("a", 1))
	p2 := p1.Limit(10)
	assert.Len(t, p1.BsonD(), 1)
	assert.Len(t, p2.BsonD(), 2)
}

func TestPipeline_MultiStage(t *testing.T) {
	p := NewPipeline().
		Match(Gte("age", 18)).
		Group(GroupSpec("$country", GroupAcc("count", AccSum(1)))).
		Sort(Desc("count")).
		Limit(10)

	stages := p.BsonD()
	require.Len(t, stages, 4)
	expectedOps := []string{"$match", "$group", "$sort", "$limit"}
	for i, expected := range expectedOps {
		assert.Equal(t, expected, stages[i][0].Key, "stage %d", i)
	}
}

func TestPipeline_JSON(t *testing.T) {
	p := NewPipeline().Match(Eq("a", 1)).Limit(5)
	j := p.JSON()
	var arr []interface{}
	require.NoError(t, json.Unmarshal([]byte(j), &arr))
	assert.Len(t, arr, 2)
}

func TestAsc(t *testing.T) {
	d := Asc("name")
	assert.Equal(t, "name", d[0].Key)
	assert.Equal(t, 1, d[0].Value)
}

func TestDesc(t *testing.T) {
	d := Desc("name")
	assert.Equal(t, "name", d[0].Key)
	assert.Equal(t, -1, d[0].Value)
}

func TestPipeline_Facet(t *testing.T) {
	stages := NewPipeline().Facet(map[string]Pipeline{
		"byAge": NewPipeline().Group(GroupSpec("$ageRange")),
		"total": NewPipeline().Count("count"),
	}).BsonD()
	assert.Equal(t, "$facet", stages[0][0].Key)
}

func TestPipeline_Bucket(t *testing.T) {
	stages := NewPipeline().Bucket(BucketOpts{
		GroupBy:    "$age",
		Boundaries: []interface{}{0, 18, 65, 100},
		Default:    "other",
	}).BsonD()
	assert.Equal(t, "$bucket", stages[0][0].Key)
}

func TestPipeline_UnionWith(t *testing.T) {
	stages := NewPipeline().UnionWith("archive", nil).BsonD()
	assert.Equal(t, "$unionWith", stages[0][0].Key)
}

func ExampleNewPipeline() {
	p := NewPipeline().
		Match(Gte("age", 18)).
		Group(bson.D{{"_id", "$country"}, {"count", bson.D{{"$sum", 1}}}}).
		Sort(Desc("count")).
		Limit(10)
	_ = p.BsonD()
}
