package gmqb

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// --- Accumulator Operators (for $group and $setWindowFields) ---
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/#accumulators-group-project-addfields-etc

// AccSum returns the sum of numeric values. Use 1 to count documents.
//
// MongoDB equivalent: { $sum: expression }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/sum/
func AccSum(expression interface{}) bson.D { return bson.D{{Key: "$sum", Value: expression}} }

// AccAvg returns the average of numeric values.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/avg/
func AccAvg(expression interface{}) bson.D { return bson.D{{Key: "$avg", Value: expression}} }

// AccMin returns the minimum value.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/min/
func AccMin(expression interface{}) bson.D { return bson.D{{Key: "$min", Value: expression}} }

// AccMax returns the maximum value.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/max/
func AccMax(expression interface{}) bson.D { return bson.D{{Key: "$max", Value: expression}} }

// AccFirst returns the first value in a group.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/first/
func AccFirst(expression interface{}) bson.D { return bson.D{{Key: "$first", Value: expression}} }

// AccLast returns the last value in a group.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/last/
func AccLast(expression interface{}) bson.D { return bson.D{{Key: "$last", Value: expression}} }

// AccPush returns an array of all values in a group.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/push/
func AccPush(expression interface{}) bson.D { return bson.D{{Key: "$push", Value: expression}} }

// AccAddToSet returns an array of unique values in a group.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/addToSet/
func AccAddToSet(expression interface{}) bson.D {
	return bson.D{{Key: "$addToSet", Value: expression}}
}

// AccStdDevPop calculates the population standard deviation.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/stdDevPop/
func AccStdDevPop(expression interface{}) bson.D {
	return bson.D{{Key: "$stdDevPop", Value: expression}}
}

// AccStdDevSamp calculates the sample standard deviation.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/stdDevSamp/
func AccStdDevSamp(expression interface{}) bson.D {
	return bson.D{{Key: "$stdDevSamp", Value: expression}}
}

// AccCount returns the count of documents in a group. (MongoDB 5.0+)
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/count-accumulator/
func AccCount() bson.D { return bson.D{{Key: "$count", Value: bson.D{}}} }

// AccFirstN returns the first N values. (MongoDB 5.2+)
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/firstN/
func AccFirstN(expression interface{}, n interface{}) bson.D {
	return bson.D{{Key: "$firstN", Value: bson.D{{Key: "input", Value: expression}, {Key: "n", Value: n}}}}
}

// AccLastN returns the last N values. (MongoDB 5.2+)
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/lastN/
func AccLastN(expression interface{}, n interface{}) bson.D {
	return bson.D{{Key: "$lastN", Value: bson.D{{Key: "input", Value: expression}, {Key: "n", Value: n}}}}
}

// AccMaxN returns the N largest values. (MongoDB 5.2+)
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/maxN/
func AccMaxN(expression interface{}, n interface{}) bson.D {
	return bson.D{{Key: "$maxN", Value: bson.D{{Key: "input", Value: expression}, {Key: "n", Value: n}}}}
}

// AccMinN returns the N smallest values. (MongoDB 5.2+)
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/minN/
func AccMinN(expression interface{}, n interface{}) bson.D {
	return bson.D{{Key: "$minN", Value: bson.D{{Key: "input", Value: expression}, {Key: "n", Value: n}}}}
}

// AccTop returns the top element. (MongoDB 5.2+)
//
// MongoDB equivalent: { $top: { sortBy: { field: 1 }, output: expression } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/top/
func AccTop(sortBy bson.D, output interface{}) bson.D {
	return bson.D{{Key: "$top", Value: bson.D{{Key: "sortBy", Value: sortBy}, {Key: "output", Value: output}}}}
}

// AccBottom returns the bottom element. (MongoDB 5.2+)
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/bottom/
func AccBottom(sortBy bson.D, output interface{}) bson.D {
	return bson.D{{Key: "$bottom", Value: bson.D{{Key: "sortBy", Value: sortBy}, {Key: "output", Value: output}}}}
}

// AccTopN returns the top N elements. (MongoDB 5.2+)
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/topN/
func AccTopN(sortBy bson.D, output interface{}, n interface{}) bson.D {
	return bson.D{{Key: "$topN", Value: bson.D{
		{Key: "sortBy", Value: sortBy},
		{Key: "output", Value: output},
		{Key: "n", Value: n},
	}}}
}

// AccBottomN returns the bottom N elements. (MongoDB 5.2+)
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/bottomN/
func AccBottomN(sortBy bson.D, output interface{}, n interface{}) bson.D {
	return bson.D{{Key: "$bottomN", Value: bson.D{
		{Key: "sortBy", Value: sortBy},
		{Key: "output", Value: output},
		{Key: "n", Value: n},
	}}}
}

