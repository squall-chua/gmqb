package generator

import (
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func translatePipeline(arr bson.A) (string, error) {
	if len(arr) == 0 {
		return "gmqb.NewPipeline()", nil
	}

	var parts []string
	for _, item := range arr {
		doc, ok := item.(bson.D)
		if !ok {
			return "", fmt.Errorf("expected document as pipeline stage, got %T", item)
		}

		stageStr, err := translatePipelineStage(doc)
		if err != nil {
			return "", err
		}
		parts = append(parts, stageStr)
	}

	return fmt.Sprintf("gmqb.NewPipeline().\n%s", strings.Join(parts, ".\n")), nil
}

func translatePipelineStage(doc bson.D) (string, error) {
	if len(doc) == 0 {
		return "", fmt.Errorf("empty pipeline stage")
	}

	if len(doc) > 1 {
		return "", fmt.Errorf("pipeline stage must have exactly one top-level operator, got %d", len(doc))
	}

	stageOp := doc[0].Key
	stageVal := doc[0].Value

	switch stageOp {
	case "$match":
		// Similar to translateDoc but wrapped in .Match() depending on contents
		filterDoc, ok := stageVal.(bson.D)
		if !ok {
			return "", fmt.Errorf("expected document $match, got %T", stageVal)
		}

		// Translate interior
		part, err := translateDoc(filterDoc)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Match(%s)", part), nil

	case "$addFields", "$set":
		bMethod := "AddFields"
		if stageOp == "$set" {
			bMethod = "SetFields"
		}
		d, ok := stageVal.(bson.D)
		if !ok {
			return "", fmt.Errorf("expected document for %s, got %T", stageOp, stageVal)
		}
		parts := []string{}
		for _, e := range d {
			valStr := formatExpressionValue(e.Value)
			parts = append(parts, fmt.Sprintf("gmqb.AddField(%q, %s)", e.Key, valStr))
		}
		return fmt.Sprintf("%s(gmqb.AddFieldsSpec(\n\t%s,\n))", bMethod, strings.Join(parts, ",\n\t")), nil

	case "$project", "$sort":
		d, ok := stageVal.(bson.D)
		if !ok {
			return "", fmt.Errorf("expected document for %s, got %T", stageOp, stageVal)
		}
		parts := []string{}
		for _, e := range d {
			valStr := formatExpressionValue(e.Value)
			parts = append(parts, fmt.Sprintf("bson.E{%q, %s}", e.Key, valStr))
		}
		bMethod := "Project"
		if stageOp == "$sort" {
			bMethod = "Sort"
		}
		return fmt.Sprintf("%s(bson.D{\n\t%s,\n})", bMethod, strings.Join(parts, ",\n\t")), nil

	case "$group":
		d, ok := stageVal.(bson.D)
		if !ok {
			return "", fmt.Errorf("expected document for %s, got %T", stageOp, stageVal)
		}
		parts := []string{}
		idVal := formatExpressionValue(getMapValue(d, "_id"))
		for _, e := range d {
			if e.Key == "_id" {
				continue
			}
			valStr := formatExpressionValue(e.Value)
			parts = append(parts, fmt.Sprintf("gmqb.GroupAcc(%q, %s)", e.Key, valStr))
		}
		if len(parts) == 0 {
			return fmt.Sprintf("Group(gmqb.GroupSpec(%s))", idVal), nil
		}
		return fmt.Sprintf("Group(gmqb.GroupSpec(%s,\n\t%s,\n))", idVal, strings.Join(parts, ",\n\t")), nil

	case "$limit":
		return fmt.Sprintf("Limit(%d)", getInt64(stageVal)), nil
	case "$skip":
		return fmt.Sprintf("Skip(%d)", getInt64(stageVal)), nil
	case "$count":
		return fmt.Sprintf("Count(%q)", stageVal), nil
	case "$unwind":
		switch v := stageVal.(type) {
		case string:
			return fmt.Sprintf("Unwind(%q)", v), nil
		case bson.D:
			optsStr := formatValue(v) // Usually UnwindWithOpts logic here, fallback to RawStage for now if complex payload but we will do simple format
			return fmt.Sprintf("RawStage(\"$unwind\", %s)", optsStr), nil
		}
	case "$replaceRoot":
		return fmt.Sprintf("ReplaceRoot(%s)", formatValue(stageVal)), nil
	case "$replaceWith":
		return fmt.Sprintf("ReplaceWith(%s)", formatValue(stageVal)), nil
	case "$out":
		if str, ok := stageVal.(string); ok {
			return fmt.Sprintf("Out(%q)", str), nil
		}
		if d, ok := stageVal.(bson.D); ok {
			db := getMapString(d, "db")
			coll := getMapString(d, "coll")
			return fmt.Sprintf("OutToDb(%q, %q)", db, coll), nil
		}
		return fmt.Sprintf("RawStage(\"$out\", %s)", formatValue(stageVal)), nil
	}

	// Fallback to RawStage
	return fmt.Sprintf("RawStage(%q, %s)", stageOp, formatValue(stageVal)), nil
}

func getInt64(val interface{}) int64 {
	switch v := val.(type) {
	case int32:
		return int64(v)
	case int64:
		return v
	case float64:
		return int64(v)
	default:
		return 0
	}
}

func getMapString(d bson.D, key string) string {
	for _, e := range d {
		if e.Key == key {
			if s, ok := e.Value.(string); ok {
				return s
			}
		}
	}
	return ""
}

func translateDoc(doc bson.D) (string, error) {
	if len(doc) == 0 {
		return "", nil
	}

	var parts []string
	for _, elem := range doc {
		part, err := translateElement(elem.Key, elem.Value)
		if err != nil {
			return "", err
		}
		parts = append(parts, part)
	}

	if len(parts) == 1 {
		return parts[0], nil
	}

	// Implicit AND for multiple top-level fields
	return fmt.Sprintf("gmqb.And(\n%s,\n)", strings.Join(parts, ",\n")), nil
}

func translateElement(key string, val interface{}) (string, error) {
	// Check if the key is a logical top level operator ($and, $or, etc)
	switch key {
	case "$and":
		return translateLogicalOp("gmqb.And", val)
	case "$or":
		return translateLogicalOp("gmqb.Or", val)
	case "$nor":
		return translateLogicalOp("gmqb.Nor", val)
	}

	// If val is a document, check if it contains operators
	if subDoc, ok := val.(bson.D); ok && len(subDoc) > 0 {
		// Does the subdocument have operator keys?
		hasOps := true
		for _, e := range subDoc {
			if !strings.HasPrefix(e.Key, "$") {
				hasOps = false
				break
			}
		}

		if hasOps {
			var parts []string
			for _, e := range subDoc {
				part, err := translateOperator(key, e.Key, e.Value)
				if err != nil {
					return "", err
				}
				parts = append(parts, part)
			}
			if len(parts) == 1 {
				return parts[0], nil
			}
			return fmt.Sprintf("gmqb.And(\n%s,\n)", strings.Join(parts, ",\n")), nil
		}
	}

	// Default to gmqb.Eq("key", val)
	valStr := formatValue(val)
	return fmt.Sprintf("gmqb.Eq(%q, %s)", key, valStr), nil
}

func translateLogicalOp(funcName string, val interface{}) (string, error) {
	arr, ok := val.(bson.A)
	if !ok {
		return "", fmt.Errorf("expected array for logical operator, got %T", val)
	}

	var parts []string
	for _, item := range arr {
		doc, ok := item.(bson.D)
		if !ok {
			return "", fmt.Errorf("expected document in logical operator array, got %T", item)
		}
		part, err := translateDoc(doc)
		if err != nil {
			return "", err
		}
		parts = append(parts, part)
	}

	return fmt.Sprintf("%s(\n%s,\n)", funcName, strings.Join(parts, ",\n")), nil
}

func translateOperator(field string, op string, val interface{}) (string, error) {
	switch op {
	case "$eq":
		return fmt.Sprintf("gmqb.Eq(%q, %s)", field, formatValue(val)), nil
	case "$ne":
		return fmt.Sprintf("gmqb.Ne(%q, %s)", field, formatValue(val)), nil
	case "$gt":
		return fmt.Sprintf("gmqb.Gt(%q, %s)", field, formatValue(val)), nil
	case "$gte":
		return fmt.Sprintf("gmqb.Gte(%q, %s)", field, formatValue(val)), nil
	case "$lt":
		return fmt.Sprintf("gmqb.Lt(%q, %s)", field, formatValue(val)), nil
	case "$lte":
		return fmt.Sprintf("gmqb.Lte(%q, %s)", field, formatValue(val)), nil
	case "$in":
		return fmt.Sprintf("gmqb.In(%q, %s)", field, formatValues(val)), nil
	case "$nin":
		return fmt.Sprintf("gmqb.Nin(%q, %s)", field, formatValues(val)), nil
	case "$exists":
		return fmt.Sprintf("gmqb.Exists(%q, %s)", field, formatValue(val)), nil
	case "$type":
		return fmt.Sprintf("gmqb.Type(%q, %s)", field, formatValue(val)), nil
	case "$regex":
		return fmt.Sprintf("gmqb.Regex(%q, %s, \"\")", field, formatValue(val)), nil // TODO: regex opts
	case "$mod":
		arr, ok := val.(bson.A)
		if !ok || len(arr) != 2 {
			return "", fmt.Errorf("invalid $mod value")
		}
		return fmt.Sprintf("gmqb.Mod(%q, %s, %s)", field, formatValue(arr[0]), formatValue(arr[1])), nil
	case "$size":
		return fmt.Sprintf("gmqb.Size(%q, %s)", field, formatValue(val)), nil
	case "$all":
		return fmt.Sprintf("gmqb.All(%q, %s)", field, formatValues(val)), nil
	case "$elemMatch":
		subDoc, ok := val.(bson.D)
		if !ok {
			return "", fmt.Errorf("invalid $elemMatch value")
		}
		part, err := translateDoc(subDoc)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("gmqb.ElemMatch(%q, %s)", field, part), nil
	default:
		return "", fmt.Errorf("unsupported operator: %s", op)
	}
}

func formatValues(val interface{}) string {
	arr, ok := val.(bson.A)
	if !ok {
		return formatValue(val)
	}
	var parts []string
	for _, item := range arr {
		parts = append(parts, formatValue(item))
	}
	return strings.Join(parts, ", ")
}

// formatExpressionValue attempts to translate expressions before falling back to formatValue
func formatExpressionValue(val interface{}) string {
	if d, ok := val.(bson.D); ok && len(d) == 1 && strings.HasPrefix(d[0].Key, "$") {
		expr, err := translateExpression(d)
		if err == nil {
			return expr
		}
	}
	return formatValue(val)
}

func formatValue(val interface{}) string {
	switch v := val.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	case int32:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%v", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case bson.A:
		var inner []string
		for _, e := range v {
			inner = append(inner, formatValue(e))
		}
		return fmt.Sprintf("bson.A{%s}", strings.Join(inner, ", "))
	case bson.D:
		var inner []string
		for _, e := range v {
			inner = append(inner, fmt.Sprintf("{%q, %s}", e.Key, formatValue(e.Value)))
		}
		return fmt.Sprintf("bson.D{%s}", strings.Join(inner, ", "))
	case bson.ObjectID:
		return fmt.Sprintf("func() bson.ObjectID { id, _ := bson.ObjectIDFromHex(%q); return id }()", v.Hex())
	case bson.Regex:
		// Not strictly correct if they pass regex as {"$regex": "pat"} vs {"$regex": {"$regularExpression": ...}}
		return fmt.Sprintf("%q", v.Pattern)
	default:
		return fmt.Sprintf("%#v", v)
	}
}
