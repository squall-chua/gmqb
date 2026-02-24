package gmqb

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// --- String Expression Operators ---
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/#string-expression-operators

// ExprConcat concatenates strings.
//
// MongoDB equivalent: { $concat: [ expr1, expr2, ... ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/concat/
func ExprConcat(expressions ...interface{}) bson.D {
	return bson.D{{Key: "$concat", Value: bson.A(expressions)}}
}

// ExprSubstr returns a substring. start is 0-based, length is the char count.
//
// MongoDB equivalent: { $substr: [ string, start, length ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/substr/
func ExprSubstr(str interface{}, start, length int) bson.D {
	return bson.D{{Key: "$substr", Value: bson.A{str, start, length}}}
}

// ExprToLower converts a string to lowercase.
//
// MongoDB equivalent: { $toLower: expression }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/toLower/
func ExprToLower(expression interface{}) bson.D {
	return bson.D{{Key: "$toLower", Value: expression}}
}

// ExprToUpper converts a string to uppercase.
//
// MongoDB equivalent: { $toUpper: expression }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/toUpper/
func ExprToUpper(expression interface{}) bson.D {
	return bson.D{{Key: "$toUpper", Value: expression}}
}

// ExprTrim removes whitespace or specified characters from a string.
//
// MongoDB equivalent: { $trim: { input: expr, chars: charsExpr } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/trim/
func ExprTrim(input interface{}, chars interface{}) bson.D {
	doc := bson.D{{Key: "input", Value: input}}
	if chars != nil {
		doc = append(doc, bson.E{Key: "chars", Value: chars})
	}
	return bson.D{{Key: "$trim", Value: doc}}
}

// ExprLTrim removes leading whitespace or characters.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/ltrim/
func ExprLTrim(input interface{}, chars interface{}) bson.D {
	doc := bson.D{{Key: "input", Value: input}}
	if chars != nil {
		doc = append(doc, bson.E{Key: "chars", Value: chars})
	}
	return bson.D{{Key: "$ltrim", Value: doc}}
}

// ExprRTrim removes trailing whitespace or characters.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/rtrim/
func ExprRTrim(input interface{}, chars interface{}) bson.D {
	doc := bson.D{{Key: "input", Value: input}}
	if chars != nil {
		doc = append(doc, bson.E{Key: "chars", Value: chars})
	}
	return bson.D{{Key: "$rtrim", Value: doc}}
}

// ExprStrLenCP returns the number of UTF-8 code points in a string.
//
// MongoDB equivalent: { $strLenCP: expression }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/strLenCP/
func ExprStrLenCP(expression interface{}) bson.D {
	return bson.D{{Key: "$strLenCP", Value: expression}}
}

// ExprRegexMatch returns true if a string matches a regex pattern.
//
// MongoDB equivalent: { $regexMatch: { input: str, regex: pattern, options: opts } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/regexMatch/
func ExprRegexMatch(input interface{}, regex string, options string) bson.D {
	doc := bson.D{{Key: "input", Value: input}, {Key: "regex", Value: regex}}
	if options != "" {
		doc = append(doc, bson.E{Key: "options", Value: options})
	}
	return bson.D{{Key: "$regexMatch", Value: doc}}
}

// ExprRegexFind returns the first regex match.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/regexFind/
func ExprRegexFind(input interface{}, regex string, options string) bson.D {
	doc := bson.D{{Key: "input", Value: input}, {Key: "regex", Value: regex}}
	if options != "" {
		doc = append(doc, bson.E{Key: "options", Value: options})
	}
	return bson.D{{Key: "$regexFind", Value: doc}}
}

// ExprRegexFindAll returns all regex matches.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/regexFindAll/
func ExprRegexFindAll(input interface{}, regex string, options string) bson.D {
	doc := bson.D{{Key: "input", Value: input}, {Key: "regex", Value: regex}}
	if options != "" {
		doc = append(doc, bson.E{Key: "options", Value: options})
	}
	return bson.D{{Key: "$regexFindAll", Value: doc}}
}

// ExprSplit splits a string by a delimiter.
//
// MongoDB equivalent: { $split: [ string, delimiter ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/split/
func ExprSplit(str, delimiter interface{}) bson.D {
	return bson.D{{Key: "$split", Value: bson.A{str, delimiter}}}
}

