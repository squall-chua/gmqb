# gmqb — Go MongoDB Query Builder

A type-safe, functional MongoDB query builder for Go that generates `bson.D` output compatible with `go.mongodb.org/mongo-driver/v2` and supports JSON string serialization.

## Motivation

Building MongoDB queries with raw `bson.D` is verbose, error-prone, and hard to read:

```go
// Raw bson.D — hard to maintain
filter := bson.D{
    {"$and", bson.A{
        bson.D{{"age", bson.D{{"$gte", 18}}}},
        bson.D{{"status", bson.D{{"$in", bson.A{"active", "pending"}}}}},
    }},
}
```

With gmqb, the same query becomes:

```go
filter := gmqb.And(
    gmqb.Gte("age", 18),
    gmqb.In("status", "active", "pending"),
)
```

## Features

- **Full MQL coverage** — Query predicates, update operators, 30+ aggregation pipeline stages, and ~120 expression operators
- **Type-safe CRUD** — Generic `Collection[T]` wrapper with typed results
- **Struct schema reflection** — Resolve BSON field names from Go struct tags
- **Immutable builders** — Thread-safe, no side effects
- **Chainable filter API** — Both standalone constructors and fluent method chaining
- **GeoJSON helpers** — `Point`, `LineString`, `Polygon` for geospatial queries
- **Pipeline stage helpers** — `GroupSpec`, `FillSpec`, `DensifySpec`, `SetWindowFieldsSpec`, etc.
- **JSON output** — Print any query as JSON for debugging
- **Functional options** — Clean API for find/update options
- **Zero sub-packages** — Single `import "github.com/squall-chua/gmqb"`

## Installation

```bash
go get github.com/squall-chua/gmqb
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "github.com/squall-chua/gmqb"

    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
)

type User struct {
    Name  string `bson:"name"`
    Age   int    `bson:"age"`
    Email string `bson:"email"`
}

func main() {
    // Connect
    client, _ := mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017"))
    coll := gmqb.Wrap[User](client.Database("mydb").Collection("users"))

    ctx := context.Background()

    // Build a filter
    f := gmqb.Field[User]
    filter := gmqb.And(
        gmqb.Gte(f("Age"), 18),
        gmqb.Regex(f("Email"), `@company\.com$`, "i"),
    )

    // Debug — print as JSON
    fmt.Println(filter.JSON())

    // Execute
    users, _ := coll.Find(ctx, filter,
        gmqb.WithSort(gmqb.Desc(f("Age"))),
        gmqb.WithLimit(10),
    )
    fmt.Println(users)
}
```

## API Overview

### Query Predicates (Filter)

Filters can be created with standalone constructors or chained fluently:

```go
// Standalone constructors
gmqb.Eq("field", value)          // $eq
gmqb.Ne("field", value)          // $ne
gmqb.Gt("field", value)          // $gt
gmqb.Gte("field", value)         // $gte
gmqb.Lt("field", value)          // $lt
gmqb.Lte("field", value)         // $lte
gmqb.In("field", v1, v2)         // $in
gmqb.Nin("field", v1, v2)        // $nin
gmqb.And(filters...)             // $and
gmqb.Or(filters...)              // $or
gmqb.Nor(filters...)             // $nor
gmqb.Not("field", filter)        // $not
gmqb.Exists("field", true)       // $exists
gmqb.Type("field", "string")     // $type
gmqb.Regex("field", "pat", "i")  // $regex
gmqb.Mod("field", 4, 0)          // $mod
gmqb.All("field", v1, v2)        // $all
gmqb.ElemMatch("field", filter)  // $elemMatch
gmqb.Size("field", 3)            // $size
gmqb.Expr(expression)            // $expr
gmqb.Where("js expression")      // $where
gmqb.JsonSchema(schema)          // $jsonSchema

// Chainable filter API (same operators available as methods)
filter := gmqb.NewFilter().
    Eq("status", "active").
    Gte("age", 18).
    In("role", "admin", "staff")
```

#### Geospatial Queries

```go
// Geometry helpers — no more manual bson.D construction
gmqb.Point(-73.9667, 40.78)                                        // GeoJSON Point
gmqb.LineString([2]float64{0, 0}, [2]float64{1, 1})                // GeoJSON LineString
gmqb.Polygon([][2]float64{{0, 0}, {3, 6}, {6, 1}, {0, 0}})         // GeoJSON Polygon

// Geospatial operators
gmqb.GeoIntersects("location", gmqb.Polygon(...))
gmqb.GeoWithin("location", gmqb.Polygon(...))
gmqb.Near("location", gmqb.Point(-73.9, 40.7), 5000, 100)          // maxDist, minDist
gmqb.NearSphere("location", gmqb.Point(-73.9, 40.7), 5000, 100)

// Bitwise operators
gmqb.BitsAllClear("field", mask)
gmqb.BitsAllSet("field", mask)
gmqb.BitsAnyClear("field", mask)
gmqb.BitsAnySet("field", mask)
```

### Update Operators

