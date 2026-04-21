package gmqb

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// IndexModel represents a MongoDB index specification.
// It wraps a set of keys and optional configuration.
type IndexModel struct {
	keys    bson.D
	options *options.IndexOptionsBuilder
}

// NewIndex creates a new IndexModel with the specified keys.
// Keys can be created using gmqb.Asc, gmqb.Desc, or gmqb.SortSpec.
//
// Example:
//
//	index := gmqb.NewIndex(gmqb.Asc("email")).Unique()
func NewIndex(keys bson.D) IndexModel {
	return IndexModel{
		keys:    keys,
		options: options.Index(),
	}
}

// Unique sets the index to be unique.
func (m IndexModel) Unique() IndexModel {
	m.options.SetUnique(true)
	return m
}

// Sparse sets the index to be sparse.
func (m IndexModel) Sparse() IndexModel {
	m.options.SetSparse(true)
	return m
}

// TTL sets the time-to-live for documents in seconds.
func (m IndexModel) TTL(seconds int32) IndexModel {
	m.options.SetExpireAfterSeconds(seconds)
	return m
}

// Name sets a custom name for the index.
func (m IndexModel) Name(name string) IndexModel {
	m.options.SetName(name)
	return m
}

// MongoIndexModel converts the gmqb.IndexModel to a mongo.IndexModel.
func (m IndexModel) MongoIndexModel() mongo.IndexModel {
	return mongo.IndexModel{
		Keys:    m.keys,
		Options: m.options,
	}
}
