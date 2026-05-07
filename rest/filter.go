package rest

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/squall-chua/gmqb"
)

var (
	ErrUnknownFilterField = errors.New("unknown filter field")
	ErrOperatorNotAllowed = errors.New("operator not allowed for field")
)

// BuildFilter parses query parameters and builds a gmqb.Filter based on the allowlist.
func BuildFilter(q url.Values, fields []FilterField) (gmqb.Filter, error) {
	filter := gmqb.NewFilter()

	// Map to look up filter fields by name
	fieldMap := make(map[string]FilterField)
	for _, f := range fields {
		fieldMap[f.Name] = f
	}

	for key, values := range q {
		// Ignore pagination and sort params
		if key == "limit" || key == "skip" || key == "offset" || key == "cursor" || key == "after" || key == "before" || key == "sort" {
			continue
		}

		// Parse field name and operator suffix (e.g., "age_gte" -> "age", "gte")
		fieldName, opSuffix := parseKey(key)

		ff, ok := fieldMap[fieldName]
		if !ok {
			return gmqb.Filter{}, fmt.Errorf("%w: %s", ErrUnknownFilterField, fieldName)
		}

		op, err := getOpFromSuffix(opSuffix)
		if err != nil {
			return gmqb.Filter{}, err
		}

		// Check if the operator is allowed for this field
		if ff.Op&op == 0 {
			return gmqb.Filter{}, fmt.Errorf("%w: %s for field %s", ErrOperatorNotAllowed, opSuffix, fieldName)
		}

		// Apply filter to the builder
		for _, val := range values {
			if op == OpIn || op == OpNin {
				parts := strings.Split(val, ",")
				ifaces := make([]any, len(parts))
				for i, p := range parts {
					if ff.ValueParser != nil {
						parsed, err := ff.ValueParser(p)
						if err != nil {
							return gmqb.Filter{}, fmt.Errorf("invalid value in list for field %s: %v", fieldName, err)
						}
						ifaces[i] = parsed
					} else {
						ifaces[i] = p
					}
				}
				filter = applyOp(filter, ff.BsonKey, op, ifaces)
			} else {
				var finalVal any = val
				if ff.ValueParser != nil {
					var err error
					finalVal, err = ff.ValueParser(val)
					if err != nil {
						return gmqb.Filter{}, fmt.Errorf("invalid value for field %s: %v", fieldName, err)
					}
				}
				filter = applyOp(filter, ff.BsonKey, op, finalVal)
			}
		}
	}

	return filter, nil
}

func parseKey(key string) (string, string) {
	// Try bracket notation: field[op]
	if idx := strings.Index(key, "["); idx != -1 && strings.HasSuffix(key, "]") {
		return key[:idx], key[idx+1 : len(key)-1]
	}

	// Fallback to underscore notation: field_op (e.g. created_at_gte)
	parts := strings.Split(key, "_")
	if len(parts) < 2 {
		return key, ""
	}
	// Check if the last part is a known operator suffix
	suffix := parts[len(parts)-1]
	switch suffix {
	case "eq", "ne", "gt", "gte", "lt", "lte", "in", "nin":
		return strings.Join(parts[:len(parts)-1], "_"), suffix
	default:
		return key, ""
	}
}

func getOpFromSuffix(suffix string) (OpFlags, error) {
	switch suffix {
	case "", "eq":
		return OpEq, nil
	case "ne":
		return OpNe, nil
	case "gt":
		return OpGt, nil
	case "gte":
		return OpGte, nil
	case "lt":
		return OpLt, nil
	case "lte":
		return OpLte, nil
	case "in":
		return OpIn, nil
	case "nin":
		return OpNin, nil
	default:
		return 0, fmt.Errorf("invalid operator suffix: %s", suffix)
	}
}

func applyOp(f gmqb.Filter, bsonKey string, op OpFlags, val any) gmqb.Filter {
	switch op {
	case OpEq:
		return f.Eq(bsonKey, val)
	case OpNe:
		return f.Ne(bsonKey, val)
	case OpGt:
		return f.Gt(bsonKey, val)
	case OpGte:
		return f.Gte(bsonKey, val)
	case OpLt:
		return f.Lt(bsonKey, val)
	case OpLte:
		return f.Lte(bsonKey, val)
	case OpIn:
		if ifaces, ok := val.([]any); ok {
			return f.In(bsonKey, ifaces...)
		}
		return f.In(bsonKey, val)
	case OpNin:
		if ifaces, ok := val.([]any); ok {
			return f.Nin(bsonKey, ifaces...)
		}
		return f.Nin(bsonKey, val)
	}
	return f
}
