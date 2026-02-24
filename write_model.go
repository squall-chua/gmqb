package gmqb

import (
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// WriteModel is an interface that wraps mongo.WriteModel, restricted to document type T.
type WriteModel[T any] interface {
	MongoWriteModel() mongo.WriteModel
}

// InsertOneModel is a type-safe wrapper for mongo.InsertOneModel.
type InsertOneModel[T any] struct {
	model *mongo.InsertOneModel
}

// NewInsertOneModel creates a new InsertOneModel.
func NewInsertOneModel[T any]() *InsertOneModel[T] {
	return &InsertOneModel[T]{
		model: mongo.NewInsertOneModel(),
	}
}

// SetDocument sets the document to insert.
func (m *InsertOneModel[T]) SetDocument(doc *T) *InsertOneModel[T] {
	m.model.SetDocument(doc)
	return m
}

// MongoWriteModel implements WriteModel interface.
func (m *InsertOneModel[T]) MongoWriteModel() mongo.WriteModel {
	return m.model
}

// ReplaceOneModel is a type-safe wrapper for mongo.ReplaceOneModel.
type ReplaceOneModel[T any] struct {
	model *mongo.ReplaceOneModel
}

// NewReplaceOneModel creates a new ReplaceOneModel.
func NewReplaceOneModel[T any]() *ReplaceOneModel[T] {
	return &ReplaceOneModel[T]{
		model: mongo.NewReplaceOneModel(),
	}
}

// SetFilter sets the filter to match the document to replace.
func (m *ReplaceOneModel[T]) SetFilter(filter Filter) *ReplaceOneModel[T] {
	m.model.SetFilter(filter.BsonD())
	return m
}

// SetReplacement sets the replacement document.
func (m *ReplaceOneModel[T]) SetReplacement(rep *T) *ReplaceOneModel[T] {
	m.model.SetReplacement(rep)
	return m
}

// SetUpsert sets the upsert flag.
func (m *ReplaceOneModel[T]) SetUpsert(upsert bool) *ReplaceOneModel[T] {
	m.model.SetUpsert(upsert)
	return m
}

// MongoWriteModel implements WriteModel interface.
func (m *ReplaceOneModel[T]) MongoWriteModel() mongo.WriteModel {
	return m.model
}

// UpdateOneModel is a type-safe wrapper for mongo.UpdateOneModel.
type UpdateOneModel[T any] struct {
	model *mongo.UpdateOneModel
}

// NewUpdateOneModel creates a new UpdateOneModel.
func NewUpdateOneModel[T any]() *UpdateOneModel[T] {
	return &UpdateOneModel[T]{
		model: mongo.NewUpdateOneModel(),
	}
}

// SetFilter sets the filter to match the document to update.
func (m *UpdateOneModel[T]) SetFilter(filter Filter) *UpdateOneModel[T] {
	m.model.SetFilter(filter.BsonD())
	return m
}

// SetUpdate sets the update operations.
func (m *UpdateOneModel[T]) SetUpdate(update Updater) *UpdateOneModel[T] {
	m.model.SetUpdate(update.BsonD())
	return m
}

// SetUpsert sets the upsert flag.
func (m *UpdateOneModel[T]) SetUpsert(upsert bool) *UpdateOneModel[T] {
	m.model.SetUpsert(upsert)
	return m
}

// MongoWriteModel implements WriteModel interface.
func (m *UpdateOneModel[T]) MongoWriteModel() mongo.WriteModel {
	return m.model
}

// UpdateManyModel is a type-safe wrapper for mongo.UpdateManyModel.
type UpdateManyModel[T any] struct {
	model *mongo.UpdateManyModel
}

// NewUpdateManyModel creates a new UpdateManyModel.
func NewUpdateManyModel[T any]() *UpdateManyModel[T] {
	return &UpdateManyModel[T]{
		model: mongo.NewUpdateManyModel(),
	}
}

// SetFilter sets the filter to match the documents to update.
func (m *UpdateManyModel[T]) SetFilter(filter Filter) *UpdateManyModel[T] {
	m.model.SetFilter(filter.BsonD())
	return m
}

// SetUpdate sets the update operations.
func (m *UpdateManyModel[T]) SetUpdate(update Updater) *UpdateManyModel[T] {
	m.model.SetUpdate(update.BsonD())
	return m
}

// SetUpsert sets the upsert flag.
func (m *UpdateManyModel[T]) SetUpsert(upsert bool) *UpdateManyModel[T] {
	m.model.SetUpsert(upsert)
	return m
}

// MongoWriteModel implements WriteModel interface.
func (m *UpdateManyModel[T]) MongoWriteModel() mongo.WriteModel {
	return m.model
}

// DeleteOneModel is a type-safe wrapper for mongo.DeleteOneModel.
type DeleteOneModel[T any] struct {
	model *mongo.DeleteOneModel
}

// NewDeleteOneModel creates a new DeleteOneModel.
func NewDeleteOneModel[T any]() *DeleteOneModel[T] {
	return &DeleteOneModel[T]{
		model: mongo.NewDeleteOneModel(),
	}
}

// SetFilter sets the filter to match the document to delete.
func (m *DeleteOneModel[T]) SetFilter(filter Filter) *DeleteOneModel[T] {
	m.model.SetFilter(filter.BsonD())
	return m
}

// MongoWriteModel implements WriteModel interface.
func (m *DeleteOneModel[T]) MongoWriteModel() mongo.WriteModel {
	return m.model
}

// DeleteManyModel is a type-safe wrapper for mongo.DeleteManyModel.
type DeleteManyModel[T any] struct {
	model *mongo.DeleteManyModel
}

// NewDeleteManyModel creates a new DeleteManyModel.
func NewDeleteManyModel[T any]() *DeleteManyModel[T] {
	return &DeleteManyModel[T]{
		model: mongo.NewDeleteManyModel(),
	}
}

// SetFilter sets the filter to match the documents to delete.
func (m *DeleteManyModel[T]) SetFilter(filter Filter) *DeleteManyModel[T] {
	m.model.SetFilter(filter.BsonD())
	return m
}

// MongoWriteModel implements WriteModel interface.
func (m *DeleteManyModel[T]) MongoWriteModel() mongo.WriteModel {
	return m.model
}
