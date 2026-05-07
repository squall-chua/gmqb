---
name: gmqb
description: "Comprehensive guide to the gmqb Go MongoDB Query Builder library. Use this for building type-safe filters, updates, aggregation pipelines, and expressions. Covers typed CRUD, indexing, caching, and code generation."
category: library
risk: safe
source: local
tags: "[go, mongodb, query-builder, gmqb, code-generator, type-safe]"
date_added: "2026-03-27"
---

# gmqb — Go MongoDB Query Builder

## Purpose

The `gmqb` library provides a type-safe, fluent API for building MongoDB queries (filters, updates, aggregations) in Go. It eliminates common BSON syntax errors, provides generic CRUD wrappers, and includes a code generator to convert raw MongoDB JSON into Go code.

## 1. Getting Started

### Installation

```bash
go get github.com/squall-chua/gmqb
```

### Core Import

```go
import "github.com/squall-chua/gmqb"
```

---

## 2. Type-Safe Field Resolution

Avoid hardcoding strings for field names. Use `gmqb.Field[T]` to resolve field names from struct tags.

```go
type User struct {
    ID    string `bson:"_id"`
    Email string `bson:"email"`
    Profile struct {
        Age int `bson:"age"`
    } `bson:"profile"`
}

f := gmqb.Field[User]
filter := gmqb.Eq(f("Email"), "alice@example.com")

// Nested fields:
ageFilter := gmqb.Gte(f("Profile", "Age"), 18) // "profile.age"
```

---

## 3. Query Filters

Build complex predicates using functional constructors or the fluent `FilterBuilder`.

### Basic Operators

- **Comparison**: `Eq`, `Ne`, `Gt`, `Gte`, `Lt`, `Lte`, `In`, `Nin`
- **Logical**: `And`, `Or`, `Not`, `Nor`
- **Element**: `Exists`, `Type`
- **Evaluation**: `Regex`, `Mod`, `Text`, `Where`

### Filter Array Operators

```go
// Match array elements
gmqb.All("tags", "go", "mongodb")
gmqb.Size("comments", 5)
gmqb.ElemMatch("scores", gmqb.Gte("score", 80))
```

### Geospatial Operators

```go
gmqb.Near("location", 100.0, 10.0, gmqb.NearOpts{MaxDistance: 5000})
gmqb.GeoWithinCenter("location", 100.0, 10.0, 5.0)
```

### Fluent Filtering

```go
filter := gmqb.NewFilter().
    Eq("status", "active").
    In("tags", "tech", "news").
    Gte("views", 1000)
```

---

## 4. Update Operations

MongoDB updates can be performed using either standard update operators (via `Updater`) or aggregation pipelines (via `Pipeline`). Both implement the `UpdateDoc` interface and can be passed to `UpdateOne`, `UpdateMany`, and `FindOneAndUpdate`.

### 4.1 Updating with Operators (`Updater`)

Use `gmqb.NewUpdate()` for a fluent, chainable API.

#### Field Operators

- **Value**: `Set`, `Unset`, `Inc`, `Mul`, `Min`, `Max`, `Rename`, `SetOnInsert`
- **Date**: `CurrentDate`, `CurrentDateAsTimestamp`
- **Bitwise**: `BitAnd`, `BitOr`, `BitXor`

#### Update Array Operators

```go
update := gmqb.NewUpdate().
    Push("history", "login").
    AddToSet("tags", "new-tag").
    PopFirst("queue").
    Pull("blacklist", "bad-user")

// Advanced Push with modifiers ($each, $sort, $slice)
gmqb.NewUpdate().PushWithOpts("scores", gmqb.PushOpts{
    Each:  []any{90, 80},
    Sort:  gmqb.Desc("score"),
    Slice: 10,
})
```

### 4.2 Updating with Aggregation Pipelines

You can pass a `Pipeline` as an update to perform complex transformations (e.g., setting fields based on other fields).

```go
pipeline := gmqb.NewPipeline().
    SetFields(gmqb.AddFieldsSpec(
        gmqb.AddField("total", gmqb.ExprAdd("$price", "$tax")),
    ))

// Pass pipeline directly to UpdateOne
coll.UpdateOne(ctx, filter, pipeline)
```

### 4.3 Upsert Operations

Upserts create a new document if no document matches the filter.

- **Option**: Use `gmqb.WithUpsert(true)` with `UpdateOne/Many`.
- **Convenience**: Use `coll.UpsertOne(ctx, filter, update)`.
- **Operator**: `SetOnInsert` is only applied during an insert.

```go
// Using UpsertOne convenience
coll.UpsertOne(ctx, gmqb.Eq("email", "new@user.com"), 
    gmqb.NewUpdate().
        Set("status", "active").
        SetOnInsert("createdAt", time.Now()))
```

---

## 5. Aggregation Pipelines

Build pipelines with type-safe stage helpers.

### Common Stages

`Match`, `Project`, `Group`, `Sort`, `Limit`, `Skip`, `Unwind`, `Lookup`, `Facet`, `Count`, `Merge`, `Out`, `ReplaceRoot`, `SetWindowFields`.

