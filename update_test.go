package gmqb

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func assertUpdateJSON(t *testing.T, u Updater, expected string) {
	t.Helper()
	got := u.CompactJSON()
	var gotMap, expectedMap interface{}
	require.NoError(t, json.Unmarshal([]byte(got), &gotMap), "invalid JSON from update: %s", got)
	require.NoError(t, json.Unmarshal([]byte(expected), &expectedMap), "invalid expected JSON: %s", expected)
	assert.JSONEq(t, expected, got)
}

func TestNewUpdate_Empty(t *testing.T) {
	assert.True(t, NewUpdate().IsEmpty())
}

func TestUpdate_Set(t *testing.T) {
	assertUpdateJSON(t, NewUpdate().Set("name", "Alice"), `{"$set":{"name":"Alice"}}`)
}

func TestUpdate_SetMultiple(t *testing.T) {
	assertUpdateJSON(t, NewUpdate().Set("name", "Alice").Set("age", 30), `{"$set":{"name":"Alice","age":30}}`)
}

func TestUpdate_Unset(t *testing.T) {
	assertUpdateJSON(t, NewUpdate().Unset("oldField"), `{"$unset":{"oldField":""}}`)
}

func TestUpdate_Inc(t *testing.T) {
	assertUpdateJSON(t, NewUpdate().Inc("views", 1), `{"$inc":{"views":1}}`)
}

func TestUpdate_Mul(t *testing.T) {
	assertUpdateJSON(t, NewUpdate().Mul("price", 1.1), `{"$mul":{"price":1.1}}`)
}

func TestUpdate_Min(t *testing.T) {
	assertUpdateJSON(t, NewUpdate().Min("lowScore", 50), `{"$min":{"lowScore":50}}`)
}

func TestUpdate_Max(t *testing.T) {
	assertUpdateJSON(t, NewUpdate().Max("highScore", 950), `{"$max":{"highScore":950}}`)
}

func TestUpdate_Rename(t *testing.T) {
	assertUpdateJSON(t, NewUpdate().Rename("nmae", "name"), `{"$rename":{"nmae":"name"}}`)
}

func TestUpdate_CurrentDate(t *testing.T) {
	assertUpdateJSON(t, NewUpdate().CurrentDate("lastModified"), `{"$currentDate":{"lastModified":true}}`)
}

func TestUpdate_SetOnInsert(t *testing.T) {
	assertUpdateJSON(t, NewUpdate().SetOnInsert("created", "now"), `{"$setOnInsert":{"created":"now"}}`)
}

func TestUpdate_Push(t *testing.T) {
	assertUpdateJSON(t, NewUpdate().Push("scores", 95), `{"$push":{"scores":95}}`)
}

func TestUpdate_AddToSet(t *testing.T) {
	assertUpdateJSON(t, NewUpdate().AddToSet("tags", "new"), `{"$addToSet":{"tags":"new"}}`)
}

func TestUpdate_AddToSetEach(t *testing.T) {
	assertUpdateJSON(t, NewUpdate().AddToSetEach("tags", "a", "b", "c"), `{"$addToSet":{"tags":{"$each":["a","b","c"]}}}`)
}

func TestUpdate_Pop(t *testing.T) {
	assertUpdateJSON(t, NewUpdate().Pop("arr", 1), `{"$pop":{"arr":1}}`)
}

func TestUpdate_Pull(t *testing.T) {
	assertUpdateJSON(t, NewUpdate().Pull("tags", "obsolete"), `{"$pull":{"tags":"obsolete"}}`)
}

func TestUpdate_PullAll(t *testing.T) {
	assertUpdateJSON(t, NewUpdate().PullAll("tags", "a", "b"), `{"$pullAll":{"tags":["a","b"]}}`)
}

func TestUpdate_BitAnd(t *testing.T) {
	assertUpdateJSON(t, NewUpdate().BitAnd("flags", 10), `{"$bit":{"flags":{"and":10}}}`)
}

func TestUpdate_BitOr(t *testing.T) {
	assertUpdateJSON(t, NewUpdate().BitOr("flags", 5), `{"$bit":{"flags":{"or":5}}}`)
}

func TestUpdate_BitXor(t *testing.T) {
	assertUpdateJSON(t, NewUpdate().BitXor("flags", 15), `{"$bit":{"flags":{"xor":15}}}`)
}

func TestUpdate_Chaining_Immutable(t *testing.T) {
	u1 := NewUpdate().Set("a", 1)
	u2 := u1.Set("b", 2)
	assert.NotEqual(t, u1.CompactJSON(), u2.CompactJSON(), "chaining should not mutate original")
}

func TestUpdate_MultiOperator(t *testing.T) {
	u := NewUpdate().Set("name", "Bob").Inc("age", 1).Push("tags", "new")
	j := u.CompactJSON()
	var m map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(j), &m))
	assert.Contains(t, m, "$set")
	assert.Contains(t, m, "$inc")
	assert.Contains(t, m, "$push")
}

func TestUpdate_PushWithOpts(t *testing.T) {
	pos := 0
	sl := -5
	u := NewUpdate().PushWithOpts("scores", PushOpts{
		Each:     []interface{}{89, 92},
		Position: &pos,
		Slice:    &sl,
	})
	j := u.CompactJSON()
	var m map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(j), &m))
	assert.Contains(t, m, "$push")
}

func ExampleNewUpdate() {
	u := NewUpdate().
		Set("name", "Alice").
		Inc("loginCount", 1)
	_ = u.BsonD()
}

func TestUpdate_CurrentDateAsTimestamp(t *testing.T) {
	u := NewUpdate().CurrentDateAsTimestamp("lastModified")
	d := u.BsonD()
	assert.Equal(t, "$currentDate", d[0].Key)
	assert.Equal(t, "lastModified", d[0].Value.(bson.D)[0].Key)
	assert.Equal(t, bson.D{{Key: "$type", Value: "timestamp"}}, d[0].Value.(bson.D)[0].Value)
}

func TestUpdate_PushWithOpts_Sort(t *testing.T) {
	u := NewUpdate().PushWithOpts("scores", PushOpts{
		Each: []interface{}{89, 92},
		Sort: bson.D{{Key: "score", Value: -1}},
	})
	d := u.BsonD()
	opts := d[0].Value.(bson.D)[0].Value.(bson.D)
	assert.Equal(t, "$sort", opts[1].Key)
}

func TestUpdate_JSON(t *testing.T) {
	u := NewUpdate().Set("status", "active")
	jsonStr := u.JSON()
	assert.Contains(t, jsonStr, `"status"`)
	assert.Contains(t, jsonStr, `"$set"`)
}
