package main

import (
	"fmt"
	"github.com/squall-chua/gmqb"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func main() {
	// Add an item to a set (only if it doesn't already exist)
	u1 := gmqb.NewUpdate().AddToSet("tags", "premium")

	slice := -5
	// Pushing multiple items with a sort and slice applied
	u2 := gmqb.NewUpdate().PushWithOpts("scores", gmqb.PushOpts{
		Each:  []interface{}{89, 92},
		Sort:  bson.D{{Key: "score", Value: -1}}, // Keep highest scores
		Slice: &slice,                            // Keep only the top 5
	})

	fmt.Println("AddToSet Update:")
	fmt.Println(u1.JSON())

	fmt.Println("PushWithOpts Update:")
	fmt.Println(u2.JSON())
}
