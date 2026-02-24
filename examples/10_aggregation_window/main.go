package main

import (
	"fmt"

	"github.com/squall-chua/gmqb"
)

func main() {
	pipeline := gmqb.NewPipeline().SetWindowFields(gmqb.SetWindowFieldsSpec(
		"$state", // partitionBy
		gmqb.SortSpec(gmqb.SortRule("orderDate", 1)),
		gmqb.WindowOutput(
			"cumulativeQuantity",
			gmqb.AccSum("$quantity"),
			gmqb.Window("documents", "unbounded", "current"),
		),
	))

	fmt.Println("SetWindowFields Aggregation:")
	fmt.Println(pipeline.JSON())
}