// ExprReplaceOne replaces the first occurrence of a substring.
//
// MongoDB equivalent: { $replaceOne: { input: str, find: substr, replacement: repl } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/replaceOne/
func ExprReplaceOne(input, find, replacement interface{}) bson.D {
	return bson.D{{Key: "$replaceOne", Value: bson.D{
		{Key: "input", Value: input},
		{Key: "find", Value: find},
		{Key: "replacement", Value: replacement},
	}}}
}

// ExprReplaceAll replaces all occurrences of a substring.
//
// MongoDB equivalent: { $replaceAll: { input: str, find: substr, replacement: repl } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/replaceAll/
func ExprReplaceAll(input, find, replacement interface{}) bson.D {
	return bson.D{{Key: "$replaceAll", Value: bson.D{
		{Key: "input", Value: input},
		{Key: "find", Value: find},
		{Key: "replacement", Value: replacement},
	}}}
}

// --- Array Expression Operators ---
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/#array-expression-operators

// ExprArrayElemAt returns the element at a specified index.
//
// MongoDB equivalent: { $arrayElemAt: [ array, index ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/arrayElemAt/
func ExprArrayElemAt(array interface{}, index int) bson.D {
	return bson.D{{Key: "$arrayElemAt", Value: bson.A{array, index}}}
}

// ExprConcatArrays concatenates arrays.
//
// MongoDB equivalent: { $concatArrays: [ array1, array2, ... ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/concatArrays/
func ExprConcatArrays(arrays ...interface{}) bson.D {
	return bson.D{{Key: "$concatArrays", Value: bson.A(arrays)}}
}

// ExprFilter selects a subset of an array based on a condition.
//
// MongoDB equivalent: { $filter: { input: array, as: var, cond: expr } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/filter/
//
// Example:
//
//	gmqb.ExprFilter("$items", "item", gmqb.ExprGte("$$item.price", 100))
func ExprFilter(input interface{}, as string, cond interface{}) bson.D {
	return bson.D{{Key: "$filter", Value: bson.D{
		{Key: "input", Value: input},
		{Key: "as", Value: as},
		{Key: "cond", Value: cond},
	}}}
}

// ExprIsArray returns true if the expression is an array.
//
// MongoDB equivalent: { $isArray: expression }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/isArray/
func ExprIsArray(expression interface{}) bson.D {
	return bson.D{{Key: "$isArray", Value: expression}}
}

// ExprMap applies an expression to each element of an array.
//
// MongoDB equivalent: { $map: { input: array, as: var, in: expr } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/map/
func ExprMap(input interface{}, as string, in interface{}) bson.D {
	return bson.D{{Key: "$map", Value: bson.D{
		{Key: "input", Value: input},
		{Key: "as", Value: as},
		{Key: "in", Value: in},
	}}}
}

// ExprReduce applies an expression to each element and combines them into a single value.
//
// MongoDB equivalent: { $reduce: { input: array, initialValue: init, in: expr } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/reduce/
func ExprReduce(input, initialValue, in interface{}) bson.D {
	return bson.D{{Key: "$reduce", Value: bson.D{
		{Key: "input", Value: input},
		{Key: "initialValue", Value: initialValue},
		{Key: "in", Value: in},
	}}}
}

// ExprSlice returns a subset of an array.
//
// MongoDB equivalent: { $slice: [ array, n ] } or { $slice: [ array, position, n ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/slice/
func ExprSlice(array interface{}, args ...int) bson.D {
	a := bson.A{array}
	for _, v := range args {
		a = append(a, v)
	}
	return bson.D{{Key: "$slice", Value: a}}
}

// ExprIn returns true if a value is in an array. (Aggregation expression version.)
//
// MongoDB equivalent: { $in: [ value, array ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/in/
func ExprIn(value, array interface{}) bson.D {
	return bson.D{{Key: "$in", Value: bson.A{value, array}}}
}

// ExprIndexOfArray returns the index of the first occurrence of a value in an array.
//
// MongoDB equivalent: { $indexOfArray: [ array, value ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/indexOfArray/
func ExprIndexOfArray(array, value interface{}) bson.D {
	return bson.D{{Key: "$indexOfArray", Value: bson.A{array, value}}}
}

// ExprReverseArray reverses an array.
//
// MongoDB equivalent: { $reverseArray: expression }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/reverseArray/
func ExprReverseArray(expression interface{}) bson.D {
	return bson.D{{Key: "$reverseArray", Value: expression}}
}

