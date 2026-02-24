package main

import (
	"fmt"

	"github.com/squall-chua/gmqb"
)

func main() {
	// Process documents through multiple sub-pipelines simultaneously
	pipeline := gmqb.NewPipeline().Facet(map[string]gmqb.Pipeline{
		"categorizedByTags": gmqb.NewPipeline().
			Unwind("tags").
			Group(gmqb.GroupSpec(gmqb.GroupID("$tags"), gmqb.GroupAcc("count", gmqb.AccSum(1)))).
			Sort(gmqb.SortSpec(gmqb.SortRule("count", -1))),

		"categorizedByPrice": gmqb.NewPipeline().
			MatchRaw(gmqb.Exists("price", true).BsonD()).
			BucketAuto(gmqb.BucketAutoOpts{GroupBy: "$price", Buckets: 4}),
	})

	fmt.Println("Facet Aggregation:")
	fmt.Println(pipeline.JSON())
}
