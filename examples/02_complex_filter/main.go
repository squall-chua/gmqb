package main

import (
	"fmt"

	"github.com/squall-chua/gmqb"
)

type Product struct {
	Sku      string   `bson:"sku"`
	Tags     []string `bson:"tags"`
	Metadata struct {
		Weight int `bson:"weight"`
	} `bson:"metadata"`
}

func main() {
	f := gmqb.Field[Product]

	filter := gmqb.Or(
		gmqb.Regex(f("Sku"), "^ABC", "i"),
		gmqb.And(
			gmqb.In(f("Tags"), "sale", "clearance"),
			gmqb.Exists(f("Metadata"), true),
			gmqb.Gt(f("Metadata.Weight"), 50),
		),
	)

	fmt.Println("Complex Filter:")
	fmt.Println(filter.JSON())
}
