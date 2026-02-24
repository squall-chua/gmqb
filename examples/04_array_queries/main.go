package main

import (
	"fmt"

	"github.com/squall-chua/gmqb"
)

func main() {
	// Match documents where the 'results' array contains at least one element
	// that matches *both* conditions below.
	filter := gmqb.ElemMatch("results", gmqb.And(
		gmqb.Gte("score", 80),
		gmqb.Lt("score", 90),
	))

	// Also filter by array size
	filter2 := gmqb.Size("tags", 3)

	fmt.Println("ElemMatch Filter:")
	fmt.Println(filter.JSON())

	fmt.Println("Size Filter:")
	fmt.Println(filter2.JSON())
}