// ExprSortArray sorts an array by the given sort specification.
//
// MongoDB equivalent: { $sortArray: { input: array, sortBy: spec } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/sortArray/
func ExprSortArray(input interface{}, sortBy interface{}) bson.D {
	return bson.D{{Key: "$sortArray", Value: bson.D{
		{Key: "input", Value: input},
		{Key: "sortBy", Value: sortBy},
	}}}
}

// ExprZip transposes an array of input arrays.
//
// MongoDB equivalent: { $zip: { inputs: [arr1, arr2], useLongestLength: true, defaults: [...] } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/zip/
func ExprZip(inputs bson.A, useLongestLength bool, defaults bson.A) bson.D {
	doc := bson.D{{Key: "inputs", Value: inputs}}
	if useLongestLength {
		doc = append(doc, bson.E{Key: "useLongestLength", Value: true})
	}
	if len(defaults) > 0 {
		doc = append(doc, bson.E{Key: "defaults", Value: defaults})
	}
	return bson.D{{Key: "$zip", Value: doc}}
}

// ExprObjectToArray converts a document to an array of key-value pairs.
//
// MongoDB equivalent: { $objectToArray: expression }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/objectToArray/
func ExprObjectToArray(expression interface{}) bson.D {
	return bson.D{{Key: "$objectToArray", Value: expression}}
}

// ExprArrayToObject converts an array of key-value pairs to a document.
//
// MongoDB equivalent: { $arrayToObject: expression }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/arrayToObject/
func ExprArrayToObject(expression interface{}) bson.D {
	return bson.D{{Key: "$arrayToObject", Value: expression}}
}

// --- Date Expression Operators ---
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/#date-expression-operators

// ExprDateFromString converts a date string to a Date object.
//
// MongoDB equivalent: { $dateFromString: { dateString: str, format: fmt, timezone: tz } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/dateFromString/
func ExprDateFromString(dateString interface{}, format, timezone interface{}) bson.D {
	doc := bson.D{{Key: "dateString", Value: dateString}}
	if format != nil {
		doc = append(doc, bson.E{Key: "format", Value: format})
	}
	if timezone != nil {
		doc = append(doc, bson.E{Key: "timezone", Value: timezone})
	}
	return bson.D{{Key: "$dateFromString", Value: doc}}
}

// ExprDateToString converts a Date to a formatted string.
//
// MongoDB equivalent: { $dateToString: { date: dateExpr, format: fmt, timezone: tz } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/dateToString/
func ExprDateToString(date interface{}, format, timezone interface{}) bson.D {
	doc := bson.D{{Key: "date", Value: date}}
	if format != nil {
		doc = append(doc, bson.E{Key: "format", Value: format})
	}
	if timezone != nil {
		doc = append(doc, bson.E{Key: "timezone", Value: timezone})
	}
	return bson.D{{Key: "$dateToString", Value: doc}}
}

// ExprDateAdd adds a specified amount of time to a date.
//
// MongoDB equivalent: { $dateAdd: { startDate: date, unit: "hour", amount: 3 } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/dateAdd/
func ExprDateAdd(startDate interface{}, unit string, amount interface{}) bson.D {
	return bson.D{{Key: "$dateAdd", Value: bson.D{
		{Key: "startDate", Value: startDate},
		{Key: "unit", Value: unit},
		{Key: "amount", Value: amount},
	}}}
}

// ExprDateSubtract subtracts a specified amount of time from a date.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/dateSubtract/
func ExprDateSubtract(startDate interface{}, unit string, amount interface{}) bson.D {
	return bson.D{{Key: "$dateSubtract", Value: bson.D{
		{Key: "startDate", Value: startDate},
		{Key: "unit", Value: unit},
		{Key: "amount", Value: amount},
	}}}
}

// ExprDateDiff returns the difference between two dates.
//
// MongoDB equivalent: { $dateDiff: { startDate: d1, endDate: d2, unit: "day" } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/dateDiff/
func ExprDateDiff(startDate, endDate interface{}, unit string) bson.D {
	return bson.D{{Key: "$dateDiff", Value: bson.D{
		{Key: "startDate", Value: startDate},
		{Key: "endDate", Value: endDate},
		{Key: "unit", Value: unit},
	}}}
}

// ExprDateTrunc truncates a date to a specified unit.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/dateTrunc/
func ExprDateTrunc(date interface{}, unit string) bson.D {
	return bson.D{{Key: "$dateTrunc", Value: bson.D{
		{Key: "date", Value: date},
		{Key: "unit", Value: unit},
	}}}
}

// ExprYear extracts the year from a date.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/year/
func ExprYear(date interface{}) bson.D { return bson.D{{Key: "$year", Value: date}} }

