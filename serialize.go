package gmqb

import (
	"bytes"
	"encoding/json"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Serializable is implemented by all builders (Filter, Updater, Pipeline)
// to produce BSON and JSON output.
type Serializable interface {
	// BsonD returns the builder's output as a bson.D, suitable for passing
	// directly to the go-mongodb-driver.
	BsonD() bson.D

	// JSON returns the builder's output as a pretty-printed JSON string.
	// Useful for debugging and logging queries.
	JSON() string

	// CompactJSON returns the builder's output as a compact (no whitespace) JSON string.
	CompactJSON() string
}

// toJSON converts a bson.D to a pretty-printed JSON string.
func toJSON(d bson.D) string {
	raw, err := bson.MarshalExtJSON(d, false, false)
	if err != nil {
		return "{}"
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, raw, "", "  "); err != nil {
		return string(raw)
	}
	return buf.String()
}

// toCompactJSON converts a bson.D to a compact JSON string.
func toCompactJSON(d bson.D) string {
	raw, err := bson.MarshalExtJSON(d, false, false)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

// pipelineToJSON converts a mongo.Pipeline ([]bson.D) to a pretty-printed JSON array string.
func pipelineToJSON(stages []bson.D) string {
	result := make([]json.RawMessage, 0, len(stages))
	for _, stage := range stages {
		raw, err := bson.MarshalExtJSON(stage, false, false)
		if err != nil {
			continue
		}
		result = append(result, raw)
	}
	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "[]"
	}
	return string(out)
}

// pipelineToCompactJSON converts a mongo.Pipeline to a compact JSON array string.
func pipelineToCompactJSON(stages []bson.D) string {
	result := make([]json.RawMessage, 0, len(stages))
	for _, stage := range stages {
		raw, err := bson.MarshalExtJSON(stage, false, false)
		if err != nil {
			continue
		}
		result = append(result, raw)
	}
	out, err := json.Marshal(result)
	if err != nil {
		return "[]"
	}
	return string(out)
}
