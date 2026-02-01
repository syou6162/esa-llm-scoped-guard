package guard

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/syou6162/esa-llm-scoped-guard/internal/esa"
	"gopkg.in/yaml.v3"
)

// ExecutePost はesa.io記事の作成/更新を実行します
func ExecutePost(yamlPath string, teamName string, allowedCategories []string, accessToken string) error {
	client := esa.NewEsaClient(teamName, accessToken)
	return executePostWithClient(yamlPath, allowedCategories, client)
}

// executePostWithClient はesa.io記事の作成/更新を実行します（テスト可能なバージョン）
func executePostWithClient(yamlPath string, allowedCategories []string, client esa.EsaClientInterface) error {
	// 1. YAMLファイルの読み込みとバリデーション
	input, err := ReadPostInputFromFile(yamlPath)
	if err != nil {
		return fmt.Errorf("failed to read YAML file: %w", err)
	}

	// フィールドのトリミング
	TrimPostInput(input)

	// バリデーション実行
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
	var postNumber int
	if input.CreateNew {
		postNumber, err = createPost(client, input, repoName)
		if err != nil {
			return err
		}

		// 新規作成成功時にYAMLファイルを自動更新
		if err := updateYAMLAfterCreate(yamlPath, postNumber); err != nil {
			// 警告を出すが、投稿自体は成功しているのでエラーにしない
			fmt.Fprintf(os.Stderr, "Warning: failed to update YAML file: %v\n", err)
			fmt.Fprintf(os.Stderr, "You may need to manually update the YAML file to use diff/update commands.\n")
		} else {
			fmt.Printf("YAML file updated: create_new removed, post_number set to %d\n", postNumber)
		}
	} else {
		err = updatePost(client, input, allowedCategories, repoName)
	}
	return err
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

	// BodyからマークダウンGenerate（YAML埋め込み）
	bodyMD, err := GenerateMarkdownWithYAML(input)
	if err != nil {
		return fmt.Errorf("failed to generate markdown with YAML: %w", err)
	}

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
func createPost(client esa.EsaClientInterface, input *PostInput, repoName string) (int, error) {
	// 現在のリポジトリ名のみをタグに設定
	var tags []string
	if repoName != "" {
		tags = []string{repoName}
	}

	// BodyからマークダウンGenerate（YAML埋め込み）
	// 注: この時点ではpost_numberは未設定
	bodyMD, err := GenerateMarkdownWithYAML(input)
	if err != nil {
		return 0, fmt.Errorf("failed to generate markdown with YAML: %w", err)
	}

	esaInput := &esa.PostInput{
		Name:     input.Name,
		Category: input.Category,
		Tags:     tags,
		BodyMD:   bodyMD,
		WIP:      false, // 常にShip It!
	}

	post, err := client.CreatePost(esaInput)
	if err != nil {
		return 0, fmt.Errorf("failed to create post: %w", err)
	}
	fmt.Printf("Created post: %s (Number: %d)\n", post.URL, post.Number)

	// 新規作成後、post_numberを含むYAMLを埋め込むために記事を自動更新
	// これによりfetchコマンドが即座に使用可能になる
	input.PostNumber = &post.Number
	input.CreateNew = false

	bodyMDWithPostNumber, err := GenerateMarkdownWithYAML(input)
	if err != nil {
		// Fail-closed: post_number埋め込み失敗はエラーとして扱う
		return 0, fmt.Errorf("post created at %s (Number: %d) but failed to embed post_number: %w", post.URL, post.Number, err)
	}

	updateInput := &esa.PostInput{
		Name:     input.Name,
		Category: input.Category,
		Tags:     tags,
		BodyMD:   bodyMDWithPostNumber,
		WIP:      false,
	}

	_, err = client.UpdatePost(post.Number, updateInput)
	if err != nil {
		// Fail-closed: post_number埋め込み失敗はエラーとして扱う
		return 0, fmt.Errorf("post created at %s (Number: %d) but failed to update with post_number: %w", post.URL, post.Number, err)
	}

	return post.Number, nil
}

// updateYAMLAfterCreate は新規作成成功後にYAMLファイルを更新します
func updateYAMLAfterCreate(yamlPath string, postNumber int) error {
	// シンボリックリンクを解決（ReadPostInputFromFileと同じ処理）
	absPath, err := filepath.Abs(yamlPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}
	realPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlink: %w", err)
	}

	// 元のファイルのパーミッションとタイプを取得
	fileInfo, err := os.Stat(realPath)
	if err != nil {
		return fmt.Errorf("failed to stat YAML file: %w", err)
	}

	// レギュラーファイルかどうかを確認（セキュリティ強化）
	if !fileInfo.Mode().IsRegular() {
		return fmt.Errorf("YAML file is not a regular file: %s", realPath)
	}

	// YAMLファイルを読み込み
	input, err := ReadPostInputFromFile(yamlPath)
	if err != nil {
		return err
	}

	// create_newをfalseに、post_numberを設定
	input.CreateNew = false
	input.PostNumber = &postNumber

	// YAMLに変換
	data, err := yaml.Marshal(input)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	// サイズチェック（10MB上限）
	if len(data) > MaxInputSize {
		return fmt.Errorf("updated YAML size exceeds %d bytes (got %d bytes)", MaxInputSize, len(data))
	}

	// 一時ファイルに書き込み（原子的更新のため）
	// 実体パスと同一ディレクトリにユニークな一時ファイルを作成
	dir := filepath.Dir(realPath)
	tmpFile, err := os.CreateTemp(dir, ".esa-guard-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath) // 失敗時のクリーンアップ

	// パーミッションを設定して書き込み
	if err := tmpFile.Chmod(fileInfo.Mode().Perm()); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to set temp file permissions: %w", err)
	}
	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// ディスクへの同期（耐障害性向上）
	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// 原子的にリネーム（実体パスに書き込む）
	if err := os.Rename(tmpPath, realPath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

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
