package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (res *Resource[T, ID]) handleCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var doc T
	if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
		writeError(w, res.cfg.Encoder, http.StatusBadRequest, "invalid json body", "INVALID_JSON")
		return
	}

	// Hook: BeforeCreate
	if res.cfg.Hooks.BeforeCreate != nil {
		if err := res.cfg.Hooks.BeforeCreate(ctx, &doc); err != nil {
			writeError(w, res.cfg.Encoder, http.StatusUnprocessableEntity, err.Error(), "BEFORE_CREATE_FAILED")
			return
		}
	}

	result, err := res.coll.InsertOne(ctx, &doc)
	if err != nil {
		writeError(w, res.cfg.Encoder, http.StatusInternalServerError, "failed to insert document", "INSERT_FAILED")
		return
	}

	// Hook: AfterCreate
	if res.cfg.Hooks.AfterCreate != nil {
		res.cfg.Hooks.AfterCreate(ctx, &doc)
	}

	// Set Location header
	w.Header().Set("Location", fmt.Sprintf("%s/%v", r.URL.Path, result.InsertedID))
	writeJSON(w, res.cfg.Encoder, http.StatusCreated, doc, nil)
}