```go
update := gmqb.NewUpdate().
    Set("field", value).              // $set
    Unset("field").                   // $unset
    Inc("field", 1).                  // $inc
    Mul("field", 1.5).                // $mul
    Min("field", value).              // $min
    Max("field", value).              // $max
    Rename("old", "new").             // $rename
    CurrentDate("field").             // $currentDate
    CurrentDateAsTimestamp("field").   // $currentDate (timestamp)
    SetOnInsert("field", value).      // $setOnInsert
    Push("arr", value).               // $push
    AddToSet("arr", value).           // $addToSet
    AddToSetEach("arr", v1, v2).      // $addToSet + $each
    Pop("arr", 1).                    // $pop
    Pull("arr", condition).           // $pull
    PullAll("arr", v1, v2).           // $pullAll
    PushWithOpts("arr", PushOpts{}).  // $push + modifiers ($each, $sort, $slice, $position)
    BitAnd("field", mask).            // $bit (and)
    BitOr("field", mask).             // $bit (or)
    BitXor("field", mask)             // $bit (xor)
```

### Aggregation Pipeline

```go
pipeline := gmqb.NewPipeline().
    Match(filter).                         // $match
    Group(gmqb.GroupSpec(                   // $group — with helpers
        gmqb.GroupID("$country", "$city"),
        gmqb.GroupAcc("total", gmqb.AccSum(1)),
        gmqb.GroupAcc("avgAge", gmqb.AccAvg("$age")),
    )).
    Project(spec).                         // $project
    Sort(gmqb.Desc("total")).              // $sort
    Limit(10).                             // $limit
    Skip(20).                              // $skip
    Unwind("$field").                       // $unwind
    Lookup(gmqb.LookupOpts{...}).          // $lookup
    LookupPipeline(gmqb.LookupPipelineOpts{...}). // $lookup (sub-pipeline)
    AddFields(gmqb.AddFieldsSpec(          // $addFields — with helpers
        gmqb.AddField("fullName", gmqb.ExprConcat("$first", " ", "$last")),
    )).
    Unset("field1", "field2").             // $unset
    Count("total").                        // $count
    Facet(facets).                         // $facet
    Bucket(gmqb.BucketOpts{...}).          // $bucket
    BucketAuto(gmqb.BucketAutoOpts{...}).  // $bucketAuto
    GraphLookup(gmqb.GraphLookupOpts{...}).// $graphLookup
    GeoNear(gmqb.GeoNearOpts{...}).        // $geoNear
    Fill(gmqb.FillSpec(                    // $fill — with helpers
        gmqb.FillOutput("qty", gmqb.FillMethod("linear")),
    )).
    Densify(gmqb.DensifySpec("ts",         // $densify — with helpers
        gmqb.DensifyRange(1, "hour", "full"),
    )).
    SetWindowFields(gmqb.SetWindowFieldsSpec( // $setWindowFields — with helpers
        "$state",
        gmqb.SortSpec(gmqb.SortRule("date", 1)),
        gmqb.WindowOutput("cumQty", gmqb.AccSum("$qty"),
            gmqb.Window("documents", "unbounded", "current")),
    )).
    Out("collection").                     // $out
    RawStage("$custom", value)             // any custom stage
```

### Typed CRUD

```go
coll := gmqb.Wrap[User](db.Collection("users"))

users, err := coll.Find(ctx, filter, gmqb.WithLimit(10))
user, err := coll.FindOne(ctx, filter)
res, err := coll.InsertOne(ctx, &user)
res, err := coll.UpdateOne(ctx, filter, update, gmqb.WithUpsert(true))
res, err := coll.DeleteMany(ctx, filter)
count, err := coll.CountDocuments(ctx, filter)
```

### JSON Output

```go
filter := gmqb.And(gmqb.Gte("age", 18), gmqb.Eq("active", true))
fmt.Println(filter.JSON())        // pretty-printed
fmt.Println(filter.CompactJSON()) // compact
```

## Examples

The `examples/` directory contains 13 runnable programs demonstrating every major feature:

| Example | Feature |
|---------|---------|
| `01_basic_find` | Simple filter with comparison operators |
| `02_complex_filter` | Logical, regex, and element operators |
| `03_geospatial` | `Point`, `Near` geospatial queries |
| `04_array_queries` | `ElemMatch`, `Size` |
| `05_update_fields` | `Set`, `Inc`, `Unset`, `CurrentDateAsTimestamp` |
| `06_update_arrays` | `AddToSet`, `PushWithOpts` with sort/slice |
| `07_aggregation_basic` | `GroupSpec`, `GroupAcc`, `AccSum`, `AccAvg` |
| `08_aggregation_lookup` | `LookupOpts`, `Unwind` |
| `09_aggregation_facet` | `Facet` with sub-pipelines |
| `10_aggregation_window` | `SetWindowFieldsSpec`, `WindowOutput`, `Window` |
| `11_expressions` | `ExprCond`, `ExprMultiply`, `AddFieldsSpec` |
| `12_crud_generics` | Typed `Collection[T]` CRUD patterns |
| `13_json_output` | `JSON()` and `CompactJSON()` serialization |

## Documentation

Every exported function includes GoDoc with:
1. Operator behavior description
2. MongoDB documentation link
3. Code example

Run `go doc github.com/squall-chua/gmqb`.

## License

MIT
