package rest

import (
	"net/http"

	"github.com/squall-chua/gmqb"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (res *Resource[T, ID]) handleList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()

	// 1. Build Filter
	filter, err := BuildFilter(q, res.cfg.FilterableFields)
	if err != nil {
		writeError(w, res.cfg.Encoder, http.StatusBadRequest, err.Error(), "INVALID_FILTER")
		return
	}

	// 2. Parse Pagination
	pg, err := ParsePagination(q, res.cfg)
	if err != nil {
		writeError(w, res.cfg.Encoder, http.StatusBadRequest, err.Error(), "INVALID_PAGINATION")
		return
	}

	// 3. Apply Pagination & Sort to Filter
	var findOpts []gmqb.FindOpt
	findOpts = append(findOpts, gmqb.WithLimit(pg.Limit))

	currentSort := pg.Sort
	if pg.Mode == ModeOffset {
		findOpts = append(findOpts, gmqb.WithSkip(pg.Skip))
	} else {
		// Cursor Mode
		if pg.After != nil {
			filter = applyCursorFilter(filter, pg.Sort, pg.After, true)
		} else if pg.Before != nil {
			filter = applyCursorFilter(filter, pg.Sort, pg.Before, false)

			// For backward navigation, we must invert the sort order during fetch
			// to get the items closest to the anchor first.
			currentSort = make(bson.D, len(pg.Sort))
			for i, e := range pg.Sort {
				currentSort[i] = bson.E{Key: e.Key, Value: -e.Value.(int)}
			}
		}
	}
	findOpts = append(findOpts, gmqb.WithSort(currentSort))

	// 4. Fetch Documents
	docs, err := res.coll.Find(ctx, filter, findOpts...)
	if err != nil {
		writeError(w, res.cfg.Encoder, http.StatusInternalServerError, "failed to fetch documents", "FETCH_FAILED")
		return
	}

	// If we were navigating backward, reverse the results to restore original order
	if pg.Before != nil {
		for i, j := 0, len(docs)-1; i < j; i, j = i+1, j-1 {
			docs[i], docs[j] = docs[j], docs[i]
		}
	}

	// 5. Build Meta
	meta := &Meta{
		Limit: pg.Limit,
	}

	if pg.Mode == ModeOffset {
		meta.Skip = &pg.Skip
		total, err := res.coll.CountDocuments(ctx, filter)
		if err == nil {
			meta.Total = &total
		}
	}

	// Build Next Cursor from last document if we have results
	if len(docs) > 0 {
		lastDoc := docs[len(docs)-1]
		cursorData := extractSortKeys(lastDoc, pg.Sort)
		next, _ := EncodeCursor(cursorData, false)
		meta.Next = &next
	}

	// Build Prev Cursor from first document if we are not on the first page
	if len(docs) > 0 && (pg.Skip > 0 || pg.After != nil || pg.Before != nil) {
		firstDoc := docs[0]
		cursorData := extractSortKeys(firstDoc, pg.Sort)
		prev, _ := EncodeCursor(cursorData, true)
		meta.Prev = &prev
	}

	writeJSON(w, res.cfg.Encoder, http.StatusOK, docs, meta)
}

func applyCursorFilter(f gmqb.Filter, sort bson.D, cursor bson.D, after bool) gmqb.Filter {
	if len(sort) == 0 {
		return f
	}

	// Create a map for quick cursor value lookup
	cursorMap := make(map[string]interface{})
	for _, e := range cursor {
		cursorMap[e.Key] = e.Value
	}

	orClauses := make([]bson.D, 0, len(sort))

	// Build the compound comparison logic:
	// ($or: [
	//   { f1: { $gt: v1 } },
	//   { f1: v1, f2: { $gt: v2 } },
	//   { f1: v1, f2: v2, f3: { $gt: v3 } }
	// ])
	for i := 0; i < len(sort); i++ {
		clause := bson.D{}

		// 1. Add equality constraints for all preceding fields to ensure we are on the same "branch"
		valid := true
		for j := 0; j < i; j++ {
			key := sort[j].Key
			val, ok := cursorMap[key]
			if !ok {
				valid = false
				break
			}
			clause = append(clause, bson.E{Key: key, Value: val})
		}
		if !valid {
			continue
		}

		// 2. Add the inequality constraint for the current field
		target := sort[i]
		val, ok := cursorMap[target.Key]
		if !ok {
			continue
		}

		direction := 1
		if d, ok := target.Value.(int); ok {
			direction = d
		} else if d, ok := target.Value.(int32); ok {
			direction = int(d)
		} else if d, ok := target.Value.(int64); ok {
			direction = int(d)
		}

		op := "$gt"
		// Invert operator if:
		// - Ascending (1) and we want "before" (!after)
		// - Descending (-1) and we want "after" (after)
		if (direction == 1 && !after) || (direction == -1 && after) {
			op = "$lt"
		}

		clause = append(clause, bson.E{Key: target.Key, Value: bson.D{{Key: op, Value: val}}})
		orClauses = append(orClauses, clause)
	}

	if len(orClauses) == 0 {
		return f
	}

	d := f.BsonD()
	if len(orClauses) == 1 {
		// Optimization: single field doesn't need $or
		d = append(d, orClauses[0]...)
	} else {
		d = append(d, bson.E{Key: "$or", Value: orClauses})
	}

	return gmqb.Raw(d)
}

func extractSortKeys(doc interface{}, sort bson.D) bson.D {
	// Naive implementation: marshal to bson and extract
	data, _ := bson.Marshal(doc)
	var raw bson.D
	_ = bson.Unmarshal(data, &raw)

	res := make(bson.D, 0, len(sort))
	rawMap := make(map[string]interface{})
	for _, e := range raw {
		rawMap[e.Key] = e.Value
	}

	for _, se := range sort {
		if val, ok := rawMap[se.Key]; ok {
			res = append(res, bson.E{Key: se.Key, Value: val})
		}
	}
	return res
}
