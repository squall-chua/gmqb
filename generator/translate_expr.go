package generator

import (
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// translateExpression handles aggregation expressions like $add, $sum, $filter
func translateExpression(doc bson.D) (string, error) {
	if len(doc) != 1 {
		return "", fmt.Errorf("expression document must have exactly one field, got %d", len(doc))
	}

	op := doc[0].Key
	val := doc[0].Value

	// Map of exact Operator -> gmqb Function Name
	// We handle them by argument arity
	switch op {
	// --- 1-Arg Expressions ---
	case "$abs", "$ceil", "$floor", "$sqrt", "$ln", "$type", "$isArray", "$toLower", "$toUpper",
		"$strLenCP", "$year", "$month", "$dayOfMonth", "$hour", "$minute", "$second", "$millisecond",
		"$dayOfYear", "$dayOfWeek", "$isoWeek", "$isoWeekYear", "$toDate", "$toDecimal", "$toDouble",
		"$toInt", "$toLong", "$toObjectId", "$toString", "$toBool", "$not", "$reverseArray",
		"$objectToArray", "$arrayToObject":
		funcName := "Expr" + strings.ToUpper(op[1:2]) + op[2:]
		if op == "$strLenCP" {
			funcName = "ExprStrLenCP"
		} else if op == "$isArray" {
			funcName = "ExprIsArray"
		} else if op == "$objectToArray" {
			funcName = "ExprObjectToArray"
		} else if op == "$arrayToObject" {
			funcName = "ExprArrayToObject"
		} else if strings.HasPrefix(op, "$to") {
			funcName = "Expr" + strings.ToUpper(op[1:2]) + op[2:]
			funcName = strings.ReplaceAll(funcName, "Id", "ID") // if any
		}

		return fmt.Sprintf("gmqb.%s(%s)", funcName, formatExpressionValue(val)), nil

	// --- Accumulators (1-Arg usually) ---
	case "$sum", "$avg", "$min", "$max", "$first", "$last", "$push", "$addToSet", "$stdDevPop", "$stdDevSamp", "$count":
		funcName := "Acc" + strings.ToUpper(op[1:2]) + op[2:]
		if op == "$count" {
			return "gmqb.AccCount()", nil
		}
		return fmt.Sprintf("gmqb.%s(%s)", funcName, formatExpressionValue(val)), nil

	// --- 2-Arg Array Expressions ---
	case "$subtract", "$divide", "$mod", "$pow", "$log", "$cmp", "$eq", "$ne", "$gt", "$gte", "$lt", "$lte",
		"$split", "$arrayElemAt", "$setDifference", "$setIsSubset", "$in", "$indexOfArray":
		arr, ok := val.(bson.A)
		if !ok || len(arr) != 2 {
			return fmt.Sprintf("bson.D{{%q, %s}}", op, formatValue(val)), nil // fallback
		}
		funcName := "Expr" + strings.ToUpper(op[1:2]) + op[2:]
		if op == "$arrayElemAt" {
			funcName = "ExprArrayElemAt"
		}
		if op == "$setDifference" {
			funcName = "ExprSetDifference"
		}
		if op == "$setIsSubset" {
			funcName = "ExprSetIsSubset"
		}
		if op == "$indexOfArray" {
			funcName = "ExprIndexOfArray"
		}

		return fmt.Sprintf("gmqb.%s(%s, %s)", funcName, formatValue(arr[0]), formatValue(arr[1])), nil

	// --- Variadic Array Expressions ---
	case "$add", "$multiply", "$concat", "$concatArrays", "$setEquals", "$setIntersection", "$setUnion",
		"$anyElementTrue", "$allElementsTrue", "$mergeObjects", "$zip", "$or", "$and":
		arr, ok := val.(bson.A)
		if !ok {
			return fmt.Sprintf("bson.D{{%q, %s}}", op, formatValue(val)), nil // fallback
		}
		funcName := "Expr" + strings.ToUpper(op[1:2]) + op[2:]
		if op == "$concatArrays" {
			funcName = "ExprConcatArrays"
		}
		if op == "$setEquals" {
			funcName = "ExprSetEquals"
		}
		if op == "$setIntersection" {
			funcName = "ExprSetIntersection"
		}
		if op == "$setUnion" {
			funcName = "ExprSetUnion"
		}
		if op == "$anyElementTrue" {
			funcName = "ExprAnyElementTrue"
		}
		if op == "$allElementsTrue" {
			funcName = "ExprAllElementsTrue"
		}
		if op == "$mergeObjects" {
			funcName = "ExprMergeObjects"
		}

		var parts []string
		for _, item := range arr {
			parts = append(parts, formatExpressionValue(item))
		}
		return fmt.Sprintf("gmqb.%s(%s)", funcName, strings.Join(parts, ", ")), nil

	// --- Structured Object Expressions ---
	case "$filter":
		d, ok := val.(bson.D)
		if ok {
			inp := getMapValue(d, "input")
			as := getMapString(d, "as")
			cond := getMapValue(d, "cond")
			return fmt.Sprintf("gmqb.ExprFilter(%s, %q, %s)", formatExpressionValue(inp), as, formatExpressionValue(cond)), nil
		}
	case "$map":
		d, ok := val.(bson.D)
		if ok {
			inp := getMapValue(d, "input")
			as := getMapString(d, "as")
			in := getMapValue(d, "in")
			return fmt.Sprintf("gmqb.ExprMap(%s, %q, %s)", formatExpressionValue(inp), as, formatExpressionValue(in)), nil
		}
	case "$reduce":
		d, ok := val.(bson.D)
		if ok {
			inp := getMapValue(d, "input")
			init := getMapValue(d, "initialValue")
			in := getMapValue(d, "in")
			return fmt.Sprintf("gmqb.ExprReduce(%s, %s, %s)", formatExpressionValue(inp), formatExpressionValue(init), formatExpressionValue(in)), nil
		}
	case "$let":
		d, ok := val.(bson.D)
		if ok {
			vars := getMapValue(d, "vars")
			in := getMapValue(d, "in")
			return fmt.Sprintf("gmqb.ExprLet(%s, %s)", formatExpressionValue(vars), formatExpressionValue(in)), nil
		}
	case "$regexMatch", "$regexFind", "$regexFindAll":
		d, ok := val.(bson.D)
		if ok {
			inp := getMapValue(d, "input")
			regex := getMapString(d, "regex")
			options := getMapString(d, "options")
			funcName := "ExprRegexMatch"
			if op == "$regexFind" {
				funcName = "ExprRegexFind"
			}
			if op == "$regexFindAll" {
				funcName = "ExprRegexFindAll"
			}
			return fmt.Sprintf("gmqb.%s(%s, %q, %q)", funcName, formatExpressionValue(inp), regex, options), nil
		}
	case "$replaceOne", "$replaceAll":
		d, ok := val.(bson.D)
		if ok {
			inp := getMapValue(d, "input")
			find := getMapValue(d, "find")
			repl := getMapValue(d, "replacement")
			funcName := "ExprReplaceOne"
			if op == "$replaceAll" {
				funcName = "ExprReplaceAll"
			}
			return fmt.Sprintf("gmqb.%s(%s, %s, %s)", funcName, formatExpressionValue(inp), formatExpressionValue(find), formatExpressionValue(repl)), nil
		}
	case "$trim", "$ltrim", "$rtrim":
		d, ok := val.(bson.D)
		if ok {
			inp := getMapValue(d, "input")
			chars := getMapValue(d, "chars")
			funcName := "ExprTrim"
			if op == "$ltrim" {
				funcName = "ExprLTrim"
			}
			if op == "$rtrim" {
				funcName = "ExprRTrim"
			}
			return fmt.Sprintf("gmqb.%s(%s, %s)", funcName, formatExpressionValue(inp), formatExpressionValue(chars)), nil
		}
	case "$cond":
		if d, ok := val.(bson.D); ok {
			ifV := getMapValue(d, "if")
			thenV := getMapValue(d, "then")
			elseV := getMapValue(d, "else")
			return fmt.Sprintf("gmqb.ExprCond(%s, %s, %s)", formatExpressionValue(ifV), formatExpressionValue(thenV), formatExpressionValue(elseV)), nil
		} else if arr, ok := val.(bson.A); ok && len(arr) == 3 {
			return fmt.Sprintf("gmqb.ExprCond(%s, %s, %s)", formatExpressionValue(arr[0]), formatExpressionValue(arr[1]), formatExpressionValue(arr[2])), nil
		}
	case "$switch":
		d, ok := val.(bson.D)
		if ok {
			branches := getMapValue(d, "branches")
			def := getMapValue(d, "default")
			return fmt.Sprintf("gmqb.ExprSwitch(%s, %s)", formatExpressionValue(branches), formatExpressionValue(def)), nil
		}
	}

	// Native fallback
	return fmt.Sprintf("bson.D{{%q, %s}}", op, formatExpressionValue(val)), nil
}

func getMapValue(d bson.D, key string) interface{} {
	for _, e := range d {
		if e.Key == key {
			return e.Value
		}
	}
	return nil
}
