package main

import (
	"fmt"

	"github.com/squall-chua/gmqb"
)

func main() {
	update := gmqb.NewUpdate().
		Set("status", "active").
		Inc("loginCount", 1).
		Unset("temporaryToken").
		CurrentDateAsTimestamp("lastModified")

	fmt.Println("Field Update:")
	fmt.Println(update.JSON())
}