### Complex Stages

```go
pipeline := gmqb.NewPipeline().
    Match(gmqb.Eq("active", true)).
    Lookup(gmqb.LookupOpts{
        From: "orders",
        LocalField: "_id",
        ForeignField: "userId",
        As: "user_orders",
    }).
    Group(gmqb.GroupSpec(
        gmqb.GroupID("$category"),
        gmqb.GroupAcc("avgPrice", gmqb.AccAvg("$price")),
    )).
    Facet(gmqb.FacetSpec(
        gmqb.FacetPipeline("topDocs", gmqb.NewPipeline().Sort(gmqb.Desc("score")).Limit(5)),
        gmqb.FacetPipeline("stats", gmqb.NewPipeline().Count("total")),
    ))
```

---

## 6. Aggregation Expressions

Used within `$project`, `$addFields`, `$group`, etc. (Found in `expr_*.go`).

- **Arithmetic**: `ExprAdd`, `ExprSubtract`, `ExprMultiply`, `ExprDivide`, `ExprAbs`, `ExprFloor`, `ExprCeil`, `ExprRound`
- **Comparison**: `ExprEq`, `ExprNe`, `ExprGt`, `ExprGte`, `ExprLt`, `ExprLte`, `ExprCmp`
- **Logical**: `ExprBoolAnd`, `ExprBoolOr`, `ExprBoolNot`
- **Conditional**: `ExprCond`, `ExprIfNull`, `ExprSwitch`
- **String**: `ExprConcat`, `ExprSubstr`, `ExprToUpper`, `ExprToLower`, `ExprTrim`
- **Array**: `ExprFilter`, `ExprMap`, `ExprReduce`, `ExprSize`, `ExprArrayElemAt`
- **Date**: `ExprDateToString`, `ExprYear`, `ExprMonth`, `ExprDayOfMonth`
- **Accumulators**: `AccSum`, `AccAvg`, `AccMin`, `AccMax`, `AccPush`, `AccAddToSet`, `AccFirst`, `AccLast`

---

## 7. Typed CRUD (`Collection[T]`)

Wrap a raw `mongo.Collection` to get a typed API that accepts `gmqb` builders.

```go
coll := gmqb.Wrap[User](rawColl)

// Read
users, _ := coll.Find(ctx, gmqb.Eq("status", "active"), gmqb.WithLimit(10))
user, _  := coll.FindOne(ctx, gmqb.Eq("_id", "123"))
count, _ := coll.CountDocuments(ctx, filter, gmqb.WithLimitCount(100))

// Write
coll.InsertOne(ctx, &newUser)
coll.UpdateOne(ctx, gmqb.Eq("_id", "123"), gmqb.NewUpdate().Set("name", "Bob"))
coll.ReplaceOne(ctx, gmqb.Eq("_id", "123"), &replacementUser)

// Atomic (FindOneAnd...)
updated, _ := coll.FindOneAndUpdate(ctx, filter, update, gmqb.WithReturnDocument(options.After))
deleted, _ := coll.FindOneAndDelete(ctx, filter, gmqb.WithReturnDocumentDelete(options.Before))
replaced, _ := coll.FindOneAndReplace(ctx, filter, &replacement, gmqb.WithReturnDocumentReplace(options.After))

// Aggregate (returns []R)
stats, _ := gmqb.Aggregate[UserStats](coll, ctx, pipeline)
```

### Bulk Write

```go
models := []gmqb.WriteModel[User]{
    gmqb.NewInsertOneModel[User]().SetDocument(&User{Name: "Alice"}),
    gmqb.NewUpdateOneModel[User]().SetFilter(gmqb.Eq("name", "Bob")).SetUpdate(gmqb.NewUpdate().Set("age", 25)),
}
res, err := coll.BulkWrite(ctx, models)
```

---

## 8. Index Management

Fluent builder for creating and managing indexes.

```go
// Create Unique Index
coll.CreateIndex(ctx, gmqb.NewIndex(gmqb.Asc("email")).Unique().Name("idx_email"))

// TTL Index
coll.CreateIndex(ctx, gmqb.NewIndex(gmqb.Asc("createdAt")).TTL(3600))

// List & Drop
indexes, _ := coll.ListIndexes(ctx)
_ = coll.DropIndex(ctx, "idx_email")
```

---

## 9. Query Caching

Wrap collections with `WrapWithCache` to enable automatic query caching using `go-cache` or any compatible store.

```go
cachedColl := gmqb.WrapWithCache[User](rawColl, cacheManager, 5*time.Minute)

// Reads hit cache. Writes invalidate or bypass as configured.
user, _ := cachedColl.FindOne(ctx, gmqb.Eq("email", "alice@example.com"))

// Auto-invalidation via Change Streams
inv := gmqb.NewChangeStreamInvalidator(rawColl, cacheManager)
go inv.Watch(ctx)
```

---

## 10. Pub/Sub (`pubsub` package)

Type-safe pub/sub powered by MongoDB capped collections and tailable cursors.

