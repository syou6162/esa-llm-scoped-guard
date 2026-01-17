package main

import (
	"flag"
	"fmt"
	"os"
)

const usage = `esa-llm-scoped-guard - Write to esa.io with category restrictions

Usage:
  esa-llm-scoped-guard -json <path>

Flags:
  -json string
        Path to JSON file containing post data
  -help
        Show this help message and JSON schema

JSON Schema:
  {
    "post_number": 123,           // Optional: omit for new post creation
    "name": "Post Title",          // Required: max 255 bytes
    "category": "LLM/Tasks/2025-01", // Required: must match allowed categories
    "tags": ["tag1", "tag2"],     // Optional: max 10 tags, each max 50 bytes
    "body_md": "## Content\n..."  // Required: max 1MB
  }

Environment Variables:
  ESA_ACCESS_TOKEN    esa.io API access token

Configuration:
  ~/.config/esa-llm-scoped-guard/config.yaml

Example:
  esa-llm-scoped-guard -json ./tasks/123.json
`

func main() {
	var jsonPath string
	var showHelp bool

	flag.StringVar(&jsonPath, "json", "", "Path to JSON file containing post data")
	flag.BoolVar(&showHelp, "help", false, "Show help message")
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}
	flag.Parse()

	if showHelp || jsonPath == "" {
		flag.Usage()
		if showHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}

	fmt.Printf("Processing JSON file: %s\n", jsonPath)
	// TODO: Implement actual logic
}