// AccMedian approximates the median. (MongoDB 7.0+)
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/median/
func AccMedian(input interface{}, method string) bson.D {
	return bson.D{{Key: "$median", Value: bson.D{
		{Key: "input", Value: input},
		{Key: "method", Value: method},
	}}}
}

// AccPercentile calculates the specified percentile(s). (MongoDB 7.0+)
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/percentile/
func AccPercentile(input interface{}, p bson.A, method string) bson.D {
	return bson.D{{Key: "$percentile", Value: bson.D{
		{Key: "input", Value: input},
		{Key: "p", Value: p},
		{Key: "method", Value: method},
	}}}
}

// --- Set Expression Operators ---
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/#set-expression-operators

// ExprSetEquals returns true if two sets have the same elements.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/setEquals/
func ExprSetEquals(arrays ...interface{}) bson.D {
	return bson.D{{Key: "$setEquals", Value: bson.A(arrays)}}
}

// ExprSetIntersection returns elements common to all input sets.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/setIntersection/
func ExprSetIntersection(arrays ...interface{}) bson.D {
	return bson.D{{Key: "$setIntersection", Value: bson.A(arrays)}}
}

// ExprSetUnion returns elements from any of the input sets.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/setUnion/
func ExprSetUnion(arrays ...interface{}) bson.D {
	return bson.D{{Key: "$setUnion", Value: bson.A(arrays)}}
}

// ExprSetDifference returns elements in the first set but not the second.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/setDifference/
func ExprSetDifference(arr1, arr2 interface{}) bson.D {
	return bson.D{{Key: "$setDifference", Value: bson.A{arr1, arr2}}}
}

// ExprSetIsSubset returns true if the first set is a subset of the second.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/setIsSubset/
func ExprSetIsSubset(arr1, arr2 interface{}) bson.D {
	return bson.D{{Key: "$setIsSubset", Value: bson.A{arr1, arr2}}}
}

// ExprAnyElementTrue returns true if any element of a set evaluates to true.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/anyElementTrue/
func ExprAnyElementTrue(array interface{}) bson.D {
	return bson.D{{Key: "$anyElementTrue", Value: array}}
}

// ExprAllElementsTrue returns true if all elements of a set evaluate to true.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/allElementsTrue/
func ExprAllElementsTrue(array interface{}) bson.D {
	return bson.D{{Key: "$allElementsTrue", Value: array}}
}

// --- Object Expression Operators ---
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/#object-expression-operators

// ExprMergeObjects combines multiple documents into a single document.
//
// MongoDB equivalent: { $mergeObjects: [ doc1, doc2, ... ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/mergeObjects/
func ExprMergeObjects(documents ...interface{}) bson.D {
	return bson.D{{Key: "$mergeObjects", Value: bson.A(documents)}}
}

// ExprGetField extracts a field from a document. (MongoDB 5.0+)
//
// MongoDB equivalent: { $getField: { field: "name", input: doc } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/getField/
func ExprGetField(field interface{}, input interface{}) bson.D {
	return bson.D{{Key: "$getField", Value: bson.D{
		{Key: "field", Value: field},
		{Key: "input", Value: input},
	}}}
}

// ExprSetField sets a field in a document. (MongoDB 5.0+)
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/setField/
func ExprSetField(field interface{}, input, value interface{}) bson.D {
	return bson.D{{Key: "$setField", Value: bson.D{
		{Key: "field", Value: field},
		{Key: "input", Value: input},
		{Key: "value", Value: value},
	}}}
}

// ExprUnsetField removes a field from a document. (MongoDB 5.0+)
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/unsetField/
func ExprUnsetField(field interface{}, input interface{}) bson.D {
	return bson.D{{Key: "$unsetField", Value: bson.D{
		{Key: "field", Value: field},
		{Key: "input", Value: input},
	}}}
}

// --- Literal & Miscellaneous ---

// ExprLiteral returns a value without parsing. Useful to pass a string that looks
// like an expression (starting with $) as a literal value.
//
// MongoDB equivalent: { $literal: value }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/literal/
func ExprLiteral(value interface{}) bson.D {
	return bson.D{{Key: "$literal", Value: value}}
}

// ExprRand generates a random float between 0 and 1. (MongoDB 4.4.2+)
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/rand/
func ExprRand() bson.D {
	return bson.D{{Key: "$rand", Value: bson.D{}}}
}

// ExprSampleRate probabilistically selects documents. rate is between 0 and 1. (MongoDB 4.4.2+)
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/sampleRate/
func ExprSampleRate(rate float64) bson.D {
	return bson.D{{Key: "$sampleRate", Value: rate}}
}

// ExprLet binds variables for use in a sub-expression.
//
// MongoDB equivalent: { $let: { vars: { var1: expr1 }, in: bodyExpr } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/let/
func ExprLet(vars bson.D, in interface{}) bson.D {
	return bson.D{{Key: "$let", Value: bson.D{
		{Key: "vars", Value: vars},
		{Key: "in", Value: in},
	}}}
}
