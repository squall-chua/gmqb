package gmqb

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// helper to compare filter JSON output (normalizes key ordering).
func assertFilterJSON(t *testing.T, f Filter, expected string) {
	t.Helper()
	got := f.CompactJSON()
	var gotMap, expectedMap interface{}
	require.NoError(t, json.Unmarshal([]byte(got), &gotMap), "invalid JSON from filter: %s", got)
	require.NoError(t, json.Unmarshal([]byte(expected), &expectedMap), "invalid expected JSON: %s", expected)
	gotBytes, _ := json.Marshal(gotMap)
	expectedBytes, _ := json.Marshal(expectedMap)
	assert.JSONEq(t, string(expectedBytes), string(gotBytes))
}

// --- Comparison Operators ---

func TestEq(t *testing.T) {
	assertFilterJSON(t, Eq("name", "Alice"), `{"name":{"$eq":"Alice"}}`)
}

func TestNe(t *testing.T) {
	assertFilterJSON(t, Ne("status", "archived"), `{"status":{"$ne":"archived"}}`)
}

func TestGt(t *testing.T) {
	assertFilterJSON(t, Gt("age", 18), `{"age":{"$gt":18}}`)
}

func TestGte(t *testing.T) {
	assertFilterJSON(t, Gte("age", 18), `{"age":{"$gte":18}}`)
}

func TestLt(t *testing.T) {
	assertFilterJSON(t, Lt("price", 100), `{"price":{"$lt":100}}`)
}

func TestLte(t *testing.T) {
	assertFilterJSON(t, Lte("qty", 50), `{"qty":{"$lte":50}}`)
}

func TestIn(t *testing.T) {
	assertFilterJSON(t, In("status", "active", "pending"), `{"status":{"$in":["active","pending"]}}`)
}

func TestNin(t *testing.T) {
	assertFilterJSON(t, Nin("role", "banned", "suspended"), `{"role":{"$nin":["banned","suspended"]}}`)
}

// --- Logical Operators ---

func TestAnd(t *testing.T) {
	f := And(Gte("age", 18), Lt("age", 65))
	assertFilterJSON(t, f, `{"$and":[{"age":{"$gte":18}},{"age":{"$lt":65}}]}`)
}

func TestOr(t *testing.T) {
	f := Or(Eq("status", "active"), Eq("status", "pending"))
	assertFilterJSON(t, f, `{"$or":[{"status":{"$eq":"active"}},{"status":{"$eq":"pending"}}]}`)
}

func TestNor(t *testing.T) {
	f := Nor(Eq("status", "archived"))
	assertFilterJSON(t, f, `{"$nor":[{"status":{"$eq":"archived"}}]}`)
}

func TestNot(t *testing.T) {
	f := Not("age", Gte("age", 18))
	assertFilterJSON(t, f, `{"age":{"$not":{"$gte":18}}}`)
}

// --- Element Operators ---

func TestExists(t *testing.T) {
	assertFilterJSON(t, Exists("email", true), `{"email":{"$exists":true}}`)
}

func TestType(t *testing.T) {
	assertFilterJSON(t, Type("age", "int"), `{"age":{"$type":"int"}}`)
}

// --- Evaluation Operators ---

func TestMod(t *testing.T) {
	assertFilterJSON(t, Mod("qty", 4, 0), `{"qty":{"$mod":[4,0]}}`)
}

func TestRegex(t *testing.T) {
	f := Regex("email", `^test`, "i")
	got := f.CompactJSON()
	var m map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(got), &m))
	emailVal := m["email"].(map[string]interface{})
	assert.Equal(t, "^test", emailVal["$regex"])
	assert.Equal(t, "i", emailVal["$options"])
}

func TestRegexNoOptions(t *testing.T) {
	f := Regex("name", "^A", "")
	got := f.CompactJSON()
	var m map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(got), &m))
	nameVal := m["name"].(map[string]interface{})
	assert.NotContains(t, nameVal, "$options")
}

func TestWhere(t *testing.T) {
	assertFilterJSON(t, Where("this.a > this.b"), `{"$where":"this.a > this.b"}`)
}

func TestExpr(t *testing.T) {
	f := Expr(bson.D{{"$gt", bson.A{"$spent", "$budget"}}})
	assert.NotEqual(t, "{}", f.CompactJSON())
}

// --- Array Operators ---

func TestAll(t *testing.T) {
	assertFilterJSON(t, All("tags", "ssl", "security"), `{"tags":{"$all":["ssl","security"]}}`)
}

func TestSize(t *testing.T) {
	assertFilterJSON(t, Size("tags", 3), `{"tags":{"$size":3}}`)
}

func TestElemMatch(t *testing.T) {
	inner := And(Gte("score", 80), Lt("score", 100))
	f := ElemMatch("results", inner)
	assert.NotEqual(t, "{}", f.CompactJSON())
}

// --- Geospatial ---

func TestGeoIntersects(t *testing.T) {
	geo := bson.D{{"type", "Point"}, {"coordinates", bson.A{1.0, 2.0}}}
	assert.False(t, GeoIntersects("location", geo).IsEmpty())
}

func TestNear(t *testing.T) {
	geo := bson.D{{"type", "Point"}, {"coordinates", bson.A{-73.9667, 40.78}}}
	assert.False(t, Near("location", geo, 1000, 0).IsEmpty())
}

