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

	var output string
	var execErr error
	output = captureStdout(func() {
		execErr = executeDiffWithClient(tmpFile, mockClient)
	})

	if execErr != nil {
		t.Errorf("expected no error, got %v", execErr)
	}

	if output == "" {
		t.Error("expected diff output, got empty string")
	}

	// DiffPrettyTextは変更箇所を示す出力を生成する
	// 差分がある場合は何らかの出力があることを確認
	if len(output) < 10 {
		t.Errorf("expected substantial diff output, got: %q", output)
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

	mockClient := &mockEsaClient{}
	err := executeDiffWithClient(tmpFile, mockClient)
	if err == nil {
		t.Error("expected error for create_new")
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

	mockClient := &mockEsaClient{}
	err := executeDiffWithClient(tmpFile, mockClient)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
