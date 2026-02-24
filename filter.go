package gmqb

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Filter represents an immutable MongoDB query predicate.
// Use method chaining to compose multiple conditions — each method returns
// a new Filter instance, leaving the original unchanged. Chained conditions
// are implicitly ANDed by MongoDB.
//
// Use the output methods BsonD(), BsonM(), JSON(), or CompactJSON() to extract
// the query in the format needed by the go-mongodb-driver or for debugging.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/
//
// Example:
//
//	filter := gmqb.NewFilter().
//	    Eq("status", "active").
//	    Gte("age", 18).
//	    Exists("email", true)
//	cursor, err := coll.Find(ctx, filter.BsonD())
type Filter struct {
	d bson.D
}

// NewFilter creates an empty Filter ready for chaining.
//
// Example:
//
//	filter := gmqb.NewFilter().Eq("name", "Alice").Gt("age", 18)
func NewFilter() Filter {
	return Filter{}
}

// append returns a new Filter with an additional predicate element appended.
// The original filter is never mutated.
func (f Filter) append(e bson.E) Filter {
	newD := make(bson.D, len(f.d), len(f.d)+1)
	copy(newD, f.d)
	newD = append(newD, e)
	return Filter{d: newD}
}

// BsonD returns the filter as a bson.D, suitable for passing directly to
// mongo.Collection.Find(), FindOne(), DeleteMany(), etc.
//
// Example:
//
//	filter := gmqb.Eq("name", "Alice")
//	cursor, err := coll.Find(ctx, filter.BsonD())
func (f Filter) BsonD() bson.D {
	return f.d
}

// BsonM returns the filter as a bson.M (unordered map).
//
// Example:
//
//	filter := gmqb.Eq("name", "Alice")
//	m := filter.BsonM() // bson.M{"name": bson.M{"$eq": "Alice"}}
func (f Filter) BsonM() bson.M {
	m := bson.M{}
	for _, e := range f.d {
		m[e.Key] = e.Value
	}
	return m
}

// JSON returns the filter as a pretty-printed JSON string.
// Useful for debugging and logging.
//
// Example:
//
//	filter := gmqb.Eq("name", "Alice")
//	fmt.Println(filter.JSON())
//	// {
//	//   "name": { "$eq": "Alice" }
//	// }
func (f Filter) JSON() string {
	return toJSON(f.d)
}

// CompactJSON returns the filter as a compact JSON string with no whitespace.
//
// Example:
//
//	filter := gmqb.Eq("name", "Alice")
//	fmt.Println(filter.CompactJSON())
//	// {"name":{"$eq":"Alice"}}
func (f Filter) CompactJSON() string {
	return toCompactJSON(f.d)
}

// IsEmpty returns true if the filter contains no predicates.
func (f Filter) IsEmpty() bool {
	return len(f.d) == 0
}

// Raw creates a Filter from a raw bson.D. Use this for operators not yet
// supported by the builder, or for passing pre-built queries.
//
// Example:
//
//	filter := gmqb.Raw(bson.D{{"$text", bson.D{{"$search", "coffee"}}}})
func Raw(d bson.D) Filter {
	return Filter{d: d}
}

// --- Comparison Operators ---

// Eq creates a filter that matches documents where the field equals the specified value.
//
// MongoDB equivalent:
//
//	{ field: { $eq: value } }
//
// If value is a simple type, this is equivalent to { field: value }.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/eq/
//
// Example:
//
//	filter := gmqb.Eq("age", 21)
//	fmt.Println(filter.JSON())
//	// {"age": {"$eq": 21}}
func Eq(field string, value interface{}) Filter {
	return Filter{d: bson.D{{Key: field, Value: bson.D{{Key: "$eq", Value: value}}}}}
}

// Eq chains an $eq condition onto the filter. Matches documents where the field
// equals the specified value. Multiple chained conditions are implicitly ANDed.
//
// MongoDB equivalent:
//
//	{ field: { $eq: value } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/eq/
//
// Example:
//
//	filter := gmqb.NewFilter().Eq("name", "Alice").Eq("active", true)
func (f Filter) Eq(field string, value interface{}) Filter {
	return f.append(bson.E{Key: field, Value: bson.D{{Key: "$eq", Value: value}}})
}