// --- Bitwise ---

func TestBitsAllClear(t *testing.T) {
	assertFilterJSON(t, BitsAllClear("flags", 35), `{"flags":{"$bitsAllClear":35}}`)
}

func TestBitsAllSet(t *testing.T) {
	assertFilterJSON(t, BitsAllSet("flags", 50), `{"flags":{"$bitsAllSet":50}}`)
}

// --- Output Methods ---

func TestFilter_BsonM(t *testing.T) {
	m := Eq("name", "Alice").BsonM()
	assert.Contains(t, m, "name")
}

func TestFilter_IsEmpty(t *testing.T) {
	assert.True(t, Filter{}.IsEmpty())
	assert.False(t, Eq("a", 1).IsEmpty())
}

func TestFilter_JSON(t *testing.T) {
	assert.NotEqual(t, "{}", Eq("name", "test").JSON())
}

func TestRaw(t *testing.T) {
	f := Raw(bson.D{{"$text", bson.D{{"$search", "coffee"}}}})
	assert.False(t, f.IsEmpty())
}

// --- Chaining Tests ---

func TestNewFilter_Empty(t *testing.T) {
	assert.True(t, NewFilter().IsEmpty())
}

func TestFilter_Chain_Eq(t *testing.T) {
	f := NewFilter().Eq("name", "Alice")
	assertFilterJSON(t, f, `{"name":{"$eq":"Alice"}}`)
}

func TestFilter_Chain_Multiple(t *testing.T) {
	f := NewFilter().
		Eq("status", "active").
		Gte("age", 18).
		Lt("age", 65)
	d := f.BsonD()
	require.Len(t, d, 3)
	assert.Equal(t, "status", d[0].Key)
	assert.Equal(t, "age", d[1].Key)
	assert.Equal(t, "age", d[2].Key)
}

func TestFilter_Chain_AllOperators(t *testing.T) {
	f := NewFilter().
		Ne("status", "archived").
		Gt("score", 50).
		Lte("price", 100).
		In("country", "US", "UK").
		Nin("role", "banned").
		Exists("email", true).
		Type("age", "int").
		Size("tags", 2).
		Regex("name", "^A", "i")
	d := f.BsonD()
	assert.Len(t, d, 9, "should have 9 chained conditions")
}

func TestFilter_Chain_Immutable(t *testing.T) {
	f1 := NewFilter().Eq("a", 1)
	f2 := f1.Eq("b", 2)
	assert.Len(t, f1.BsonD(), 1, "original should be unchanged")
	assert.Len(t, f2.BsonD(), 2, "chained should have both")
}

// --- Testable Examples ---

func ExampleEq() {
	f := Eq("age", 21)
	_ = f.BsonD()
}

func ExampleAnd() {
	f := And(Gte("age", 18), Lt("age", 65))
	_ = f.BsonD()
}

func ExampleOr() {
	f := Or(Eq("status", "active"), Eq("status", "pending"))
	_ = f.BsonD()
}

func TestFilter_Not_NoMatch(t *testing.T) {
	// If the inner filter is an operator but not for identical field, Not wraps everything
	f := Not("status", Eq("other", "value"))
	d := f.BsonD()
	assert.Equal(t, "status", d[0].Key)
	assert.Equal(t, "$not", d[0].Value.(bson.D)[0].Key)
}

func TestFilter_JsonSchema(t *testing.T) {
	f := JsonSchema(bson.D{{"type", "object"}})
	assert.Equal(t, "$jsonSchema", f.BsonD()[0].Key)
}

func TestFilter_GeoWithin(t *testing.T) {
	f := GeoWithin("location", Polygon([][2]float64{{0, 0}, {3, 6}, {6, 1}, {0, 0}}))
	assert.Equal(t, "location", f.BsonD()[0].Key)
	assert.Equal(t, "$geoWithin", f.BsonD()[0].Value.(bson.D)[0].Key)
}

func TestFilter_Near_MinDistance(t *testing.T) {
	f := Near("location", Point(10, 20), 1000, 100)
	assert.Equal(t, "location", f.BsonD()[0].Key)
	opts := f.BsonD()[0].Value.(bson.D)[0].Value.(bson.D)
	assert.Equal(t, "$minDistance", opts[2].Key)
	assert.Equal(t, float64(100), opts[2].Value)
}

func TestFilter_NearSphere(t *testing.T) {
	f := NearSphere("location", Point(10, 20), 1000, 100)
	assert.Equal(t, "location", f.BsonD()[0].Key)
	assert.Equal(t, "$nearSphere", f.BsonD()[0].Value.(bson.D)[0].Key)
}

func TestFilter_BitsAnyClear(t *testing.T) {
	f := BitsAnyClear("permissions", 4)
	assert.Equal(t, "permissions", f.BsonD()[0].Key)
	assert.Equal(t, "$bitsAnyClear", f.BsonD()[0].Value.(bson.D)[0].Key)
}

func TestFilter_BitsAnySet(t *testing.T) {
	f := BitsAnySet("permissions", []int{1, 5})
	assert.Equal(t, "permissions", f.BsonD()[0].Key)
	assert.Equal(t, "$bitsAnySet", f.BsonD()[0].Value.(bson.D)[0].Key)
}
