package gmqb

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Updater represents an immutable MongoDB update document.
// Use method chaining to compose multiple update operations — each method returns
// a new Updater instance, leaving the original unchanged.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/update/
//
// Example:
//
//	update := gmqb.NewUpdate().
//	    Set("name", "Bob").
//	    Inc("age", 1).
//	    Push("tags", "verified")
//	bsonDoc := update.BsonD()
type Updater struct {
	ops bson.D
}

// NewUpdate creates an empty Updater ready for chaining.
//
// Example:
//
//	update := gmqb.NewUpdate().Set("name", "Alice")
func NewUpdate() Updater {
	return Updater{}
}

// BsonD returns the update document as a bson.D, suitable for passing directly
// to mongo.Collection.UpdateOne(), UpdateMany(), etc.
func (u Updater) BsonD() bson.D {
	return u.ops
}

// JSON returns the update document as a pretty-printed JSON string.
func (u Updater) JSON() string {
	return toJSON(u.ops)
}

// CompactJSON returns the update document as a compact JSON string.
func (u Updater) CompactJSON() string {
	return toCompactJSON(u.ops)
}

// IsEmpty returns true if no update operations have been added.
func (u Updater) IsEmpty() bool {
	return len(u.ops) == 0
}

// addOp adds or merges an operator entry into the update document.
// If the operator already exists, the new field is appended to the existing sub-document.
func (u Updater) addOp(op string, field string, value interface{}) Updater {
	newOps := make(bson.D, len(u.ops))
	copy(newOps, u.ops)

	for i, e := range newOps {
		if e.Key == op {
			existing := e.Value.(bson.D)
			merged := make(bson.D, len(existing), len(existing)+1)
			copy(merged, existing)
			merged = append(merged, bson.E{Key: field, Value: value})
			newOps[i] = bson.E{Key: op, Value: merged}
			return Updater{ops: newOps}
		}
	}

	newOps = append(newOps, bson.E{Key: op, Value: bson.D{{Key: field, Value: value}}})
	return Updater{ops: newOps}
}

// --- Field Update Operators ---

// Set sets the value of a field in a document. If the field does not exist, it is created.
//
// MongoDB equivalent:
//
//	{ $set: { field: value } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/update/set/
//
// Example:
//
//	update := gmqb.NewUpdate().Set("name", "Alice").Set("age", 30)
//	fmt.Println(update.JSON())
//	// {"$set": {"name": "Alice", "age": 30}}
func (u Updater) Set(field string, value interface{}) Updater {
	return u.addOp("$set", field, value)
}

// Unset removes the specified field from a document.
//
// MongoDB equivalent:
//
//	{ $unset: { field: "" } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/update/unset/
//
// Example:
//
//	update := gmqb.NewUpdate().Unset("obsoleteField")
func (u Updater) Unset(field string) Updater {
	return u.addOp("$unset", field, "")
}

// Inc increments the value of a field by the specified amount. If the field does not
// exist, it is created with the increment value. Use a negative value to decrement.
//
// MongoDB equivalent:
//
//	{ $inc: { field: amount } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/update/inc/
//
// Example:
//
//	update := gmqb.NewUpdate().Inc("views", 1).Inc("stock", -5)
func (u Updater) Inc(field string, amount interface{}) Updater {
	return u.addOp("$inc", field, amount)
}

// Mul multiplies the value of a field by the specified amount.
//
// MongoDB equivalent:
//
//	{ $mul: { field: number } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/update/mul/
//
// Example:
//
//	update := gmqb.NewUpdate().Mul("price", 1.1) // 10% price increase
func (u Updater) Mul(field string, number interface{}) Updater {
	return u.addOp("$mul", field, number)
}

// Min only updates the field if the specified value is less than the existing value.
//
// MongoDB equivalent:
//
//	{ $min: { field: value } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/update/min/
//
// Example:
//
//	update := gmqb.NewUpdate().Min("lowScore", 50)
func (u Updater) Min(field string, value interface{}) Updater {
	return u.addOp("$min", field, value)
}