// Ne creates a filter that matches documents where the field is not equal to the specified value.
//
// MongoDB equivalent:
//
//	{ field: { $ne: value } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/ne/
//
// Example:
//
//	filter := gmqb.Ne("status", "archived")
//	fmt.Println(filter.JSON())
//	// {"status": {"$ne": "archived"}}
func Ne(field string, value interface{}) Filter {
	return Filter{d: bson.D{{Key: field, Value: bson.D{{Key: "$ne", Value: value}}}}}
}

// Ne chains a $ne condition onto the filter. Matches documents where the field
// is not equal to the specified value.
//
// MongoDB equivalent:
//
//	{ field: { $ne: value } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/ne/
//
// Example:
//
//	filter := gmqb.NewFilter().Ne("status", "archived").Eq("active", true)
func (f Filter) Ne(field string, value interface{}) Filter {
	return f.append(bson.E{Key: field, Value: bson.D{{Key: "$ne", Value: value}}})
}

// Gt creates a filter that matches documents where the field is greater than the specified value.
//
// MongoDB equivalent:
//
//	{ field: { $gt: value } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/gt/
//
// Example:
//
//	filter := gmqb.Gt("age", 18)
//	fmt.Println(filter.JSON())
//	// {"age": {"$gt": 18}}
func Gt(field string, value interface{}) Filter {
	return Filter{d: bson.D{{Key: field, Value: bson.D{{Key: "$gt", Value: value}}}}}
}

// Gt chains a $gt condition onto the filter. Matches documents where the field
// is greater than the specified value.
//
// MongoDB equivalent:
//
//	{ field: { $gt: value } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/gt/
//
// Example:
//
//	filter := gmqb.NewFilter().Gt("age", 18).Lt("age", 65)
func (f Filter) Gt(field string, value interface{}) Filter {
	return f.append(bson.E{Key: field, Value: bson.D{{Key: "$gt", Value: value}}})
}

// Gte creates a filter that matches documents where the field is greater than or equal
// to the specified value.
//
// MongoDB equivalent:
//
//	{ field: { $gte: value } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/gte/
//
// Example:
//
//	filter := gmqb.Gte("age", 18)
//	fmt.Println(filter.JSON())
//	// {"age": {"$gte": 18}}
func Gte(field string, value interface{}) Filter {
	return Filter{d: bson.D{{Key: field, Value: bson.D{{Key: "$gte", Value: value}}}}}
}

// Gte chains a $gte condition onto the filter. Matches documents where the field
// is greater than or equal to the specified value.
//
// MongoDB equivalent:
//
//	{ field: { $gte: value } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/gte/
//
// Example:
//
//	filter := gmqb.NewFilter().Gte("age", 18).Lte("age", 65)
func (f Filter) Gte(field string, value interface{}) Filter {
	return f.append(bson.E{Key: field, Value: bson.D{{Key: "$gte", Value: value}}})
}

// Lt creates a filter that matches documents where the field is less than the specified value.
//
// MongoDB equivalent:
//
//	{ field: { $lt: value } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/lt/
//
// Example:
//
//	filter := gmqb.Lt("price", 100.0)
//	fmt.Println(filter.JSON())
//	// {"price": {"$lt": 100.0}}
func Lt(field string, value interface{}) Filter {
	return Filter{d: bson.D{{Key: field, Value: bson.D{{Key: "$lt", Value: value}}}}}
}

// Lt chains a $lt condition onto the filter. Matches documents where the field
// is less than the specified value.
//
// MongoDB equivalent:
//
//	{ field: { $lt: value } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/lt/
//
// Example:
//
//	filter := gmqb.NewFilter().Gte("price", 10).Lt("price", 100)
func (f Filter) Lt(field string, value interface{}) Filter {
	return f.append(bson.E{Key: field, Value: bson.D{{Key: "$lt", Value: value}}})
}

