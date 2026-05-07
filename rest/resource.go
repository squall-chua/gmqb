package rest

import (
	"net/http"

	"github.com/squall-chua/gmqb"
)

// Resource represents a RESTful resource mapped to a gmqb.Collection.
type Resource[T any, ID comparable] struct {
	coll *gmqb.Collection[T]
	cfg  Config[T, ID]
}

// NewResource creates a new REST resource for the given collection and config.
func NewResource[T any, ID comparable](coll *gmqb.Collection[T], cfg Config[T, ID]) *Resource[T, ID] {
	return &Resource[T, ID]{
		coll: coll,
		cfg:  cfg,
	}
}

// ServeHTTP implements the http.Handler interface.
// It dispatches requests to the appropriate CRUD handlers based on method and path.
func (res *Resource[T, ID]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// We use Go 1.22+ PathValue if available, or fallback to simple path suffix check.
	// Since we are framework agnostic, we try to be as simple as possible.

	id := r.PathValue("id")

	// Route based on method and presence of ID
	if id == "" {
		switch r.Method {
		case http.MethodGet:
			res.handleList(w, r)
		case http.MethodPost:
			res.handleCreate(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	} else {
		parsedID, err := res.cfg.IDParser(id)
		if err != nil {
			writeError(w, res.cfg.Encoder, http.StatusBadRequest, "invalid id format", "INVALID_ID")
			return
		}

		switch r.Method {
		case http.MethodGet:
			res.handleRead(w, r, parsedID)
		case http.MethodPut:
			res.handleReplace(w, r, parsedID)
		case http.MethodPatch:
			res.handlePatch(w, r, parsedID)
		case http.MethodDelete:
			res.handleDelete(w, r, parsedID)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	}
}
