package guard

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/syou6162/esa-llm-scoped-guard/internal/esa"
)

// ExecutePost はesa.io記事の作成/更新を実行します
func ExecutePost(jsonPath string, teamName string, allowedCategories []string, accessToken string) error {
	// 1. JSONファイルの読み込みとバリデーション
	input, err := ReadPostInputFromFile(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to read JSON file: %w", err)
	}

	// フィールドのトリミング（スキーマバリデーション前）
	TrimPostInput(input)

	// JSONスキーマバリデーション
	if err := ValidatePostInputSchema(input); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	// 詳細なバリデーション実行
	if err := ValidatePostInput(input); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// 2. カテゴリ権限チェック
	allowed, err := IsAllowedCategory(input.Category, allowedCategories)
	if err != nil {
		return fmt.Errorf("category validation failed: %w", err)
	}
	if !allowed {
		return fmt.Errorf("category %s is not allowed", input.Category)
	}

	// 3. リポジトリ名を取得
	repoName, err := getRepositoryName()
	if err != nil {
		repoName = "" // gitリポジトリじゃない場合は空
	}

	// 4. esa.io APIクライアントで投稿
	client := esa.NewEsaClient(teamName, accessToken)

	if input.CreateNew {
		return createPost(client, input, repoName)
	}
	return updatePost(client, input, allowedCategories, repoName)
}

// updatePost は既存記事を更新します
func updatePost(client esa.EsaClientInterface, input *PostInput, allowedCategories []string, repoName string) error {
	// 既存記事のカテゴリを検証
	existingPost, err := client.GetPost(*input.PostNumber)
	if err != nil {
		return fmt.Errorf("failed to get existing post: %w", err)
	}

	// 更新リクエストの妥当性を検証
	if err := ValidateUpdateRequest(existingPost.Category, input.Category, allowedCategories); err != nil {
		return err
	}

	// 既存のタグを保持し、現在のリポジトリ名がなければ追加
	tags := MergeTags(existingPost.Tags, repoName)

	// BodyからマークダウンGenerate
	bodyMD := GenerateMarkdown(&input.Body)

	esaInput := &esa.PostInput{
		Name:     input.Name,
		Category: input.Category,
		Tags:     tags,
		BodyMD:   bodyMD,
		WIP:      false, // 常にShip It!
	}

	post, err := client.UpdatePost(*input.PostNumber, esaInput)
	if err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}
	fmt.Printf("Updated post: %s (Number: %d)\n", post.URL, post.Number)
	return nil
}

// createPost は新規記事を作成します
func createPost(client esa.EsaClientInterface, input *PostInput, repoName string) error {
	// 現在のリポジトリ名のみをタグに設定
	var tags []string
	if repoName != "" {
		tags = []string{repoName}
	}

	// BodyからマークダウンGenerate
	bodyMD := GenerateMarkdown(&input.Body)

	esaInput := &esa.PostInput{
		Name:     input.Name,
		Category: input.Category,
		Tags:     tags,
		BodyMD:   bodyMD,
		WIP:      false, // 常にShip It!
	}

	post, err := client.CreatePost(esaInput)
	if err != nil {
		return fmt.Errorf("failed to create post: %w", err)
	}
	fmt.Printf("Created post: %s (Number: %d)\n", post.URL, post.Number)
	return nil
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
