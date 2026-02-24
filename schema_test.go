package gmqb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testUser struct {
	ID      string      `bson:"_id"`
	Name    string      `bson:"name"`
	Age     int         `bson:"age"`
	Email   string      `bson:"email"`
	Tags    []string    `bson:"tags"`
	Address testAddress `bson:"address"`
}

type testAddress struct {
	City    string `bson:"city"`
	Country string `bson:"country"`
	Zip     string `bson:"zip_code"`
}

func TestField_SimpleField(t *testing.T) {
	assert.Equal(t, "name", Field[testUser]("Name"))
}

func TestField_NestedField(t *testing.T) {
	assert.Equal(t, "address.city", Field[testUser]("Address.City"))
}

func TestField_NestedFieldWithCustomTag(t *testing.T) {
	assert.Equal(t, "address.zip_code", Field[testUser]("Address.Zip"))
}

func TestField_PanicsOnInvalidPath(t *testing.T) {
	require.Panics(t, func() {
		Field[testUser]("NonExistent")
	})
}

func TestField_CachesResults(t *testing.T) {
	got1 := Field[testUser]("Name")
	got2 := Field[testUser]("Name")
	assert.Equal(t, got1, got2)
}

func TestFieldRef(t *testing.T) {
	assert.Equal(t, "$age", FieldRef("age"))
}

func TestField_Pointer(t *testing.T) {
	assert.Equal(t, "name", Field[*testUser]("Name"))
}
