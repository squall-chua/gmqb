package main

import (
	"fmt"

	"github.com/squall-chua/gmqb"
)

type User struct {
	Name string `bson:"name"`
	Age  int    `bson:"age"`
}

func main() {
	f := gmqb.Field[User]

	filter := gmqb.And(
		gmqb.Eq(f("Name"), "Alice"),
		gmqb.Gte(f("Age"), 18),
	)

	fmt.Println("Basic Find Filter:")
	fmt.Println(filter.JSON())
}
