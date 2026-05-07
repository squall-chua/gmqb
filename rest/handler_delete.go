package rest

import (
	"net/http"

	"github.com/squall-chua/gmqb"
)

func (res *Resource[T, ID]) handleDelete(w http.ResponseWriter, r *http.Request, id ID) {
	ctx := r.Context()

	// Hook: BeforeDelete
	if res.cfg.Hooks.BeforeDelete != nil {
		if err := res.cfg.Hooks.BeforeDelete(ctx, id); err != nil {
			writeError(w, res.cfg.Encoder, http.StatusUnprocessableEntity, err.Error(), "BEFORE_DELETE_FAILED")
			return
		}
	}

	filter := gmqb.Eq(res.cfg.IDField, id)
	result, err := res.coll.DeleteOne(ctx, filter)
	if err != nil {
		writeError(w, res.cfg.Encoder, http.StatusInternalServerError, "failed to delete document", "DELETE_FAILED")
		return
	}

	if result.DeletedCount == 0 {
		writeError(w, res.cfg.Encoder, http.StatusNotFound, "document not found", "NOT_FOUND")
		return
	}

	// Hook: AfterDelete
	if res.cfg.Hooks.AfterDelete != nil {
		res.cfg.Hooks.AfterDelete(ctx, id)
	}

	w.WriteHeader(http.StatusNoContent)
}
