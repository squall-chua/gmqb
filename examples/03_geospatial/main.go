package main

import (
	"fmt"
	"github.com/squall-chua/gmqb"
)

func main() {
	// Find locations near a GeoJSON point
	point := gmqb.Point(-73.9667, 40.78)

	filter := gmqb.Near("location", point, 5000, 100) // max 5000m, min 100m

	fmt.Println("Geospatial Near Filter:")
	fmt.Println(filter.JSON())
}
