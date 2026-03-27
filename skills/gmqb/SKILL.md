---
name: gmqb
description: "This skill should be used when the user asks about the gmqb Go MongoDB Query Builder library, how to write filters/updates/pipelines with gmqb, how to use the code generator, or how to look up gmqb documentation."
category: library
risk: safe
source: local
tags: "[go, mongodb, query-builder, gmqb, code-generator]"
date_added: "2026-03-27"
---

# gmqb — Go MongoDB Query Builder

## Purpose

Guide for using the `gmqb` library (`github.com/squall-chua/gmqb`), checking its documentation, and converting raw MongoDB JSON into type-safe Go code via the built-in code generator.

## When to Use This Skill

Activate when the user:
- Asks how to write a MongoDB filter, update, or aggregation pipeline using gmqb
- Wants to convert a raw `bson.D` query or a MongoDB Compass JSON snippet into gmqb Go code
- Needs to look up a specific operator or helper (e.g. `$elemMatch`, `$setWindowFields`)
- Asks about the query cache, Change Stream invalidation, or `Collection[T]` typed CRUD
- Wants to run the `gmqb-gen` CLI tool

---

## 1. Installation

```bash
go get github.com/squall-chua/gmqb
```

Single import — no sub-packages needed for filters, updates, pipelines, or CRUD:

```go
import "github.com/squall-chua/gmqb"
```

The code generator lives in a separate sub-package:

```go
import "github.com/squall-chua/gmqb/generator"
```

---

## 2. Reading the Documentation

### Online (pkg.go.dev)

Full GoDoc with operator descriptions, MongoDB links, and code examples:

```
https://pkg.go.dev/github.com/squall-chua/gmqb
```

### Local (go doc)

Run from any directory that has gmqb as a dependency:

```bash
# Top-level package overview
go doc github.com/squall-chua/gmqb

# Specific function
go doc github.com/squall-chua/gmqb.ElemMatch
go doc github.com/squall-chua/gmqb.NewPipeline

# Generator sub-package
go doc github.com/squall-chua/gmqb/generator
go doc github.com/squall-chua/gmqb/generator.Generate
```

### In-Repo (source)

The repository root is `/home/mwchua/gmqb`. Key files by concern:

| File | What it covers |
|------|---------------|
| `filter.go` | All query predicate constructors (`Eq`, `Gt`, `In`, `Regex`, geospatial, …) |
| `update.go` | All update operators (`Set`, `Inc`, `Push`, `PushWithOpts`, …) |
| `pipeline.go` | All 30+ aggregation pipeline stages |
| `expr_core.go` / `expr_data.go` / `expr_acc.go` | ~120 aggregation expression operators |
| `collection.go` | Generic `Collection[T]` typed CRUD and bulk write |
| `cache.go` | `CachedCollection`, `WrapWithCache`, `WrapWithCacheAndMetrics` |
| `cache_invalidator.go` | `ChangeStreamInvalidator` — auto-invalidation via Change Streams |
| `schema.go` | `Field[T]` — struct-tag BSON field resolver |
| `options.go` | Functional options (`WithLimit`, `WithSort`, `WithUpsert`, …) |
| `generator/generate.go` | `Generate`, `GenerateFilter`, `GeneratePipeline` |

Browse runnable examples in `examples/` (17 programs, one per feature area).

---

## 3. Core API

### 3.1 Filters

```go
// Standalone constructors
gmqb.Eq("field", value)
gmqb.Gte("age", 18)
gmqb.In("status", "active", "pending")
gmqb.And(gmqb.Gte("age", 18), gmqb.Eq("active", true))
gmqb.Or(gmqb.Eq("role", "admin"), gmqb.Eq("role", "staff"))
gmqb.Not("age", gmqb.Lt("age", 18))
gmqb.ElemMatch("scores", gmqb.Gte("score", 80))
gmqb.Regex("email", `@company\.com$`, "i")
gmqb.Expr(gmqb.ExprGt("$qty", "$minQty"))   // $expr

// Fluent/chainable
filter := gmqb.NewFilter().
    Eq("status", "active").
    Gte("age", 18).
    In("role", "admin", "staff")

// BSON-aware field resolver (avoids hardcoding field names)
f := gmqb.Field[User]         // User has bson struct tags
filter := gmqb.Gte(f("Age"), 18)

// Debug output
fmt.Println(filter.JSON())        // pretty
fmt.Println(filter.CompactJSON()) // compact
```

### 3.2 Updates

```go
update := gmqb.NewUpdate().
    Set("status", "inactive").
    Inc("loginCount", 1).
    Unset("tempField").
    Push("tags", "go").
    PushWithOpts("scores", gmqb.PushOpts{
        Each:  []any{95, 87},
        Sort:  gmqb.Desc("score"),
        Slice: 10,
    })
```

### 3.3 Aggregation Pipeline

```go
pipeline := gmqb.NewPipeline().
    Match(gmqb.Eq("status", "active")).
    Group(gmqb.GroupSpec(
        gmqb.GroupID("$country"),
        gmqb.GroupAcc("total", gmqb.AccSum(1)),
        gmqb.GroupAcc("avgAge", gmqb.AccAvg("$age")),
    )).
    Sort(gmqb.Desc("total")).
    Limit(10)
```

### 3.4 Typed CRUD (`Collection[T]`)