// Lte creates a filter that matches documents where the field is less than or equal
// to the specified value.
//
// MongoDB equivalent:
//
//	{ field: { $lte: value } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/lte/
//
// Example:
//
//	filter := gmqb.Lte("quantity", 50)
//	fmt.Println(filter.JSON())
//	// {"quantity": {"$lte": 50}}
func Lte(field string, value interface{}) Filter {
	return Filter{d: bson.D{{Key: field, Value: bson.D{{Key: "$lte", Value: value}}}}}
}

// Lte chains a $lte condition onto the filter. Matches documents where the field
// is less than or equal to the specified value.
//
// MongoDB equivalent:
//
//	{ field: { $lte: value } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/lte/
//
// Example:
//
//	filter := gmqb.NewFilter().Gte("qty", 10).Lte("qty", 50)
func (f Filter) Lte(field string, value interface{}) Filter {
	return f.append(bson.E{Key: field, Value: bson.D{{Key: "$lte", Value: value}}})
}

// In creates a filter that matches documents where the field value is in the given list.
//
// MongoDB equivalent:
//
//	{ field: { $in: [value1, value2, ...] } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/in/
//
// Example:
//
//	filter := gmqb.In("status", "active", "pending")
//	fmt.Println(filter.JSON())
//	// {"status": {"$in": ["active", "pending"]}}
func In(field string, values ...interface{}) Filter {
	return Filter{d: bson.D{{Key: field, Value: bson.D{{Key: "$in", Value: bson.A(values)}}}}}
}

// In chains an $in condition onto the filter. Matches documents where the field
// value is in the given list.
//
// MongoDB equivalent:
//
//	{ field: { $in: [value1, value2, ...] } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/in/
//
// Example:
//
//	filter := gmqb.NewFilter().In("status", "active", "pending").Eq("country", "US")
func (f Filter) In(field string, values ...interface{}) Filter {
	return f.append(bson.E{Key: field, Value: bson.D{{Key: "$in", Value: bson.A(values)}}})
}

// Nin creates a filter that matches documents where the field value is not in the given list.
//
// MongoDB equivalent:
//
//	{ field: { $nin: [value1, value2, ...] } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/nin/
//
// Example:
//
//	filter := gmqb.Nin("role", "banned", "suspended")
//	fmt.Println(filter.JSON())
//	// {"role": {"$nin": ["banned", "suspended"]}}
func Nin(field string, values ...interface{}) Filter {
	return Filter{d: bson.D{{Key: field, Value: bson.D{{Key: "$nin", Value: bson.A(values)}}}}}
}

// Nin chains a $nin condition onto the filter. Matches documents where the field
// value is not in the given list.
//
// MongoDB equivalent:
//
//	{ field: { $nin: [value1, value2, ...] } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/nin/
//
// Example:
//
//	filter := gmqb.NewFilter().Nin("role", "banned", "suspended").Eq("active", true)
func (f Filter) Nin(field string, values ...interface{}) Filter {
	return f.append(bson.E{Key: field, Value: bson.D{{Key: "$nin", Value: bson.A(values)}}})
}

// --- Logical Operators ---

// And joins multiple filters with a logical AND. Returns documents that match
// all of the specified conditions.
//
// MongoDB equivalent:
//
//	{ $and: [ filter1, filter2, ... ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/and/
//
// Example:
//
//	filter := gmqb.And(
//	    gmqb.Gte("age", 18),
//	    gmqb.Lt("age", 65),
//	)
//	fmt.Println(filter.JSON())
//	// {"$and": [{"age": {"$gte": 18}}, {"age": {"$lt": 65}}]}
func And(filters ...Filter) Filter {
	arr := make(bson.A, len(filters))
	for i, f := range filters {
		arr[i] = f.d
	}
	return Filter{d: bson.D{{Key: "$and", Value: arr}}}
}

// Or joins multiple filters with a logical OR. Returns documents that match
// at least one of the specified conditions.
//
// MongoDB equivalent:
//
//	{ $or: [ filter1, filter2, ... ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/or/
//
// Example:
//
//	filter := gmqb.Or(
//	    gmqb.Eq("status", "active"),
//	    gmqb.Eq("status", "pending"),
//	)
func Or(filters ...Filter) Filter {
	arr := make(bson.A, len(filters))
	for i, f := range filters {
		arr[i] = f.d
	}
	return Filter{d: bson.D{{Key: "$or", Value: arr}}}
}

