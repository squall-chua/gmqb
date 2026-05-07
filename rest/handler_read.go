package rest

import (
	"errors"
	"net/http"

	"github.com/squall-chua/gmqb"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func (res *Resource[T, ID]) handleRead(w http.ResponseWriter, r *http.Request, id ID) {
	ctx := r.Context()
	
	filter := gmqb.Eq(res.cfg.IDField, id)
	doc, err := res.coll.FindOne(ctx, filter)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			writeError(w, res.cfg.Encoder, http.StatusNotFound, "document not found", "NOT_FOUND")
		} else {
			writeError(w, res.cfg.Encoder, http.StatusInternalServerError, "failed to read document", "READ_FAILED")
		}
		return
	}

	writeJSON(w, res.cfg.Encoder, http.StatusOK, doc, nil)
}
