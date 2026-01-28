package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/syou6162/esa-llm-scoped-guard/internal/esa"
)

type mockEsaClient struct {
	getPostFunc func(number int) (*esa.Post, error)
}

func (m *mockEsaClient) CreatePost(input *esa.PostInput) (*esa.Post, error) {
	return nil, nil
}

func (m *mockEsaClient) UpdatePost(number int, input *esa.PostInput) (*esa.Post, error) {
	return nil, nil
}

func (m *mockEsaClient) GetPost(number int) (*esa.Post, error) {
	return m.getPostFunc(number)
}

func TestExecuteDiff_WithPostNumber(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "update.json")

	updateJSON := `{
		"post_number": 123,
		"name": "Test Post",
		"category": "LLM/Tasks/2026/01/28",
		"body": {
			"background": "New background",
			"tasks": [
				{
					"id": "task-1",
					"title": "Task 1: New task",
					"status": "in_progress",
					"summary": ["New summary"],
					"description": "New description"
				}
			]
		}
	}`

	if err := os.WriteFile(tmpFile, []byte(updateJSON), 0600); err != nil {
		t.Fatal(err)
	}

	mockClient := &mockEsaClient{
		getPostFunc: func(number int) (*esa.Post, error) {
			return &esa.Post{
				Number:   123,
				Name:     "Old Post",
				Category: "LLM/Tasks/2026/01/28",
				BodyMD:   "## サマリー\n- [ ] Old task\n\n## 背景\nOld background",
			}, nil
		},
	}

	allowedCategories := []string{"LLM/Tasks"}

	var output string
	var execErr error
	output = captureStdout(func() {
		execErr = executeDiffWithClient(tmpFile, allowedCategories, mockClient)
	})

	if execErr != nil {
		t.Errorf("expected no error, got %v", execErr)
	}

	if output == "" {
		t.Error("expected diff output, got empty string")
	}

	// unified diff形式（@@ を含む）を検証
	if !strings.Contains(output, "@@") {
		t.Errorf("expected unified diff format with @@, got: %q", output)
	}
}

func TestExecuteDiff_CreateNewError(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "new.json")

	newJSON := `{
		"create_new": true,
		"name": "Test Post",
		"category": "LLM/Tasks/2026/01/28",
		"body": {
			"background": "Test background",
			"tasks": [
				{
					"id": "task-1",
					"title": "Task 1: Test",
					"status": "not_started",
					"summary": ["Test summary"],
					"description": "Test description"
				}
			]
		}
	}`

	if err := os.WriteFile(tmpFile, []byte(newJSON), 0600); err != nil {
		t.Fatal(err)
	}

	allowedCategories := []string{"LLM/Tasks"}
	mockClient := &mockEsaClient{}
	err := executeDiffWithClient(tmpFile, allowedCategories, mockClient)
	if err == nil {
		t.Fatal("expected error for create_new")
	}
	if !strings.Contains(err.Error(), "post_number") {
		t.Errorf("expected error message to mention post_number, got: %v", err)
	}
}

func TestExecuteDiff_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.json")

	invalidJSON := `{
		"post_number": 123,
		"name": "Test Post",
		"category": "Invalid/Category"
	}`

	if err := os.WriteFile(tmpFile, []byte(invalidJSON), 0600); err != nil {
		t.Fatal(err)
	}

	allowedCategories := []string{"LLM/Tasks"}
	mockClient := &mockEsaClient{}
	err := executeDiffWithClient(tmpFile, allowedCategories, mockClient)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestExecuteDiff_CategoryNotAllowed(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "update.json")

	updateJSON := `{
		"post_number": 123,
		"name": "Test Post",
		"category": "LLM/Tasks/2026/01/28",
		"body": {
			"background": "New background",
			"tasks": [
				{
					"id": "task-1",
					"title": "Task 1: New task",
					"status": "in_progress",
					"summary": ["New summary"],
					"description": "New description"
				}
			]
		}
	}`

	if err := os.WriteFile(tmpFile, []byte(updateJSON), 0600); err != nil {
		t.Fatal(err)
	}

	// 許可されていないカテゴリの既存記事を返すモック
	mockClient := &mockEsaClient{
		getPostFunc: func(number int) (*esa.Post, error) {
			return &esa.Post{
				Number:   123,
				Name:     "Old Post",
				Category: "Restricted/Category/2026/01/28", // 許可されていないカテゴリ
				BodyMD:   "## サマリー\n- [ ] Old task\n\n## 背景\nOld background",
			}, nil
		},
	}

	allowedCategories := []string{"LLM/Tasks"}
	err := executeDiffWithClient(tmpFile, allowedCategories, mockClient)
	if err == nil {
		t.Fatal("expected error for category not allowed")
	}
	if !strings.Contains(err.Error(), "category") {
		t.Errorf("expected error message to mention category, got: %v", err)
	}
}

func TestExecuteDiff_CategoryChangeAttempt(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "update.json")

	updateJSON := `{
		"post_number": 123,
		"name": "Test Post",
		"category": "LLM/Tasks/2026/01/29",
		"body": {
			"background": "New background",
			"tasks": [
				{
					"id": "task-1",
					"title": "Task 1: New task",
					"status": "in_progress",
					"summary": ["New summary"],
					"description": "New description"
				}
			]
		}
	}`

	if err := os.WriteFile(tmpFile, []byte(updateJSON), 0600); err != nil {
		t.Fatal(err)
	}

	// カテゴリが異なる既存記事を返すモック
	mockClient := &mockEsaClient{
		getPostFunc: func(number int) (*esa.Post, error) {
			return &esa.Post{
				Number:   123,
				Name:     "Old Post",
				Category: "LLM/Tasks/2026/01/28", // 日付が異なる
				BodyMD:   "## サマリー\n- [ ] Old task\n\n## 背景\nOld background",
			}, nil
		},
	}

	allowedCategories := []string{"LLM/Tasks"}
	err := executeDiffWithClient(tmpFile, allowedCategories, mockClient)
	if err == nil {
		t.Fatal("expected error for category change attempt")
	}
	if !strings.Contains(err.Error(), "category change") {
		t.Errorf("expected error message to mention category change, got: %v", err)
	}
}