// Nor joins multiple filters with a logical NOR. Returns documents that fail
// to match all of the specified conditions.
//
// MongoDB equivalent:
//
//	{ $nor: [ filter1, filter2, ... ] }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/nor/
//
// Example:
//
//	filter := gmqb.Nor(
//	    gmqb.Eq("status", "archived"),
//	    gmqb.Lt("age", 18),
//	)
func Nor(filters ...Filter) Filter {
	arr := make(bson.A, len(filters))
	for i, f := range filters {
		arr[i] = f.d
	}
	return Filter{d: bson.D{{Key: "$nor", Value: arr}}}
}

// Not inverts a filter expression for a specific field. Returns documents that do not
// match the filter expression.
//
// MongoDB equivalent:
//
//	{ field: { $not: { operator-expression } } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/not/
//
// Example:
//
//	filter := gmqb.Not("age", gmqb.Gte("age", 18))
//	// {"age": {"$not": {"$gte": 18}}}
func Not(field string, inner Filter) Filter {
	// Extract the operator expression from the inner filter's field
	var opExpr interface{}
	for _, e := range inner.d {
		if e.Key == field {
			opExpr = e.Value
			break
		}
	}
	if opExpr == nil {
		opExpr = inner.d
	}
	return Filter{d: bson.D{{Key: field, Value: bson.D{{Key: "$not", Value: opExpr}}}}}
}

// --- Element Operators ---

// Exists matches documents that have (or do not have) the specified field.
//
// MongoDB equivalent:
//
//	{ field: { $exists: true/false } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/exists/
//
// Example:
//
//	filter := gmqb.Exists("email", true)
//	fmt.Println(filter.JSON())
//	// {"email": {"$exists": true}}
func Exists(field string, exists bool) Filter {
	return Filter{d: bson.D{{Key: field, Value: bson.D{{Key: "$exists", Value: exists}}}}}
}

// Exists chains an $exists condition onto the filter. Matches documents that have
// (or do not have) the specified field.
//
// MongoDB equivalent:
//
//	{ field: { $exists: true/false } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/exists/
//
// Example:
//
//	filter := gmqb.NewFilter().Exists("email", true).Eq("active", true)
func (f Filter) Exists(field string, exists bool) Filter {
	return f.append(bson.E{Key: field, Value: bson.D{{Key: "$exists", Value: exists}}})
}

// Type matches documents where the field is of the specified BSON type.
// The typeVal can be a string type alias (e.g. "string", "int", "double") or
// a numeric BSON type code.
//
// MongoDB equivalent:
//
//	{ field: { $type: typeVal } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/type/
//
// Example:
//
//	filter := gmqb.Type("age", "int")
//	fmt.Println(filter.JSON())
//	// {"age": {"$type": "int"}}
func Type(field string, typeVal interface{}) Filter {
	return Filter{d: bson.D{{Key: field, Value: bson.D{{Key: "$type", Value: typeVal}}}}}
}

// Type chains a $type condition onto the filter. Matches documents where the field
// is of the specified BSON type. The typeVal can be a string type alias
// (e.g. "string", "int", "double") or a numeric BSON type code.
//
// MongoDB equivalent:
//
//	{ field: { $type: typeVal } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/type/
//
// Example:
//
//	filter := gmqb.NewFilter().Type("age", "int").Exists("age", true)
func (f Filter) Type(field string, typeVal interface{}) Filter {
	return f.append(bson.E{Key: field, Value: bson.D{{Key: "$type", Value: typeVal}}})
}

// --- Evaluation Operators ---

// Mod matches documents where a field value divided by a divisor has the specified remainder.
//
// MongoDB equivalent:
//
//	{ field: { $mod: [ divisor, remainder ] } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/mod/
//
// Example:
//
//	filter := gmqb.Mod("qty", 4, 0) // qty divisible by 4
//	fmt.Println(filter.JSON())
//	// {"qty": {"$mod": [4, 0]}}
func Mod(field string, divisor, remainder int64) Filter {
	return Filter{d: bson.D{{Key: field, Value: bson.D{{Key: "$mod", Value: bson.A{divisor, remainder}}}}}}
}

