package main

import (
	"fmt"

	"github.com/squall-chua/gmqb"
)

func main() {
	pipeline := gmqb.NewPipeline().
		MatchRaw(gmqb.Eq("status", "active").BsonD()).
		Group(gmqb.GroupSpec(
			gmqb.GroupID("$state", "$city"),
			gmqb.GroupAcc("totalValue", gmqb.AccSum("$amount")),
			gmqb.GroupAcc("avgScore", gmqb.AccAvg("$score")),
		)).
		Sort(gmqb.SortSpec(gmqb.SortRule("totalValue", -1))).
		Limit(10)

	fmt.Println("Basic Aggregation:")
	fmt.Println(pipeline.JSON())
}
