package main

import (
	"fmt"

	"github.com/squall-chua/gmqb"
)

func main() {
	filter := gmqb.Eq("status", "active")

	// Serializes identical to what the MongoDB shell expects
	jsonStr := filter.JSON()
	compactJson := filter.CompactJSON()

	fmt.Println("Pretty JSON:")
	fmt.Println(jsonStr)

	fmt.Println("\nCompact JSON:")
	fmt.Println(compactJson)
}