// Regex matches documents where the field value matches the specified regular expression.
//
// MongoDB equivalent:
//
//	{ field: { $regex: pattern, $options: options } }
//
// Common options: "i" (case-insensitive), "m" (multiline), "x" (extended), "s" (dot matches newline).
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/regex/
//
// Example:
//
//	filter := gmqb.Regex("email", `^.*@company\.com$`, "i")
//	fmt.Println(filter.JSON())
//	// {"email": {"$regex": "^.*@company\\.com$", "$options": "i"}}
func Regex(field string, pattern string, options string) Filter {
	expr := bson.D{{Key: "$regex", Value: pattern}}
	if options != "" {
		expr = append(expr, bson.E{Key: "$options", Value: options})
	}
	return Filter{d: bson.D{{Key: field, Value: expr}}}
}

// Regex chains a $regex condition onto the filter. Matches documents where the field
// value matches the specified regular expression.
//
// Common options: "i" (case-insensitive), "m" (multiline), "x" (extended), "s" (dot matches newline).
//
// MongoDB equivalent:
//
//	{ field: { $regex: pattern, $options: options } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/regex/
//
// Example:
//
//	filter := gmqb.NewFilter().Regex("email", `^admin`, "i").Eq("active", true)
func (f Filter) Regex(field string, pattern string, options string) Filter {
	expr := bson.D{{Key: "$regex", Value: pattern}}
	if options != "" {
		expr = append(expr, bson.E{Key: "$options", Value: options})
	}
	return f.append(bson.E{Key: field, Value: expr})
}

// Expr allows the use of aggregation expressions within query predicates.
// This enables comparisons between fields within the same document.
//
// MongoDB equivalent:
//
//	{ $expr: expression }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/expr/
//
// Example:
//
//	// Match where spent > budget
//	filter := gmqb.Expr(bson.D{{"$gt", bson.A{"$spent", "$budget"}}})
func Expr(expression interface{}) Filter {
	return Filter{d: bson.D{{Key: "$expr", Value: expression}}}
}

// Where matches documents that satisfy a JavaScript expression.
// The JS function has access to the document as "this".
//
// MongoDB equivalent:
//
//	{ $where: "javascript expression" }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/where/
//
// Example:
//
//	filter := gmqb.Where("this.credits > this.debits")
func Where(jsExpr string) Filter {
	return Filter{d: bson.D{{Key: "$where", Value: jsExpr}}}
}

// JsonSchema validates documents against the given JSON Schema.
//
// MongoDB equivalent:
//
//	{ $jsonSchema: schema }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/jsonSchema/
//
// Example:
//
//	filter := gmqb.JsonSchema(bson.D{
//	    {"required", bson.A{"name", "email"}},
//	    {"properties", bson.D{
//	        {"name", bson.D{{"bsonType", "string"}}},
//	    }},
//	})
func JsonSchema(schema interface{}) Filter {
	return Filter{d: bson.D{{Key: "$jsonSchema", Value: schema}}}
}

// --- Array Operators ---

// All matches arrays that contain all the specified elements.
//
// MongoDB equivalent:
//
//	{ field: { $all: [value1, value2, ...] } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/all/
//
// Example:
//
//	filter := gmqb.All("tags", "ssl", "security")
//	fmt.Println(filter.JSON())
//	// {"tags": {"$all": ["ssl", "security"]}}
func All(field string, values ...interface{}) Filter {
	return Filter{d: bson.D{{Key: field, Value: bson.D{{Key: "$all", Value: bson.A(values)}}}}}
}

// ElemMatch selects documents if at least one element in the array field matches all
// the specified conditions.
//
// MongoDB equivalent:
//
//	{ field: { $elemMatch: { condition1, condition2, ... } } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/elemMatch/
//
// Example:
//
//	filter := gmqb.ElemMatch("results", gmqb.And(
//	    gmqb.Gte("score", 80),
//	    gmqb.Lt("score", 100),
//	))
func ElemMatch(field string, filter Filter) Filter {
	return Filter{d: bson.D{{Key: field, Value: bson.D{{Key: "$elemMatch", Value: filter.d}}}}}
}