```go
coll := gmqb.Wrap[User](db.Collection("users"))

users, err  := coll.Find(ctx, filter, gmqb.WithLimit(10))
user, err   := coll.FindOne(ctx, filter)
res, err    := coll.InsertOne(ctx, &newUser)
res, err    := coll.UpdateOne(ctx, filter, update)
res, err    := coll.DeleteMany(ctx, filter)
count, err  := coll.CountDocuments(ctx, filter)

// Atomic compound ops
deleted, _  := coll.FindOneAndDelete(ctx, filter)
updated, _  := coll.FindOneAndUpdate(ctx, filter, update, gmqb.WithReturnDocument(options.After))
```

### 3.5 Query Cache

```go
import (
    gocache      "github.com/patrickmn/go-cache"
    gocachestore "github.com/eko/gocache/store/go_cache/v4"
    "github.com/eko/gocache/lib/v4/cache"
)

store        := gocachestore.NewGoCache(gocache.New(5*time.Minute, 10*time.Minute))
cacheManager := cache.New[[]byte](store)
cachedUsers  := gmqb.WrapWithCache[User](rawColl, cacheManager, time.Minute)

// Reads hit cache; writes bypass it automatically.
user, _ := cachedUsers.FindOne(ctx, gmqb.Eq("email", "alice@example.com"))

// Manual flush
_ = cachedUsers.InvalidateCache(ctx)

// Auto-invalidation via Change Streams (replica set / Atlas required)
inv := gmqb.NewChangeStreamInvalidator(rawColl, cacheManager)
go inv.Watch(ctx)
```

---

## 4. Code Generator

Convert raw MongoDB JSON — from MongoDB Compass, a dbshell query, or any BSON dump — directly into equivalent gmqb Go code.

### 4.1 Programmatic

```go
import "github.com/squall-chua/gmqb/generator"

// Auto-detects: object → Filter, array → Pipeline
code, err := generator.Generate(jsonStr)

// Explicit variants
filterCode, err   := generator.GenerateFilter(jsonStr)
pipelineCode, err := generator.GeneratePipeline(jsonStr)
```

**Filter example:**

```go
input := `{"$and": [{"age": {"$gte": 18}}, {"status": {"$in": ["active"]}}]}`
code, _ := generator.Generate(input)
// gmqb.And(
//     gmqb.Gte("age", 18),
//     gmqb.In("status", "active"),
// )
```

**Pipeline example:**

```go
input := `[{"$match": {"status": "A"}}, {"$limit": 10}]`
code, _ := generator.Generate(input)
// gmqb.NewPipeline().
//     Match(gmqb.Eq("status", "A")).
//     Limit(10)
```

### 4.2 CLI Tool (`gmqb-gen`)

Install once:

```bash
go install github.com/squall-chua/gmqb/cmd/gmqb-gen@latest
```

Pipe JSON from stdin:

```bash
echo '[{"$match": {"status": "active"}}, {"$limit": 10}]' | gmqb-gen
```

Pass via flag:

```bash
gmqb-gen -query='{"age": {"$gte": 18}}'
```

The CLI prints the generated gmqb code to stdout and exits non-zero on parse errors.

---

## 5. Runnable Examples

Located in `/home/mwchua/gmqb/examples/`. Each sub-directory is a self-contained `package main`:

| Directory | Feature demonstrated |
|-----------|---------------------|
| `01_basic_find` | Comparison operators (`Eq`, `Gte`, `Lt`) |
| `02_complex_filter` | Logical, regex, element operators |
| `03_geospatial` | `Point`, `Near`, `GeoWithin` |
| `04_array_queries` | `ElemMatch`, `All`, `Size` |
| `05_update_fields` | `Set`, `Inc`, `Unset`, `CurrentDateAsTimestamp` |
| `06_update_arrays` | `AddToSet`, `PushWithOpts` |
| `07_aggregation_basic` | `GroupSpec`, `AccSum`, `AccAvg` |
| `08_aggregation_lookup` | `LookupOpts`, `Unwind` |
| `09_aggregation_facet` | `Facet` with sub-pipelines |
| `10_aggregation_window` | `SetWindowFieldsSpec`, `WindowOutput` |
| `11_expressions` | `ExprCond`, `ExprMultiply`, `AddFieldsSpec` |
| `12_crud_generics` | Typed `Collection[T]` CRUD |
| `13_json_output` | `JSON()` / `CompactJSON()` |
| `14_bulk_write` | `WriteModel[T]`, `BulkWrite` |
| `15_compound_operations` | `FindOneAndDelete`, `FindOneAndUpdate`, `FindOneAndReplace` |
| `16_query_cache_basic` | In-memory cache with `go-cache` |
| `17_query_cache_invalidation` | Change Stream auto-invalidation |

Run any example (requires a running MongoDB):

```bash
cd /home/mwchua/gmqb/examples/01_basic_find
go run main.go
```

---

## 6. Quick Diagnostic Checklist

When helping the user with gmqb queries, check:

1. **Import path** — `github.com/squall-chua/gmqb` (not `gmqb/v2`)
2. **Mongo driver version** — gmqb targets `go.mongodb.org/mongo-driver/v2`
3. **Field names** — Use `gmqb.Field[T]` resolver if the user's struct has `bson:"..."` tags, to avoid field name mismatches
4. **Filter vs Update** — `coll.UpdateOne` takes `(ctx, filter, update)` — both are separate builder calls
5. **Pipeline slices** — `Pipeline.BSON()` returns `bson.A`; pass it directly to `mongo.Collection.Aggregate`
6. **Cache TTL** — `WrapWithCache` TTL applies per-query key; Change Stream invalidation fires on write
7. **Generator input** — `generator.Generate` expects valid MongoDB Extended JSON; plain JSON values (no `$`-operators) are also valid
