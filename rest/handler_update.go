package rest

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/squall-chua/gmqb"
)

func (res *Resource[T, ID]) handleReplace(w http.ResponseWriter, r *http.Request, id ID) {
	ctx := r.Context()
	var doc T
	if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
		writeError(w, res.cfg.Encoder, http.StatusBadRequest, "invalid json body", "INVALID_JSON")
		return
	}

	// Hook: BeforeReplace
	if res.cfg.Hooks.BeforeReplace != nil {
		if err := res.cfg.Hooks.BeforeReplace(ctx, id, &doc); err != nil {
			writeError(w, res.cfg.Encoder, http.StatusUnprocessableEntity, err.Error(), "BEFORE_REPLACE_FAILED")
			return
		}
	}

	filter := gmqb.Eq(res.cfg.IDField, id)
	result, err := res.coll.ReplaceOne(ctx, filter, &doc)
	if err != nil {
		writeError(w, res.cfg.Encoder, http.StatusInternalServerError, "failed to replace document", "REPLACE_FAILED")
		return
	}

	if result.MatchedCount == 0 {
		writeError(w, res.cfg.Encoder, http.StatusNotFound, "document not found", "NOT_FOUND")
		return
	}

	// Hook: AfterReplace
	if res.cfg.Hooks.AfterReplace != nil {
		res.cfg.Hooks.AfterReplace(ctx, id, &doc)
	}

	writeJSON(w, res.cfg.Encoder, http.StatusOK, doc, nil)
}

func (res *Resource[T, ID]) handlePatch(w http.ResponseWriter, r *http.Request, id ID) {
	ctx := r.Context()
	var patch map[string]any
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeError(w, res.cfg.Encoder, http.StatusBadRequest, "invalid json body", "INVALID_JSON")
		return
	}

	// Guard against NoSQL injection in keys
	for k := range patch {
		if strings.HasPrefix(k, "$") {
			writeError(w, res.cfg.Encoder, http.StatusBadRequest, "operator injection detected in patch key", "INJECTION_DETECTED")
			return
		}
	}

	// Hook: BeforeUpdate
	if res.cfg.Hooks.BeforeUpdate != nil {
		if err := res.cfg.Hooks.BeforeUpdate(ctx, id, patch); err != nil {
			writeError(w, res.cfg.Encoder, http.StatusUnprocessableEntity, err.Error(), "BEFORE_UPDATE_FAILED")
			return
		}
	}

	update := gmqb.NewUpdate()
	for k, v := range patch {
		update = update.Set(k, v)
	}

	filter := gmqb.Eq(res.cfg.IDField, id)
	result, err := res.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		writeError(w, res.cfg.Encoder, http.StatusInternalServerError, "failed to patch document", "PATCH_FAILED")
		return
	}

	if result.MatchedCount == 0 {
		writeError(w, res.cfg.Encoder, http.StatusNotFound, "document not found", "NOT_FOUND")
		return
	}

	// Hook: AfterUpdate
	if res.cfg.Hooks.AfterUpdate != nil {
		res.cfg.Hooks.AfterUpdate(ctx, id, patch)
	}

	// Fetch updated document to return it
	doc, err := res.coll.FindOne(ctx, filter)
	if err != nil {
		writeError(w, res.cfg.Encoder, http.StatusInternalServerError, "failed to fetch updated document", "FETCH_FAILED")
		return
	}

	writeJSON(w, res.cfg.Encoder, http.StatusOK, doc, nil)
}