// Max only updates the field if the specified value is greater than the existing value.
//
// MongoDB equivalent:
//
//	{ $max: { field: value } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/update/max/
//
// Example:
//
//	update := gmqb.NewUpdate().Max("highScore", 950)
func (u Updater) Max(field string, value interface{}) Updater {
	return u.addOp("$max", field, value)
}

// Rename renames a field.
//
// MongoDB equivalent:
//
//	{ $rename: { oldName: newName } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/update/rename/
//
// Example:
//
//	update := gmqb.NewUpdate().Rename("nmae", "name")
func (u Updater) Rename(oldName, newName string) Updater {
	return u.addOp("$rename", oldName, newName)
}

// CurrentDate sets the value of a field to the current date as a Date type.
//
// MongoDB equivalent:
//
//	{ $currentDate: { field: true } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/update/currentDate/
//
// Example:
//
//	update := gmqb.NewUpdate().CurrentDate("lastModified")
func (u Updater) CurrentDate(field string) Updater {
	return u.addOp("$currentDate", field, true)
}

// CurrentDateAsTimestamp sets the value of a field to the current date as a Timestamp type.
//
// MongoDB equivalent:
//
//	{ $currentDate: { field: { $type: "timestamp" } } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/update/currentDate/
//
// Example:
//
//	update := gmqb.NewUpdate().CurrentDateAsTimestamp("lastModified")
func (u Updater) CurrentDateAsTimestamp(field string) Updater {
	return u.addOp("$currentDate", field, bson.D{{Key: "$type", Value: "timestamp"}})
}

// SetOnInsert sets the value of a field only when the update results in an insert
// (i.e. upsert is true and no matching document exists). Has no effect on update
// operations that modify existing documents.
//
// MongoDB equivalent:
//
//	{ $setOnInsert: { field: value } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/update/setOnInsert/
//
// Example:
//
//	update := gmqb.NewUpdate().
//	    Set("status", "active").
//	    SetOnInsert("createdAt", time.Now())
func (u Updater) SetOnInsert(field string, value interface{}) Updater {
	return u.addOp("$setOnInsert", field, value)
}

// --- Array Update Operators ---

// AddToSet adds a value to an array only if the value does not already exist in the array.
//
// MongoDB equivalent:
//
//	{ $addToSet: { field: value } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/update/addToSet/
//
// Example:
//
//	update := gmqb.NewUpdate().AddToSet("tags", "unique-tag")
func (u Updater) AddToSet(field string, value interface{}) Updater {
	return u.addOp("$addToSet", field, value)
}

// AddToSetEach adds each element to the array if it doesn't already exist. This is the
// $addToSet operator combined with the $each modifier.
//
// MongoDB equivalent:
//
//	{ $addToSet: { field: { $each: [values...] } } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/update/addToSet/
//
// Example:
//
//	update := gmqb.NewUpdate().AddToSetEach("tags", "tag1", "tag2", "tag3")
func (u Updater) AddToSetEach(field string, values ...interface{}) Updater {
	return u.addOp("$addToSet", field, bson.D{{Key: "$each", Value: bson.A(values)}})
}

// Pop removes the first or last element of an array. Use 1 to remove the last element,
// -1 to remove the first element.
//
// MongoDB equivalent:
//
//	{ $pop: { field: 1 } }  // remove last
//	{ $pop: { field: -1 } } // remove first
//
// See: https://www.mongodb.com/docs/manual/reference/operator/update/pop/
//
// Example:
//
//	update := gmqb.NewUpdate().Pop("scores", 1)  // remove last score
func (u Updater) Pop(field string, direction int) Updater {
	return u.addOp("$pop", field, direction)
}

// Pull removes all array elements that match a specified condition.
//
// MongoDB equivalent:
//
//	{ $pull: { field: condition } }
//
// The condition can be a value for exact matching, or a Filter for complex conditions.
//
// See: https://www.mongodb.com/docs/manual/reference/operator/update/pull/
//
// Example:
//
//	// Remove exact value
//	update := gmqb.NewUpdate().Pull("tags", "obsolete")
//
//	// Remove with condition
//	update := gmqb.NewUpdate().Pull("scores", gmqb.Lt("", 50).BsonD())
func (u Updater) Pull(field string, condition interface{}) Updater {
	return u.addOp("$pull", field, condition)
}