```go
import "github.com/squall-chua/gmqb/pubsub"

// 1. Initialize the bus
bus, _ := pubsub.NewTailablePubSub[MyEvent](db, "events_topic", pubsub.DefaultCappedOpts)

// 2. Subscribe
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

events, stop := bus.Subscribe(ctx)
defer stop()

go func() {
    for event := range events {
        fmt.Printf("Received: %+v\n", event)
    }
}()

// 3. Publish
err := bus.Publish(ctx, MyEvent{ID: "evt_123"})
```

---

## 11. Message Queue (`mq` package)

Durable message queue built on MongoDB with load-balancing, fan-out, DLQ, and delivery guarantees.

```go
import "github.com/squall-chua/gmqb/mq"

// 1. Initialize
q, _ := mq.NewQueue[MyMessage](db, "email_tasks", mq.QueueOpts{
    MaxAttempts: 3,
})

// 2. Enqueue (with idempotency)
q.Enqueue(ctx, MyMessage{To: "user@example.com"}, 
    mq.WithID(customID), mq.WithIdempotent())

// 3. Process
handler := func(ctx context.Context, msg MyMessage) error {
    return nil // return err to nack
}

worker := mq.NewWorker(q, handler,
    mq.WithConsumerGroup("audit_service"), // Enables Fan-Out if not empty
    mq.WithDelivery(mq.ExactlyOnce),       // AtMostOnce, AtLeastOnce, ExactlyOnce
)

worker.Run(ctx)

// 4. DLQ Management
dlq := q.DLQ()
messages, _ := dlq.List(ctx, 10, 0)
dlq.Requeue(ctx, messages[0].ID)
```

---

## 12. REST API (`rest` package)

Zero-boilerplate REST API that wraps any `gmqb.Collection[T]` and exposes standard CRUD endpoints as a framework-agnostic `http.Handler`.

```go
import "github.com/squall-chua/gmqb/rest"

// 1. Define the resource
users := rest.NewResource[User, string](coll, rest.Config[User, string]{
    IDField:  "_id",
    IDParser: func(s string) (string, error) { return s, nil },
    FilterableFields: []rest.FilterField{
        {Name: "status", BsonKey: "status", Op: rest.OpEq},
        {Name: "age",    BsonKey: "age",    Op: rest.OpGte | rest.OpLte},
    },
    SortableFields: []string{"createdAt", "age"},
})

// 2. Mount to any router (supports Go 1.22+ PathValue)
mux := http.NewServeMux()
mux.Handle("/users", users)      // Collection (List, Create)
mux.Handle("/users/{id}", users) // Resource (Read, Update, Delete)
```

Supported features:

- **Filtering**: Allowlist-based filtering via query params. Supports both `field_op=v` and `field[op]=v` (e.g., `?age[gte]=18`).
- **Sorting**: Allowlist-based sorting (`?sort=-createdAt,age`).
- **Pagination**: Support for `offset/limit` and `cursor/limit`. The `?cursor=` parameter is direction-aware (supports both next and prev tokens).
- **Lifecycle Hooks**: Complete set of `Before` and `After` hooks for all mutating operations (`Create`, `Replace`, `Update`, `Delete`).

---

## 13. Code Generator

Convert MongoDB JSON into gmqb Go code.

### CLI

```bash
echo '{"status": "active", "age": {"$gte": 18}}' | gmqb-gen
```

### Programmatic

```go
import "github.com/squall-chua/gmqb/generator"

code, _ := generator.Generate(`{"$match": {"status": "A"}}`)
// Output: gmqb.NewPipeline().Match(gmqb.Eq("status", "A"))
```

---

## 13. Functional Options

Commonly used options across `Find`, `Update`, `BulkWrite`, etc.

- **Query**: `WithLimit`, `WithSkip`, `WithSort`, `WithProjection`
- **Count**: `WithLimitCount`, `WithSkipCount`
- **Update/Replace**: `WithUpsert`, `WithUpsertReplace`, `WithCollation`, `WithHint`
- **Bulk**: `WithOrdered`
- **FindOneAndXXX**:
  - **Delete**: `WithSortFindAndDelete`, `WithProjectionFindAndDelete`, `WithReturnDocumentDelete`
  - **Update**: `WithSortFindAndUpdate`, `WithReturnDocument`, `WithUpsertFindAndUpdate`
  - **Replace**: `WithSortFindAndReplace`, `WithReturnDocumentReplace`, `WithUpsertFindAndReplace`

---

## 14. Summary Checklist for Assistance

1. **Type-Safety**: Did you use `gmqb.Field[T]`?
2. **Pipelines**: Is it a `Filter` or a `Pipeline`? Use `Match` in pipelines.
3. **Updates**: Are you using `UpdateBuilder` for `UpdateOne/Many`?
4. **Caching**: If caching is enabled, ensure `Watch()` is running for invalidation.
5. **Pub/Sub**: Use `TailablePubSub` for lightweight event-driven needs.
6. **Message Queue**: Use `Queue[T]` for durable background processing with delivery guarantees.
7. **REST API**: Use `rest.NewResource` to instantly expose a collection over HTTP with safe filtering and pagination.
8. **Generator**: Use `gmqb-gen` to quickly port existing shell queries.
