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
    "create_new": true,            // Optional: set true for new post (cannot use with post_number)
    "post_number": 123,            // Optional: existing post number for update (cannot use with create_new)
    "name": "Post Title",          // Required: max 255 bytes, no /, （）, or ：
    "category": "LLM/Tasks/2026/01/18", // Required: allowed category + /yyyy/mm/dd
    "body": {                      // Required: structured format
      "background": "Task background (plain text, no '## 背景' header, no # or ## at line start)",
      "related_links": ["https://example.com"], // Optional: related URLs
      "tasks": [                   // Required: task array
        {
          "id": "task-1",          // Required: unique identifier
          "title": "Task title",   // Required (auto-generated: "### {title}")
          "status": "not_started", // Required: not_started/in_progress/in_review/completed (auto-generated: "Status: {status}")
          "description": "Task description", // Required (plain text, status/title auto-generated, no #/##/### at line start)
          "github_urls": ["https://github.com/owner/repo/pull/123"] // Optional: GitHub PR/Issue URLs
        }
      ]
    }
  }

Markdown Output Example:
  Input JSON with github_urls:
    {
      "id": "task-1",
      "title": "Fix bug",
      "status": "in_progress",
      "description": "Fix the authentication bug",
      "github_urls": ["https://github.com/owner/repo/pull/123"]
    }

  Output:
    ### Fix bug
    - Status: ` + "`in_progress`" + `
    - Pull Request: https://github.com/owner/repo/pull/123

    Fix the authentication bug

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
