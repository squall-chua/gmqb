package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Generate is a versatile helper that parses a MongoDB extended JSON string and returns the equivalent gmqb Go code.
// It auto-detects arrays as Pipelines and objects as Filters.
func Generate(jsonStr string) (string, error) {
	str := strings.TrimSpace(jsonStr)
	if strings.HasPrefix(str, "[") {
		return GeneratePipeline(str)
	}
	return GenerateFilter(str)
}

// GenerateFilter parses a MongoDB extended JSON object string and returns the equivalent gmqb Filter Go code.
func GenerateFilter(jsonStr string) (string, error) {
	var doc bson.D
	err := bson.UnmarshalExtJSON([]byte(jsonStr), true, &doc)
	if err != nil {
		return "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	code, err := translateDoc(doc)
	if err != nil {
		return "", err
	}

	// Format the generated code
	formattedCode, err := format.Source([]byte(code))
	if err != nil {
		return code, fmt.Errorf("failed to format generated code: %w\nOutput was:\n%s", err, code)
	}

	return string(bytes.TrimSpace(formattedCode)), nil
}

// GeneratePipeline parses a MongoDB extended JSON array string and returns the equivalent gmqb Pipeline Go code.
func GeneratePipeline(jsonStr string) (string, error) {
	var arr bson.A
	err := bson.UnmarshalExtJSON([]byte(jsonStr), true, &arr)
	if err != nil {
		return "", fmt.Errorf("failed to parse JSON array: %w", err)
	}

	code, err := translatePipeline(arr)
	if err != nil {
		return "", err
	}

	// Format the generated code
	formattedCode, err := format.Source([]byte(code))
	if err != nil {
		return code, fmt.Errorf("failed to format generated code: %w\nOutput was:\n%s", err, code)
	}

	return string(bytes.TrimSpace(formattedCode)), nil
}
