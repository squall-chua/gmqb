package gmqb

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// --- Aggregation Expression Helpers ---
// These functions construct aggregation expression operators for use within
// pipeline stages like $project, $addFields, $group, $match ($expr), etc.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/

// --- Arithmetic Expression Operators ---
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/#arithmetic-expression-operators

// ExprAdd returns an expression that adds numbers/dates together.
//
// MongoDB equivalent: { $add: [ expr1, expr2, ... ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/add/
//
// Example:
//
//	gmqb.ExprAdd("$price", "$tax") // { "$add": ["$price", "$tax"] }
func ExprAdd(expressions ...interface{}) bson.D {
	return bson.D{{Key: "$add", Value: bson.A(expressions)}}
}

// ExprSubtract returns the difference of two expressions.
//
// MongoDB equivalent: { $subtract: [ expr1, expr2 ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/subtract/
func ExprSubtract(expr1, expr2 interface{}) bson.D {
	return bson.D{{Key: "$subtract", Value: bson.A{expr1, expr2}}}
}

// ExprMultiply returns the product of expressions.
//
// MongoDB equivalent: { $multiply: [ expr1, expr2, ... ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/multiply/
func ExprMultiply(expressions ...interface{}) bson.D {
	return bson.D{{Key: "$multiply", Value: bson.A(expressions)}}
}

// ExprDivide divides one expression by another.
//
// MongoDB equivalent: { $divide: [ expr1, expr2 ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/divide/
func ExprDivide(dividend, divisor interface{}) bson.D {
	return bson.D{{Key: "$divide", Value: bson.A{dividend, divisor}}}
}

// ExprMod returns the remainder of dividing the first expression by the second.
//
// MongoDB equivalent: { $mod: [ expr1, expr2 ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/mod/
func ExprMod(dividend, divisor interface{}) bson.D {
	return bson.D{{Key: "$mod", Value: bson.A{dividend, divisor}}}
}

// ExprAbs returns the absolute value.
//
// MongoDB equivalent: { $abs: expression }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/abs/
func ExprAbs(expression interface{}) bson.D {
	return bson.D{{Key: "$abs", Value: expression}}
}

// ExprCeil returns the smallest integer >= the expression.
//
// MongoDB equivalent: { $ceil: expression }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/ceil/
func ExprCeil(expression interface{}) bson.D {
	return bson.D{{Key: "$ceil", Value: expression}}
}

// ExprFloor returns the largest integer <= the expression.
//
// MongoDB equivalent: { $floor: expression }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/floor/
func ExprFloor(expression interface{}) bson.D {
	return bson.D{{Key: "$floor", Value: expression}}
}

// ExprRound rounds to a specified decimal place.
//
// MongoDB equivalent: { $round: [ expression, place ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/round/
func ExprRound(expression interface{}, place int) bson.D {
	return bson.D{{Key: "$round", Value: bson.A{expression, place}}}
}

// ExprPow raises a number to an exponent.
//
// MongoDB equivalent: { $pow: [ base, exponent ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/pow/
func ExprPow(base, exponent interface{}) bson.D {
	return bson.D{{Key: "$pow", Value: bson.A{base, exponent}}}
}

// ExprSqrt returns the square root.
//
// MongoDB equivalent: { $sqrt: expression }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/sqrt/
func ExprSqrt(expression interface{}) bson.D {
	return bson.D{{Key: "$sqrt", Value: expression}}
}

// ExprLog calculates the log of a number in the specified base.
//
// MongoDB equivalent: { $log: [ number, base ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/log/
func ExprLog(number, base interface{}) bson.D {
	return bson.D{{Key: "$log", Value: bson.A{number, base}}}
}

// ExprLn calculates the natural log of a number.
//
// MongoDB equivalent: { $ln: expression }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/ln/
func ExprLn(expression interface{}) bson.D {
	return bson.D{{Key: "$ln", Value: expression}}
}

// --- Comparison Expression Operators ---
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/#comparison-expression-operators

// ExprCmp compares two values. Returns -1, 0, or 1.
//
// MongoDB equivalent: { $cmp: [ expr1, expr2 ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/cmp/
func ExprCmp(expr1, expr2 interface{}) bson.D {
	return bson.D{{Key: "$cmp", Value: bson.A{expr1, expr2}}}
}

// ExprEq returns true if two expressions are equal.
//
// MongoDB equivalent: { $eq: [ expr1, expr2 ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/eq/
func ExprEq(expr1, expr2 interface{}) bson.D {
	return bson.D{{Key: "$eq", Value: bson.A{expr1, expr2}}}
}

