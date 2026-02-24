package main

import (
	"fmt"

	"github.com/squall-chua/gmqb"
)

func main() {
	// Calculate a discount conditionally using aggregation expressions
	condExpr := gmqb.ExprCond(
		gmqb.ExprGte("$totalSpent", 1000),
		0.20, // 20% off
		0.05, // 5% off
	)

	pipeline := gmqb.NewPipeline().
		AddFields(gmqb.AddFieldsSpec(
			gmqb.AddField("discountRate", condExpr),
			gmqb.AddField("finalPrice", gmqb.ExprMultiply("$price", gmqb.ExprSubtract(1, "$discountRate"))),
		))

	fmt.Println("Expressions Aggregation:")
	fmt.Println(pipeline.JSON())
}
