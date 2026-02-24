package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/squall-chua/gmqb/generator"
)

func main() {
	queryFlag := flag.String("query", "", "MongoDB query JSON string to translate")
	flag.Parse()

	var jsonStr string

	if *queryFlag != "" {
		jsonStr = *queryFlag
	} else {
		// Try to read from stdin
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			bytes, err := io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
				os.Exit(1)
			}
			jsonStr = string(bytes)
		} else {
			fmt.Fprintln(os.Stderr, "Usage: gmqb-gen -query '{\"age\": {\"$gte\": 18}}' OR echo '{\"age\": {\"$gte\": 18}}' | gmqb-gen")
			os.Exit(1)
		}
	}

	jsonStr = strings.TrimSpace(jsonStr)
	if jsonStr == "" {
		fmt.Fprintln(os.Stderr, "No query provided")
		os.Exit(1)
	}

	code, err := generator.Generate(jsonStr)
	if err != nil {
		// Fallback: users often paste improperly escaped JSON on the CLI (e.g. `{\"foo\": \"bar\"}`)
		if strings.Contains(jsonStr, `\"`) {
			// Naive unescape
			unescaped := strings.ReplaceAll(jsonStr, `\"`, `"`)
			// Also replace double escapes if they occur from multiple encodings
			unescaped = strings.ReplaceAll(unescaped, `\\`, `\`)

			if codeFallback, errFallback := generator.Generate(unescaped); errFallback == nil {
				code = codeFallback
				err = nil
			}
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Generation error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(code)
}
