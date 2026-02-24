package generator

import (
	"testing"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name     string
		jsonStr  string
		expected string
	}{
		{
			name:     "basic equality",
			jsonStr:  `{"name": "Alice"}`,
			expected: `gmqb.Eq("name", "Alice")`,
		},
		{
			name:     "multiple fields implicit and",
			jsonStr:  `{"name": "Alice", "age": 30}`,
			expected: "gmqb.And(\n\tgmqb.Eq(\"name\", \"Alice\"),\n\tgmqb.Eq(\"age\", 30),\n)",
		},
		{
			name:     "basic operator",
			jsonStr:  `{"age": {"$gte": 18}}`,
			expected: `gmqb.Gte("age", 18)`,
		},
		{
			name:     "logical operator",
			jsonStr:  `{"$or": [{"status": "A"}, {"age": 50}]}`,
			expected: "gmqb.Or(\n\tgmqb.Eq(\"status\", \"A\"),\n\tgmqb.Eq(\"age\", 50),\n)",
		},
		{
			name:     "in operator",
			jsonStr:  `{"status": {"$in": ["A", "B"]}}`,
			expected: `gmqb.In("status", "A", "B")`,
		},
		{
			name:     "elem match",
			jsonStr:  `{"results": {"$elemMatch": {"product": "xyz", "score": {"$gte": 8}}}}`,
			expected: "gmqb.ElemMatch(\"results\", gmqb.And(\n\tgmqb.Eq(\"product\", \"xyz\"),\n\tgmqb.Gte(\"score\", 8),\n))",
		},
		{
			name:     "size operator",
			jsonStr:  `{"tags": {"$size": 3}}`,
			expected: `gmqb.Size("tags", 3)`,
		},
		{
			name:     "basic pipeline",
			jsonStr:  `[{"$match": {"status": "A"}}, {"$limit": 10}]`,
			expected: "gmqb.NewPipeline().\n\tMatch(gmqb.Eq(\"status\", \"A\")).\n\tLimit(10)",
		},
		{
			name:     "pipeline with group and accumulators",
			jsonStr:  `[{"$group": {"_id": "$country", "total": {"$sum": "$amount"}, "avgAge": {"$avg": "$age"}}}]`,
			expected: "gmqb.NewPipeline().\n\tGroup(gmqb.GroupSpec(\"$country\",\n\t\tgmqb.GroupAcc(\"total\", gmqb.AccSum(\"$amount\")),\n\t\tgmqb.GroupAcc(\"avgAge\", gmqb.AccAvg(\"$age\")),\n\t))",
		},
		{
			name:     "pipeline with nested expressions",
			jsonStr:  `[{"$addFields": {"total": {"$add": ["$price", "$tax"]}}}]`,
			expected: "gmqb.NewPipeline().\n\tAddFields(gmqb.AddFieldsSpec(\n\t\tgmqb.AddField(\"total\", gmqb.ExprAdd(\"$price\", \"$tax\")),\n\t))",
		},
		{
			name:     "pipeline with complex expressions",
			jsonStr:  `[{"$set": {"isActive": {"$cond": {"if": {"$gte": ["$age", 18]}, "then": true, "else": false}}}}]`,
			expected: "gmqb.NewPipeline().\n\tSetFields(gmqb.AddFieldsSpec(\n\t\tgmqb.AddField(\"isActive\", gmqb.ExprCond(gmqb.ExprGte(\"$age\", 18), true, false)),\n\t))",
		},
		{
			name:     "pipeline with project and generic expression",
			jsonStr:  `[{"$project": {"name": {"$toUpper": "$name"}, "age": 1}}]`,
			expected: "gmqb.NewPipeline().\n\tProject(bson.D{\n\t\tbson.E{\"name\", gmqb.ExprToUpper(\"$name\")},\n\t\tbson.E{\"age\", 1},\n\t})",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Generate(tt.jsonStr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("GenerateFilter()\nGot:\n%v\nWant:\n%v", got, tt.expected)
			}
		})
	}
}
