package gmqb

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// schemaCache stores resolved field paths to avoid repeated reflection.
var schemaCache sync.Map // map[reflect.Type]map[string]string

// Field resolves a Go struct field path to its BSON field name using bson struct tags.
// The type parameter T specifies the struct type to reflect on. The fieldPath argument
// is the Go struct field name (e.g. "Name") or a dotted path for nested structs
// (e.g. "Address.City").
//
// Field uses a sync.Map cache so that reflection is performed at most once per type.
// It panics if the field path does not exist in the struct — this is intentional to
// catch typos at startup rather than producing silent runtime errors.
//
// See: https://www.mongodb.com/docs/drivers/go/current/fundamentals/bson/
//
// Example:
//
//	type User struct {
//	    Name string `bson:"name"`
//	    Age  int    `bson:"age"`
//	}
//	bsonName := gmqb.Field[User]("Name") // returns "name"
func Field[T any](fieldPath string) string {
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	fields := getOrBuildFieldMap(t)
	bsonPath, ok := fields[fieldPath]
	if !ok {
		panic(fmt.Errorf("%w: field %q does not exist in struct %s", ErrInvalidField, fieldPath, t.Name()))
	}
	return bsonPath
}

// FieldRef wraps a resolved BSON field name as an aggregation field reference
// expression (prefixed with "$").
//
// See: https://www.mongodb.com/docs/manual/reference/operator/aggregation/
//
// Example:
//
//	ref := gmqb.FieldRef("age") // returns "$age"
func FieldRef(bsonFieldName string) string {
	return "$" + bsonFieldName
}

// getOrBuildFieldMap returns the cached field mapping for the given type,
// building it via reflection if not yet cached.
func getOrBuildFieldMap(t reflect.Type) map[string]string {
	if cached, ok := schemaCache.Load(t); ok {
		return cached.(map[string]string)
	}

	fields := make(map[string]string)
	buildFieldMap(t, "", "", fields)
	schemaCache.Store(t, fields)
	return fields
}

// buildFieldMap recursively inspects struct fields and maps Go field paths
// to their BSON tag names.
func buildFieldMap(t reflect.Type, goPrefix, bsonPrefix string, out map[string]string) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() {
			continue
		}

		goPath := sf.Name
		if goPrefix != "" {
			goPath = goPrefix + "." + sf.Name
		}

		bsonName := resolveBsonTag(sf)
		if bsonName == "-" {
			continue // explicitly excluded
		}

		bsonPath := bsonName
		if bsonPrefix != "" {
			bsonPath = bsonPrefix + "." + bsonName
		}

		out[goPath] = bsonPath

		// Recurse into nested structs
		ft := sf.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
		if ft.Kind() == reflect.Struct && ft.String() != "time.Time" &&
			!strings.HasPrefix(ft.PkgPath(), "go.mongodb.org") {
			buildFieldMap(ft, goPath, bsonPath, out)
		}
	}
}

// resolveBsonTag extracts the BSON field name from a struct field's bson tag.
// Falls back to the Go field name (lowercased) if no tag is present.
func resolveBsonTag(sf reflect.StructField) string {
	tag := sf.Tag.Get("bson")
	if tag == "" {
		return strings.ToLower(sf.Name)
	}
	name, _, _ := strings.Cut(tag, ",")
	if name == "" {
		return strings.ToLower(sf.Name)
	}
	return name
}