// PullAll removes all matching values from an array.
//
// MongoDB equivalent:
//
//	{ $pullAll: { field: [value1, value2, ...] } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/update/pullAll/
//
// Example:
//
//	update := gmqb.NewUpdate().PullAll("tags", "old", "deprecated")
func (u Updater) PullAll(field string, values ...interface{}) Updater {
	return u.addOp("$pullAll", field, bson.A(values))
}

// Push appends a value to an array.
//
// MongoDB equivalent:
//
//	{ $push: { field: value } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/update/push/
//
// Example:
//
//	update := gmqb.NewUpdate().Push("scores", 95)
func (u Updater) Push(field string, value interface{}) Updater {
	return u.addOp("$push", field, value)
}

// PushOpts configures modifiers for the Push operation.
// See: https://www.mongodb.com/docs/manual/reference/operator/update/push/#modifiers
type PushOpts struct {
	// Each appends multiple values. Required if using other modifiers.
	// See: https://www.mongodb.com/docs/manual/reference/operator/update/each/
	Each []interface{}

	// Position specifies where in the array to insert the elements.
	// See: https://www.mongodb.com/docs/manual/reference/operator/update/position/
	Position *int

	// Slice limits the size of the array after the push.
	// See: https://www.mongodb.com/docs/manual/reference/operator/update/slice/
	Slice *int

	// Sort orders the array elements. Use bson.D for compound sort.
	// See: https://www.mongodb.com/docs/manual/reference/operator/update/sort/
	Sort interface{}
}

// PushWithOpts pushes elements to an array with modifiers ($each, $position, $slice, $sort).
//
// MongoDB equivalent:
//
//	{ $push: { field: { $each: [...], $position: n, $slice: n, $sort: spec } } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/update/push/
//
// Example:
//
//	update := gmqb.NewUpdate().PushWithOpts("scores", gmqb.PushOpts{
//	    Each:  []interface{}{89, 92, 78},
//	    Slice: intPtr(-5),     // keep only last 5
//	    Sort:  bson.D{{"score", -1}},
//	})
func (u Updater) PushWithOpts(field string, opts PushOpts) Updater {
	modifier := bson.D{{Key: "$each", Value: bson.A(opts.Each)}}
	if opts.Position != nil {
		modifier = append(modifier, bson.E{Key: "$position", Value: *opts.Position})
	}
	if opts.Slice != nil {
		modifier = append(modifier, bson.E{Key: "$slice", Value: *opts.Slice})
	}
	if opts.Sort != nil {
		modifier = append(modifier, bson.E{Key: "$sort", Value: opts.Sort})
	}
	return u.addOp("$push", field, modifier)
}

// --- Bitwise Update Operator ---

// BitAnd performs a bitwise AND operation on an integer field value.
//
// MongoDB equivalent:
//
//	{ $bit: { field: { and: value } } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/update/bit/
//
// Example:
//
//	update := gmqb.NewUpdate().BitAnd("flags", 0b1010)
func (u Updater) BitAnd(field string, value int64) Updater {
	return u.addOp("$bit", field, bson.D{{Key: "and", Value: value}})
}

// BitOr performs a bitwise OR operation on an integer field value.
//
// MongoDB equivalent:
//
//	{ $bit: { field: { or: value } } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/update/bit/
//
// Example:
//
//	update := gmqb.NewUpdate().BitOr("flags", 0b0101)
func (u Updater) BitOr(field string, value int64) Updater {
	return u.addOp("$bit", field, bson.D{{Key: "or", Value: value}})
}

// BitXor performs a bitwise XOR operation on an integer field value.
//
// MongoDB equivalent:
//
//	{ $bit: { field: { xor: value } } }
//
// See: https://www.mongodb.com/docs/manual/reference/operator/update/bit/
//
// Example:
//
//	update := gmqb.NewUpdate().BitXor("flags", 0b1111)
func (u Updater) BitXor(field string, value int64) Updater {
	return u.addOp("$bit", field, bson.D{{Key: "xor", Value: value}})
}
