package main

import (
	"fmt"
)

func main() {
	fmt.Println("CRUD Generics Example:")
	fmt.Println(`
// 1. Wrap a literal mongo.Collection into a gmqb.Collection
// coll := gmqb.Wrap[User](mongoCollection)

// 2. Use gmqb's Find with strong typing
// filter := gmqb.Eq("status", "active")
// users, err := coll.Find(ctx, filter)
//
// users is typed as []User!

// 3. For custom projections:
// type NameOnly struct { Name string }
// names, err := gmqb.Aggregate[NameOnly](coll, ctx, pipeline)
	`)
}
