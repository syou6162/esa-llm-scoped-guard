package guard

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/syou6162/esa-llm-scoped-guard/internal/esa"
)

// mockEsaClientForExecute is a mock implementation of EsaClientInterface for execute tests
type mockEsaClientForExecute struct {
	createPostFunc func(*esa.PostInput) (*esa.Post, error)
	updatePostFunc func(int, *esa.PostInput) (*esa.Post, error)
	getPostFunc    func(int) (*esa.Post, error)
}

func (m *mockEsaClientForExecute) CreatePost(input *esa.PostInput) (*esa.Post, error) {
	if m.createPostFunc != nil {
		return m.createPostFunc(input)
	}
	return &esa.Post{Number: 123}, nil
}

func (m *mockEsaClientForExecute) UpdatePost(number int, input *esa.PostInput) (*esa.Post, error) {
	if m.updatePostFunc != nil {
		return m.updatePostFunc(number, input)
	}
	return &esa.Post{Number: number}, nil
}

func (m *mockEsaClientForExecute) GetPost(number int) (*esa.Post, error) {
	if m.getPostFunc != nil {
		return m.getPostFunc(number)
	}
	return &esa.Post{
		Number:   number,
		Category: "Claude Code/開発日誌/2026/01/28",
		Tags:     []string{},
	}, nil
}

