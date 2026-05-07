package rest

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestBuildFilter(t *testing.T) {
	fields := []FilterField{
		{Name: "name", BsonKey: "name", Op: OpEq | OpIn},
		{Name: "age", BsonKey: "age", Op: OpEq | OpGte | OpLte | OpIn, ValueParser: func(s string) (any, error) {
			return strconv.Atoi(s)
		}},
		{Name: "status", BsonKey: "status", Op: OpEq},
	}

	tests := []struct {
		name     string
		query    map[string][]string
		expected bson.M
		wantErr  bool
	}{
		{
			name:  "bracket notation eq",
			query: map[string][]string{"status[eq]": {"active"}},
			expected: bson.M{"status": bson.D{{Key: "$eq", Value: "active"}}},
		},
		{
			name:  "bracket notation gte typed",
			query: map[string][]string{"age[gte]": {"30"}},
			expected: bson.M{"age": bson.D{{Key: "$gte", Value: 30}}},
		},
		{
			name:  "underscore notation lte typed",
			query: map[string][]string{"age_lte": {"25"}},
			expected: bson.M{"age": bson.D{{Key: "$lte", Value: 25}}},
		},
		{
			name:  "plain eq (no suffix)",
			query: map[string][]string{"name": {"alice"}},
			expected: bson.M{"name": bson.D{{Key: "$eq", Value: "alice"}}},
		},
		{
			name:  "in operator with comma split and parser",
			query: map[string][]string{"age[in]": {"25,30"}},
			expected: bson.M{"age": bson.D{{Key: "$in", Value: bson.A{25, 30}}}},
		},
		{
			name:    "unknown field",
			query:   map[string][]string{"unknown": {"val"}},
			wantErr: true,
		},
		{
			name:    "operator not allowed",
			query:   map[string][]string{"status[gte]": {"active"}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := BuildFilter(tt.query, fields)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			bsonData := filter.BsonM()
			assert.Equal(t, tt.expected, bsonData)
		})
	}
}