// ExprNe returns true if two expressions are not equal.
//
// MongoDB equivalent: { $ne: [ expr1, expr2 ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/ne/
func ExprNe(expr1, expr2 interface{}) bson.D {
	return bson.D{{Key: "$ne", Value: bson.A{expr1, expr2}}}
}

// ExprGt returns true if the first expression is greater than the second.
//
// MongoDB equivalent: { $gt: [ expr1, expr2 ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/gt/
func ExprGt(expr1, expr2 interface{}) bson.D {
	return bson.D{{Key: "$gt", Value: bson.A{expr1, expr2}}}
}

// ExprGte returns true if the first expression is >= the second.
//
// MongoDB equivalent: { $gte: [ expr1, expr2 ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/gte/
func ExprGte(expr1, expr2 interface{}) bson.D {
	return bson.D{{Key: "$gte", Value: bson.A{expr1, expr2}}}
}

// ExprLt returns true if the first expression is less than the second.
//
// MongoDB equivalent: { $lt: [ expr1, expr2 ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/lt/
func ExprLt(expr1, expr2 interface{}) bson.D {
	return bson.D{{Key: "$lt", Value: bson.A{expr1, expr2}}}
}

// ExprLte returns true if the first expression is <= the second.
//
// MongoDB equivalent: { $lte: [ expr1, expr2 ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/lte/
func ExprLte(expr1, expr2 interface{}) bson.D {
	return bson.D{{Key: "$lte", Value: bson.A{expr1, expr2}}}
}

// --- Boolean Expression Operators ---
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/#boolean-expression-operators

// ExprBoolAnd evaluates one or more expressions and returns true if all evaluate to true.
//
// MongoDB equivalent: { $and: [ expr1, expr2, ... ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/and/
func ExprBoolAnd(expressions ...interface{}) bson.D {
	return bson.D{{Key: "$and", Value: bson.A(expressions)}}
}

// ExprBoolOr evaluates one or more expressions and returns true if any evaluate to true.
//
// MongoDB equivalent: { $or: [ expr1, expr2, ... ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/or/
func ExprBoolOr(expressions ...interface{}) bson.D {
	return bson.D{{Key: "$or", Value: bson.A(expressions)}}
}

// ExprBoolNot returns the boolean opposite of the expression.
//
// MongoDB equivalent: { $not: [ expression ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/not/
func ExprBoolNot(expression interface{}) bson.D {
	return bson.D{{Key: "$not", Value: bson.A{expression}}}
}

// --- Conditional Expression Operators ---
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/#conditional-expression-operators

// ExprCond evaluates a boolean expression and returns one of two values.
//
// MongoDB equivalent: { $cond: { if: bool, then: trueExpr, else: falseExpr } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/cond/
//
// Example:
//
//	gmqb.ExprCond(
//	    gmqb.ExprGte("$qty", 250),
//	    "high",
//	    "low",
//	)
func ExprCond(boolExpr, trueExpr, falseExpr interface{}) bson.D {
	return bson.D{{Key: "$cond", Value: bson.D{
		{Key: "if", Value: boolExpr},
		{Key: "then", Value: trueExpr},
		{Key: "else", Value: falseExpr},
	}}}
}

// ExprIfNull returns the first non-null expression, or a replacement value.
//
// MongoDB equivalent: { $ifNull: [ expression, replacement ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/ifNull/
func ExprIfNull(expression, replacement interface{}) bson.D {
	return bson.D{{Key: "$ifNull", Value: bson.A{expression, replacement}}}
}

// SwitchBranch represents a single branch in a $switch expression.
type SwitchBranch struct {
	Case interface{}
	Then interface{}
}

// ExprSwitch evaluates a series of case expressions and returns the value of the
// first expression that evaluates to true.
//
// MongoDB equivalent: { $switch: { branches: [...], default: defaultExpr } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/switch/
//
// Example:
//
//	gmqb.ExprSwitch([]gmqb.SwitchBranch{
//	    {Case: gmqb.ExprGte("$age", 65), Then: "senior"},
//	    {Case: gmqb.ExprGte("$age", 18), Then: "adult"},
//	}, "minor")
func ExprSwitch(branches []SwitchBranch, defaultExpr interface{}) bson.D {
	branchArr := make(bson.A, len(branches))
	for i, b := range branches {
		branchArr[i] = bson.D{{Key: "case", Value: b.Case}, {Key: "then", Value: b.Then}}
	}
	doc := bson.D{{Key: "branches", Value: branchArr}}
	if defaultExpr != nil {
		doc = append(doc, bson.E{Key: "default", Value: defaultExpr})
	}
	return bson.D{{Key: "$switch", Value: doc}}
}