// TestExecutePost_CreateNewUpdatesJSON tests that JSON file is automatically updated after successful post with create_new
func TestExecutePost_CreateNewUpdatesJSON(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")

	// 新規作成用のJSONファイルを作成（create_new: true）
	inputJSON := `{
		"create_new": true,
		"name": "Test Post",
		"category": "Claude Code/開発日誌/2026/01/28",
		"body": {
			"background": "Test background",
			"tasks": [
				{
					"id": "task-1",
					"title": "Task 1: Test task",
					"status": "not_started",
					"summary": ["Task summary"],
					"description": "Task description"
				}
			]
		}
	}`

	if err := os.WriteFile(tmpFile, []byte(inputJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// モッククライアント（新規作成を返す）
	mockClient := &mockEsaClientForExecute{
		createPostFunc: func(input *esa.PostInput) (*esa.Post, error) {
			return &esa.Post{Number: 999}, nil
		},
	}

	allowedCategories := []string{"Claude Code/開発日誌"}

	// ExecutePost実行（内部でJSON更新が行われるはず）
	err := executePostWithClient(tmpFile, allowedCategories, mockClient)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// JSONファイルが自動更新されているか確認
	updatedInput, err := ReadPostInputFromFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read updated JSON: %v", err)
	}

	// create_newがfalseになっている
	if updatedInput.CreateNew {
		t.Error("expected CreateNew to be false after successful post")
	}

	// post_numberが設定されている
	if updatedInput.PostNumber == nil {
		t.Error("expected PostNumber to be set after successful post")
	} else if *updatedInput.PostNumber != 999 {
		t.Errorf("expected PostNumber to be 999, got %d", *updatedInput.PostNumber)
	}
}

// TestExecutePost_UpdateDoesNotChangeJSON tests that JSON file is not modified when updating existing post
func TestExecutePost_UpdateDoesNotChangeJSON(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")

	// 更新用のJSONファイルを作成（post_number: 123）
	inputJSON := `{
		"post_number": 123,
		"name": "Test Post",
		"category": "Claude Code/開発日誌/2026/01/28",
		"body": {
			"background": "Test background",
			"tasks": [
				{
					"id": "task-1",
					"title": "Task 1: Test task",
					"status": "not_started",
					"summary": ["Task summary"],
					"description": "Task description"
				}
			]
		}
	}`

	if err := os.WriteFile(tmpFile, []byte(inputJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// モッククライアント（更新を返す）
	mockClient := &mockEsaClientForExecute{
		getPostFunc: func(number int) (*esa.Post, error) {
			return &esa.Post{
				Number:   123,
				Category: "Claude Code/開発日誌/2026/01/28",
				Tags:     []string{},
			}, nil
		},
		updatePostFunc: func(number int, input *esa.PostInput) (*esa.Post, error) {
			return &esa.Post{Number: 123}, nil
		},
	}

	allowedCategories := []string{"Claude Code/開発日誌"}

	// ExecutePost実行（更新なのでJSONは変更されないはず）
	err := executePostWithClient(tmpFile, allowedCategories, mockClient)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// JSONファイルが変更されていないことを確認
	updatedInput, err := ReadPostInputFromFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read JSON: %v", err)
	}

	// create_newはfalseのまま
	if updatedInput.CreateNew {
		t.Error("expected CreateNew to remain false")
	}

	// post_numberは123のまま
	if updatedInput.PostNumber == nil {
		t.Error("expected PostNumber to remain set")
	} else if *updatedInput.PostNumber != 123 {
		t.Errorf("expected PostNumber to remain 123, got %d", *updatedInput.PostNumber)
	}
}

// TestExecutePost_CreateFailureDoesNotChangeJSON tests that JSON file is not modified when post creation fails
func TestExecutePost_CreateFailureDoesNotChangeJSON(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")

	// 新規作成用のJSONファイルを作成（create_new: true）
	inputJSON := `{
		"create_new": true,
		"name": "Test Post",
		"category": "Claude Code/開発日誌/2026/01/28",
		"body": {
			"background": "Test background",
			"tasks": [
				{
					"id": "task-1",
					"title": "Task 1: Test task",
					"status": "not_started",
					"summary": ["Task summary"],
					"description": "Task description"
				}
			]
		}
	}`

	if err := os.WriteFile(tmpFile, []byte(inputJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// モッククライアント（新規作成が失敗する）
	mockClient := &mockEsaClientForExecute{
		createPostFunc: func(input *esa.PostInput) (*esa.Post, error) {
			return nil, fmt.Errorf("API error")
		},
	}

	allowedCategories := []string{"Claude Code/開発日誌"}

	// ExecutePost実行（失敗するのでJSONは変更されないはず）
	err := executePostWithClient(tmpFile, allowedCategories, mockClient)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// JSONファイルが変更されていないことを確認
	updatedInput, err := ReadPostInputFromFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read JSON: %v", err)
	}

	// create_newはtrueのまま
	if !updatedInput.CreateNew {
		t.Error("expected CreateNew to remain true after failed post")
	}

	// post_numberは設定されていない
	if updatedInput.PostNumber != nil {
		t.Errorf("expected PostNumber to remain nil, got %d", *updatedInput.PostNumber)
	}
}

// TestExecutePost_EmbeddsJSONInMarkdown tests that GenerateMarkdownWithJSON correctly embeds JSON
func TestExecutePost_EmbeddsJSONInMarkdown(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")

	postNumber := 123
	// 更新用のJSONファイルを作成（post_number: 123）
	inputJSON := `{
		"post_number": 123,
		"name": "Test Post",
		"category": "Claude Code/開発日誌/2026/01/28",
		"body": {
			"background": "Test background",
			"related_links": ["https://example.com"],
			"instructions": ["Instruction 1"],
			"tasks": [
				{
					"id": "task-1",
					"title": "Task 1: Test task",
					"status": "not_started",
					"summary": ["Task summary"],
					"description": "Task description"
				}
			]
		}
	}`

	if err := os.WriteFile(tmpFile, []byte(inputJSON), 0644); err != nil {
		t.Fatal(err)
	}

	var capturedBodyMD string

	// モッククライアント（UpdatePostで送信されるbody_mdをキャプチャ）
	mockClient := &mockEsaClientForExecute{
		getPostFunc: func(number int) (*esa.Post, error) {
			return &esa.Post{
				Number:   123,
				Category: "Claude Code/開発日誌/2026/01/28",
				Tags:     []string{},
			}, nil
		},
		updatePostFunc: func(number int, input *esa.PostInput) (*esa.Post, error) {
			capturedBodyMD = input.BodyMD
			return &esa.Post{Number: 123}, nil
		},
	}

	allowedCategories := []string{"Claude Code/開発日誌"}

	// ExecutePost実行
	err := executePostWithClient(tmpFile, allowedCategories, mockClient)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 送信されたbody_mdを検証
	if capturedBodyMD == "" {
		t.Fatal("body_md was not captured")
	}

	// 1. 先頭が<!-- esa-guard-jsonで始まっていること
	if len(capturedBodyMD) < len(Sentinel) {
		t.Fatalf("body_md too short: %d bytes", len(capturedBodyMD))
	}
	if capturedBodyMD[:len(Sentinel)] != Sentinel {
		t.Errorf("body_md does not start with sentinel, got: %s", capturedBodyMD[:50])
	}

	// 2. 埋め込まれたJSONを抽出してパース可能なこと
	extracted, err := ExtractEmbeddedJSON(capturedBodyMD)
	if err != nil {
		t.Fatalf("failed to extract embedded JSON: %v", err)
	}

	// 3. 抽出したJSONが元のPostInputと一致すること
	if extracted.PostNumber == nil || *extracted.PostNumber != postNumber {
		t.Errorf("expected post_number %d, got %v", postNumber, extracted.PostNumber)
	}
	if extracted.Name != "Test Post" {
		t.Errorf("expected name 'Test Post', got '%s'", extracted.Name)
	}
	if extracted.Category != "Claude Code/開発日誌/2026/01/28" {
		t.Errorf("expected category 'Claude Code/開発日誌/2026/01/28', got '%s'", extracted.Category)
	}
	if extracted.Body.Background != "Test background" {
		t.Errorf("expected background 'Test background', got '%s'", extracted.Body.Background)
	}
	if len(extracted.Body.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(extracted.Body.Tasks))
	}
	if extracted.Body.Tasks[0].Title != "Task 1: Test task" {
		t.Errorf("expected task title 'Task 1: Test task', got '%s'", extracted.Body.Tasks[0].Title)
	}

	// 4. Markdownセクションが含まれていること
	// ExtractEmbeddedJSONが成功している時点でフォーマットは正しいので、
	// 十分な長さがあることだけ確認
	if len(capturedBodyMD) < 100 {
		t.Errorf("body_md seems too short, expected markdown sections")
	}
}
