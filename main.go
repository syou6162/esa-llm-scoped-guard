package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/syou6162/esa-llm-scoped-guard/internal/guard"
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
    "body_md": "## Content\n..."  // Required: max 1MB
  }

Note: Tags are automatically set to the Git repository name (no tags if not a git repository).

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

	if err := run(jsonPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(jsonPath string) error {
	// 1. 設定ファイルの読み込み
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	configPath := filepath.Join(homeDir, ".config", "esa-llm-scoped-guard", "config.yaml")
	config, err := LoadAndValidateConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// 2. 環境変数からESA_ACCESS_TOKENを取得
	accessToken := os.Getenv("ESA_ACCESS_TOKEN")
	if accessToken == "" {
		return fmt.Errorf("ESA_ACCESS_TOKEN environment variable is not set")
	}

	// 3. esa.io記事の作成/更新を実行
	return guard.ExecutePost(jsonPath, config.Esa.TeamName, config.AllowedCategories, accessToken)
}
