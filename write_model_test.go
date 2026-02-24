package gmqb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteModel(t *testing.T) {
	type user struct {
		Name string
		Age  int
	}

	t.Run("InsertOneModel", func(t *testing.T) {
		doc := &user{Name: "Alice", Age: 30}
		m := NewInsertOneModel[user]().SetDocument(doc)
		assert.NotNil(t, m.MongoWriteModel())
	})

	t.Run("ReplaceOneModel", func(t *testing.T) {
		doc := &user{Name: "Alice", Age: 31}
		m := NewReplaceOneModel[user]().
			SetFilter(Eq("name", "Alice")).
			SetReplacement(doc).
			SetUpsert(true)
		assert.NotNil(t, m.MongoWriteModel())
	})

	t.Run("UpdateOneModel", func(t *testing.T) {
		m := NewUpdateOneModel[user]().
			SetFilter(Eq("name", "Alice")).
			SetUpdate(NewUpdate().Set("age", 31)).
			SetUpsert(true)
		assert.NotNil(t, m.MongoWriteModel())
	})

	t.Run("UpdateManyModel", func(t *testing.T) {
		m := NewUpdateManyModel[user]().
			SetFilter(Gt("age", 20)).
			SetUpdate(NewUpdate().Inc("age", 1)).
			SetUpsert(true)
		assert.NotNil(t, m.MongoWriteModel())
	})

	t.Run("DeleteOneModel", func(t *testing.T) {
		m := NewDeleteOneModel[user]().
			SetFilter(Eq("name", "Alice"))
		assert.NotNil(t, m.MongoWriteModel())
	})

	t.Run("DeleteManyModel", func(t *testing.T) {
		m := NewDeleteManyModel[user]().
			SetFilter(Lt("age", 18))
		assert.NotNil(t, m.MongoWriteModel())
	})
}
