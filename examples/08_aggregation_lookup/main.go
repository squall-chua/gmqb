package main

import (
	"fmt"

	"github.com/squall-chua/gmqb"
)

func main() {
	pipeline := gmqb.NewPipeline().
		Lookup(gmqb.LookupOpts{From: "inventory", LocalField: "item", ForeignField: "sku", As: "inventory_docs"}).
		Unwind("inventory_docs")

	fmt.Println("Lookup and Unwind Aggregation:")
	fmt.Println(pipeline.JSON())
}
