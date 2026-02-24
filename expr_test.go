package gmqb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestExprAdd(t *testing.T) {
	assert.Equal(t, "$add", ExprAdd("$price", "$tax")[0].Key)
}

func TestExprSubtract(t *testing.T) {
	assert.Equal(t, "$subtract", ExprSubtract("$total", "$discount")[0].Key)
}

func TestExprMultiply(t *testing.T) {
	assert.Equal(t, "$multiply", ExprMultiply("$qty", "$price")[0].Key)
}

func TestExprDivide(t *testing.T) {
	assert.Equal(t, "$divide", ExprDivide("$total", "$count")[0].Key)
}

func TestExprCond(t *testing.T) {
	assert.Equal(t, "$cond", ExprCond(ExprGte("$qty", 250), "high", "low")[0].Key)
}

func TestExprIfNull(t *testing.T) {
	assert.Equal(t, "$ifNull", ExprIfNull("$description", "N/A")[0].Key)
}

func TestExprSwitch(t *testing.T) {
	d := ExprSwitch([]SwitchBranch{
		{Case: ExprGte("$age", 65), Then: "senior"},
		{Case: ExprGte("$age", 18), Then: "adult"},
	}, "minor")
	assert.Equal(t, "$switch", d[0].Key)
}

func TestExprConcat(t *testing.T) {
	assert.Equal(t, "$concat", ExprConcat("$firstName", " ", "$lastName")[0].Key)
}

func TestExprToLower(t *testing.T) {
	assert.Equal(t, "$toLower", ExprToLower("$name")[0].Key)
}

func TestExprRegexMatch(t *testing.T) {
	assert.Equal(t, "$regexMatch", ExprRegexMatch("$email", `^test`, "i")[0].Key)
}

func TestExprArrayElemAt(t *testing.T) {
	assert.Equal(t, "$arrayElemAt", ExprArrayElemAt("$items", 0)[0].Key)
}

func TestExprFilter(t *testing.T) {
	assert.Equal(t, "$filter", ExprFilter("$items", "item", ExprGte("$$item.price", 100))[0].Key)
}

func TestExprMap(t *testing.T) {
	assert.Equal(t, "$map", ExprMap("$items", "item", ExprMultiply("$$item.price", "$$item.qty"))[0].Key)
}

func TestExprDateAdd(t *testing.T) {
	assert.Equal(t, "$dateAdd", ExprDateAdd("$orderDate", "day", 3)[0].Key)
}

func TestExprYear(t *testing.T) {
	assert.Equal(t, "$year", ExprYear("$createdAt")[0].Key)
}

func TestExprConvert(t *testing.T) {
	assert.Equal(t, "$convert", ExprConvert("$value", "double", nil, nil)[0].Key)
}

// --- Accumulators ---

func TestAccSum(t *testing.T) {
	assert.Equal(t, "$sum", AccSum(1)[0].Key)
}

func TestAccAvg(t *testing.T) {
	assert.Equal(t, "$avg", AccAvg("$score")[0].Key)
}

func TestAccFirst(t *testing.T) {
	assert.Equal(t, "$first", AccFirst("$name")[0].Key)
}

func TestAccPush(t *testing.T) {
	assert.Equal(t, "$push", AccPush("$item")[0].Key)
}

func TestAccCount(t *testing.T) {
	assert.Equal(t, "$count", AccCount()[0].Key)
}

func TestAccTop(t *testing.T) {
	assert.Equal(t, "$top", AccTop(bson.D{{"score", -1}}, "$name")[0].Key)
}

// --- Set Operators ---

func TestExprSetEquals(t *testing.T) {
	assert.Equal(t, "$setEquals", ExprSetEquals("$a", "$b")[0].Key)
}

func TestExprSetUnion(t *testing.T) {
	assert.Equal(t, "$setUnion", ExprSetUnion("$a", "$b")[0].Key)
}

// --- Object Operators ---

func TestExprMergeObjects(t *testing.T) {
	assert.Equal(t, "$mergeObjects", ExprMergeObjects("$defaults", "$overrides")[0].Key)
}

// --- Miscellaneous ---

func TestExprLiteral(t *testing.T) {
	assert.Equal(t, "$literal", ExprLiteral("$notAField")[0].Key)
}

func TestExprRand(t *testing.T) {
	assert.Equal(t, "$rand", ExprRand()[0].Key)
}

func TestExprLet(t *testing.T) {
	assert.Equal(t, "$let", ExprLet(bson.D{{"total", ExprAdd("$price", "$tax")}}, "$$total")[0].Key)
}
