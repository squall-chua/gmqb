package rest

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type PaginationMode int

const (
	ModeOffset PaginationMode = iota
	ModeCursor
)

type PaginationResult struct {
	Mode   PaginationMode
	Limit  int64
	Skip   int64
	After  bson.D
	Before bson.D
	Sort   bson.D
}

// ParsePagination extracts pagination and sort parameters from the query string.
func ParsePagination[T any, ID comparable](q url.Values, cfg Config[T, ID]) (PaginationResult, error) {
	res := PaginationResult{
		Mode:  ModeOffset,
		Limit: cfg.DefaultLimit,
	}

	if res.Limit == 0 {
		res.Limit = 20
	}

	// Limit
	if l := q.Get("limit"); l != "" {
		if limit, err := strconv.ParseInt(l, 10, 64); err == nil {
			res.Limit = limit
		}
	}
	if cfg.MaxLimit > 0 && res.Limit > cfg.MaxLimit {
		res.Limit = cfg.MaxLimit
	}

	// Mode Detection & Params
	after := q.Get("after")
	if after == "" {
		after = q.Get("cursor") // Alias
	}
	before := q.Get("before")
	skipStr := q.Get("skip")
	if skipStr == "" {
		skipStr = q.Get("offset") // Alias
	}

	if after != "" || before != "" {
		res.Mode = ModeCursor
		if after != "" {
			cursor, reverse, err := DecodeCursor(after)
			if err != nil {
				return res, fmt.Errorf("invalid cursor token: %w", err)
			}
			if reverse {
				res.Before = cursor
			} else {
				res.After = cursor
			}
		}
		if before != "" {
			cursor, reverse, err := DecodeCursor(before)
			if err != nil {
				return res, fmt.Errorf("invalid before token: %w", err)
			}
			// If explicitly using ?before=, we use Before mode regardless of token flag,
			// but we handle the flag just in case.
			if reverse {
				res.After = cursor // Before a reverse cursor is forward
			} else {
				res.Before = cursor
			}
		}
	} else if skipStr != "" {
		if s, err := strconv.ParseInt(skipStr, 10, 64); err == nil {
			res.Skip = s
		}
	}

	// Sort
	sortStr := q.Get("sort")
	if sortStr != "" {
		sortD, err := parseSort(sortStr, cfg.SortableFields)
		if err != nil {
			return res, err
		}
		res.Sort = sortD
	} else {
		res.Sort = cfg.DefaultSort
	}

	return res, nil
}

func parseSort(s string, allowed []string) (bson.D, error) {
	allowedMap := make(map[string]bool)
	for _, f := range allowed {
		allowedMap[f] = true
	}

	parts := strings.Split(s, ",")
	d := make(bson.D, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		order := 1
		field := p
		if strings.HasPrefix(p, "-") {
			order = -1
			field = p[1:]
		}

		if !allowedMap[field] {
			return nil, fmt.Errorf("field not sortable: %s", field)
		}

		d = append(d, bson.E{Key: field, Value: order})
	}

	return d, nil
}
