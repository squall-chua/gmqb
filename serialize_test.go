package gmqb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestToJSON(t *testing.T) {
	d := bson.D{{Key: "name", Value: "Alice"}, {Key: "age", Value: 30}}
	got := toJSON(d)
	assert.Contains(t, got, `"name": "Alice"`)
	assert.Contains(t, got, `"age": 30`)
}

func TestToCompactJSON(t *testing.T) {
	d := bson.D{{Key: "name", Value: "Alice"}}
	got := toCompactJSON(d)
	assert.Contains(t, got, `"name":"Alice"`)
	assert.NotContains(t, got, "\n")
}

func TestPipelineToJSON(t *testing.T) {
	stages := []bson.D{
		{{Key: "$match", Value: bson.D{{Key: "age", Value: 30}}}},
	}
	got := pipelineToJSON(stages)
	assert.Contains(t, got, "$match")
}

func TestPipelineToCompactJSON(t *testing.T) {
	stages := []bson.D{
		{{Key: "$match", Value: bson.D{{Key: "age", Value: 30}}}},
	}
	got := pipelineToCompactJSON(stages)
	assert.Contains(t, got, "$match")
	assert.NotContains(t, got, "\n")
}

func TestToJSON_EmptyDoc(t *testing.T) {
	got := toJSON(bson.D{})
	assert.Equal(t, "{}", got)
}
