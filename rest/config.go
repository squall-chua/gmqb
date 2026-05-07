package rest

import (
	"context"
	"net/http"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// OpFlags represents the allowed operators for a filter field.
type OpFlags uint8

const (
	OpEq OpFlags = 1 << iota
	OpNe
	OpGt
	OpGte
	OpLt
	OpLte
	OpIn
	OpNin
)

// FilterField defines a field that can be used for filtering via query parameters.
type FilterField struct {
	Name    string  // query param key (e.g. "age")
	BsonKey string  // actual bson path (e.g. "profile.age")
	Op      OpFlags // allowed operators
	// ValueParser optionally converts the string query value to a specific type (e.g. int, bool).
	ValueParser func(s string) (any, error)
}

// Hooks defines lifecycle hooks for CRUD operations.
type Hooks[T any, ID comparable] struct {
	BeforeCreate func(ctx context.Context, doc *T) error
	AfterCreate  func(ctx context.Context, doc *T)
	BeforeUpdate func(ctx context.Context, id ID, patch map[string]any) error
	AfterUpdate  func(ctx context.Context, id ID, patch map[string]any)
	BeforeReplace func(ctx context.Context, id ID, doc *T) error
	AfterReplace  func(ctx context.Context, id ID, doc *T)
	BeforeDelete func(ctx context.Context, id ID) error
	AfterDelete  func(ctx context.Context, id ID)
}

// Encoder defines the interface for encoding responses.
type Encoder interface {
	Write(w http.ResponseWriter, v any) error
}

// Config defines the configuration for a REST resource.
type Config[T any, ID comparable] struct {
	// IDField is the bson key of the ID field, e.g. "_id".
	IDField string
	// IDParser parses the URL path segment into the ID type.
	IDParser func(s string) (ID, error)

	// FilterableFields defines the allowlist of fields that can be used for filtering.
	FilterableFields []FilterField

	// SortableFields is an allowlist of fields that can be used for sorting.
	SortableFields []string
	// DefaultSort is the default sort order if none is specified in the request.
	DefaultSort bson.D

	// Pagination settings
	DefaultLimit int64 // default page size
	MaxLimit     int64 // maximum allowed page size

	// Lifecycle hooks
	Hooks Hooks[T, ID]

	// Encoder for response serialization. Defaults to JSON if nil.
	Encoder Encoder
}
