package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/syou6162/esa-llm-scoped-guard/internal/esa"
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

	if err := run(jsonPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// getRepositoryName はGitリポジトリ名を取得します
func getRepositoryName() (string, error) {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get repository name: %w", err)
	}

	url := strings.TrimSpace(string(output))
	if url == "" {
		return "", fmt.Errorf("repository URL is empty")
	}

	// URLからリポジトリ名を抽出
	// 例: https://github.com/user/repo.git → repo
	// 例: git@github.com:user/repo.git → repo
	parts := strings.Split(url, "/")
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid repository URL format")
	}

	repoName := parts[len(parts)-1]
	repoName = strings.TrimSuffix(repoName, ".git")

	if repoName == "" {
		return "", fmt.Errorf("repository name is empty")
	}

	return repoName, nil
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

	// 3. JSONファイルの読み込みとバリデーション
	input, err := guard.ReadPostInputFromFile(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to read JSON file: %w", err)
	}

	// フィールドのトリミング（スキーマバリデーション前）
	guard.TrimPostInput(input)

	// JSONスキーマバリデーション
	if err := guard.ValidatePostInputSchema(input); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	// 詳細なバリデーション実行
	if err := guard.ValidatePostInput(input); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// 4. カテゴリ権限チェック
	allowed, err := guard.IsAllowedCategory(input.Category, config.AllowedCategories)
	if err != nil {
		return fmt.Errorf("category validation failed: %w", err)
	}
	if !allowed {
		return fmt.Errorf("category %s is not allowed", input.Category)
	}

	// 5. リポジトリ名を取得してタグに設定
	var tags []string
	repoName, err := getRepositoryName()
	if err == nil && repoName != "" {
		tags = []string{repoName}
	}
	// gitリポジトリじゃない場合はタグなし

	// 6. esa.io APIクライアントで投稿
	client := esa.NewEsaClient(config.Esa.TeamName, accessToken)

	esaInput := &esa.PostInput{
		Name:     input.Name,
		Category: input.Category,
		Tags:     tags,
		BodyMD:   input.BodyMD,
		WIP:      false, // 常にShip It!
	}

	var post *esa.Post
	if input.PostNumber != nil {
		// 更新の場合：既存記事のカテゴリを検証
		existingPost, err := client.GetPost(*input.PostNumber)
		if err != nil {
			return fmt.Errorf("failed to get existing post: %w", err)
		}

		// 更新リクエストの妥当性を検証
		if err := guard.ValidateUpdateRequest(existingPost.Category, input.Category, config.AllowedCategories); err != nil {
			return err
		}

		post, err = client.UpdatePost(*input.PostNumber, esaInput)
		if err != nil {
			return fmt.Errorf("failed to update post: %w", err)
		}
		fmt.Printf("Updated post: %s (Number: %d)\n", post.URL, post.Number)
	} else {
		// 新規作成
		post, err = client.CreatePost(esaInput)
		if err != nil {
			return fmt.Errorf("failed to create post: %w", err)
		}
		fmt.Printf("Created post: %s (Number: %d)\n", post.URL, post.Number)
	}

	return nil
}