// Size selects documents where the array field contains the specified number of elements.
//
// MongoDB equivalent:
//
//	{ field: { $size: n } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/size/
//
// Example:
//
//	filter := gmqb.Size("tags", 3)
//	fmt.Println(filter.JSON())
//	// {"tags": {"$size": 3}}
func Size(field string, n int) Filter {
	return Filter{d: bson.D{{Key: field, Value: bson.D{{Key: "$size", Value: n}}}}}
}

// Size chains a $size condition onto the filter. Matches documents where the array
// field contains exactly n elements.
//
// MongoDB equivalent:
//
//	{ field: { $size: n } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/size/
//
// Example:
//
//	filter := gmqb.NewFilter().Size("tags", 3).Exists("tags", true)
func (f Filter) Size(field string, n int) Filter {
	return f.append(bson.E{Key: field, Value: bson.D{{Key: "$size", Value: n}}})
}

// --- Geospatial Geometry Helpers ---

// Point creates a GeoJSON Point object for use in geospatial queries.
// The coordinates are specified as longitude, then latitude.
func Point(longitude, latitude float64) bson.D {
	return bson.D{
		{Key: "type", Value: "Point"},
		{Key: "coordinates", Value: bson.A{longitude, latitude}},
	}
}

// LineString creates a GeoJSON LineString object for use in geospatial queries.
// coordinates is a variadic list of [longitude, latitude] pairs.
func LineString(coordinates ...[2]float64) bson.D {
	coords := make(bson.A, len(coordinates))
	for i, c := range coordinates {
		coords[i] = bson.A{c[0], c[1]}
	}
	return bson.D{
		{Key: "type", Value: "LineString"},
		{Key: "coordinates", Value: coords},
	}
}

// Polygon creates a GeoJSON Polygon object for use in geospatial queries.
// A Polygon is defined by one or more linear rings. The first ring must be the
// exterior boundary, and any subsequent rings denote holes.
// Each ring must be closed (the first and last coordinate pair must be the same).
func Polygon(rings ...[][2]float64) bson.D {
	coords := make(bson.A, len(rings))
	for i, ring := range rings {
		ringCoords := make(bson.A, len(ring))
		for j, c := range ring {
			ringCoords[j] = bson.A{c[0], c[1]}
		}
		coords[i] = ringCoords
	}
	return bson.D{
		{Key: "type", Value: "Polygon"},
		{Key: "coordinates", Value: coords},
	}
}

// --- Geospatial Operators ---

// GeoIntersects selects documents whose geospatial data intersects with a GeoJSON geometry.
// The geometry parameter should be a GeoJSON object (bson.D or bson.M).
//
// MongoDB equivalent:
//
//	{ field: { $geoIntersects: { $geometry: geometry } } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/geoIntersects/
//
// Example:
//
//	filter := gmqb.GeoIntersects("location", gmqb.Polygon([][2]float64{
//	    {0, 0}, {3, 6}, {6, 1}, {0, 0},
//	}))
func GeoIntersects(field string, geometry interface{}) Filter {
	return Filter{d: bson.D{{Key: field, Value: bson.D{
		{Key: "$geoIntersects", Value: bson.D{{Key: "$geometry", Value: geometry}}},
	}}}}
}

// GeoWithin selects documents whose geospatial data exists entirely within a shape.
// The geometry parameter should be a GeoJSON object.
//
// MongoDB equivalent:
//
//	{ field: { $geoWithin: { $geometry: geometry } } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/geoWithin/
//
// Example:
//
//	filter := gmqb.GeoWithin("location", gmqb.Polygon([][2]float64{
//	    {0, 0}, {3, 6}, {6, 1}, {0, 0},
//	}))
func GeoWithin(field string, geometry interface{}) Filter {
	return Filter{d: bson.D{{Key: field, Value: bson.D{
		{Key: "$geoWithin", Value: bson.D{{Key: "$geometry", Value: geometry}}},
	}}}}
}