// ExprMonth extracts the month from a date (1-12).
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/month/
func ExprMonth(date interface{}) bson.D { return bson.D{{Key: "$month", Value: date}} }

// ExprDayOfMonth extracts the day of the month from a date (1-31).
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/dayOfMonth/
func ExprDayOfMonth(date interface{}) bson.D { return bson.D{{Key: "$dayOfMonth", Value: date}} }

// ExprHour extracts the hour from a date (0-23).
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/hour/
func ExprHour(date interface{}) bson.D { return bson.D{{Key: "$hour", Value: date}} }

// ExprMinute extracts the minute from a date (0-59).
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/minute/
func ExprMinute(date interface{}) bson.D { return bson.D{{Key: "$minute", Value: date}} }

// ExprSecond extracts the second from a date (0-59).
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/second/
func ExprSecond(date interface{}) bson.D { return bson.D{{Key: "$second", Value: date}} }

// ExprMillisecond extracts the millisecond from a date (0-999).
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/millisecond/
func ExprMillisecond(date interface{}) bson.D { return bson.D{{Key: "$millisecond", Value: date}} }

// ExprDayOfWeek extracts the day of the week (1=Sunday, 7=Saturday).
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/dayOfWeek/
func ExprDayOfWeek(date interface{}) bson.D { return bson.D{{Key: "$dayOfWeek", Value: date}} }

// ExprDayOfYear extracts the day of the year (1-366).
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/dayOfYear/
func ExprDayOfYear(date interface{}) bson.D { return bson.D{{Key: "$dayOfYear", Value: date}} }

// ExprISOWeek returns the ISO 8601 week number (1-53).
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/isoWeek/
func ExprISOWeek(date interface{}) bson.D { return bson.D{{Key: "$isoWeek", Value: date}} }

// ExprISOWeekYear returns the ISO 8601 week-numbering year.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/isoWeekYear/
func ExprISOWeekYear(date interface{}) bson.D { return bson.D{{Key: "$isoWeekYear", Value: date}} }

// --- Type Expression Operators ---
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/#type-expression-operators

// ExprConvert converts a value to a specified type.
//
// MongoDB equivalent: { $convert: { input: expr, to: type, onError: errExpr, onNull: nullExpr } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/convert/
func ExprConvert(input interface{}, to interface{}, onError, onNull interface{}) bson.D {
	doc := bson.D{
		{Key: "input", Value: input},
		{Key: "to", Value: to},
	}
	if onError != nil {
		doc = append(doc, bson.E{Key: "onError", Value: onError})
	}
	if onNull != nil {
		doc = append(doc, bson.E{Key: "onNull", Value: onNull})
	}
	return bson.D{{Key: "$convert", Value: doc}}
}

// ExprToBool converts to boolean. See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/toBool/
func ExprToBool(expr interface{}) bson.D { return bson.D{{Key: "$toBool", Value: expr}} }

// ExprToInt converts to integer. See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/toInt/
func ExprToInt(expr interface{}) bson.D { return bson.D{{Key: "$toInt", Value: expr}} }

// ExprToLong converts to long. See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/toLong/
func ExprToLong(expr interface{}) bson.D { return bson.D{{Key: "$toLong", Value: expr}} }

// ExprToDouble converts to double. See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/toDouble/
func ExprToDouble(expr interface{}) bson.D { return bson.D{{Key: "$toDouble", Value: expr}} }

// ExprToDecimal converts to decimal. See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/toDecimal/
func ExprToDecimal(expr interface{}) bson.D { return bson.D{{Key: "$toDecimal", Value: expr}} }

// ExprToString converts to string. See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/toString/
func ExprToString(expr interface{}) bson.D { return bson.D{{Key: "$toString", Value: expr}} }

// ExprToObjectId converts to ObjectId. See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/toObjectId/
func ExprToObjectId(expr interface{}) bson.D { return bson.D{{Key: "$toObjectId", Value: expr}} }

// ExprToDate converts to Date. See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/toDate/
func ExprToDate(expr interface{}) bson.D { return bson.D{{Key: "$toDate", Value: expr}} }

// ExprType returns the BSON type of a value. See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/type/
func ExprType(expr interface{}) bson.D { return bson.D{{Key: "$type", Value: expr}} }

// ExprIsNumber returns true if the expression is numeric. See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/isNumber/
func ExprIsNumber(expr interface{}) bson.D { return bson.D{{Key: "$isNumber", Value: expr}} }