// Near returns documents sorted by proximity to a GeoJSON point.
// maxDistance and minDistance are in meters. Pass 0 to omit either.
//
// MongoDB equivalent:
//
//	{ field: { $near: { $geometry: point, $maxDistance: m, $minDistance: m } } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/near/
//
// Example:
//
//	filter := gmqb.Near("location", gmqb.Point(-73.9667, 40.78), 1000, 0)
func Near(field string, geometry interface{}, maxDistance, minDistance float64) Filter {
	nearDoc := bson.D{{Key: "$geometry", Value: geometry}}
	if maxDistance > 0 {
		nearDoc = append(nearDoc, bson.E{Key: "$maxDistance", Value: maxDistance})
	}
	if minDistance > 0 {
		nearDoc = append(nearDoc, bson.E{Key: "$minDistance", Value: minDistance})
	}
	return Filter{d: bson.D{{Key: field, Value: bson.D{{Key: "$near", Value: nearDoc}}}}}
}

// NearSphere returns documents sorted by proximity to a point on a sphere.
// maxDistance and minDistance are in meters. Pass 0 to omit either.
//
// MongoDB equivalent:
//
//	{ field: { $nearSphere: { $geometry: point, $maxDistance: m, $minDistance: m } } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/nearSphere/
//
// Example:
//
//	filter := gmqb.NearSphere("location", gmqb.Point(-73.9667, 40.78), 5000, 100)
func NearSphere(field string, geometry interface{}, maxDistance, minDistance float64) Filter {
	nearDoc := bson.D{{Key: "$geometry", Value: geometry}}
	if maxDistance > 0 {
		nearDoc = append(nearDoc, bson.E{Key: "$maxDistance", Value: maxDistance})
	}
	if minDistance > 0 {
		nearDoc = append(nearDoc, bson.E{Key: "$minDistance", Value: minDistance})
	}
	return Filter{d: bson.D{{Key: field, Value: bson.D{{Key: "$nearSphere", Value: nearDoc}}}}}
}

// --- Bitwise Operators ---

// BitsAllClear matches documents where all of the specified bit positions are clear (0).
// bitmask can be a numeric value, a BinData value, or a position list (bson.A).
//
// MongoDB equivalent:
//
//	{ field: { $bitsAllClear: bitmask } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/bitsAllClear/
//
// Example:
//
//	filter := gmqb.BitsAllClear("flags", 35) // binary: 100011
func BitsAllClear(field string, bitmask interface{}) Filter {
	return Filter{d: bson.D{{Key: field, Value: bson.D{{Key: "$bitsAllClear", Value: bitmask}}}}}
}

// BitsAllSet matches documents where all of the specified bit positions are set (1).
//
// MongoDB equivalent:
//
//	{ field: { $bitsAllSet: bitmask } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/bitsAllSet/
//
// Example:
//
//	filter := gmqb.BitsAllSet("flags", 50) // binary: 110010
func BitsAllSet(field string, bitmask interface{}) Filter {
	return Filter{d: bson.D{{Key: field, Value: bson.D{{Key: "$bitsAllSet", Value: bitmask}}}}}
}

// BitsAnyClear matches documents where any of the specified bit positions are clear (0).
//
// MongoDB equivalent:
//
//	{ field: { $bitsAnyClear: bitmask } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/bitsAnyClear/
//
// Example:
//
//	filter := gmqb.BitsAnyClear("flags", 35)
func BitsAnyClear(field string, bitmask interface{}) Filter {
	return Filter{d: bson.D{{Key: field, Value: bson.D{{Key: "$bitsAnyClear", Value: bitmask}}}}}
}

// BitsAnySet matches documents where any of the specified bit positions are set (1).
//
// MongoDB equivalent:
//
//	{ field: { $bitsAnySet: bitmask } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/query/bitsAnySet/
//
// Example:
//
//	filter := gmqb.BitsAnySet("flags", 50)
func BitsAnySet(field string, bitmask interface{}) Filter {
	return Filter{d: bson.D{{Key: field, Value: bson.D{{Key: "$bitsAnySet", Value: bitmask}}}}}
}
