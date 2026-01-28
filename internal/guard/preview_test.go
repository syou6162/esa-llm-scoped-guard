package guard

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestExecutePreview_ValidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "valid.json")

	validJSON := `{
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

	if err := os.WriteFile(tmpFile, []byte(validJSON), 0600); err != nil {
		t.Fatal(err)
	}

	var output string
	var execErr error
	output = captureStdout(func() {
		execErr = ExecutePreview(tmpFile)
	})

	if execErr != nil {
		t.Errorf("expected no error for valid JSON, got %v", execErr)
	}

	if !strings.Contains(output, "## サマリー") {
		t.Error("expected markdown output with '## サマリー'")
	}

	if !strings.Contains(output, "## 背景") {
		t.Error("expected markdown output with '## 背景'")
	}

	if !strings.Contains(output, "## タスク") {
		t.Error("expected markdown output with '## タスク'")
	}
}

func TestExecutePreview_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.json")

	invalidJSON := `{
		"name": "Test Post",
		"category": "Invalid/Category"
	}`

	if err := os.WriteFile(tmpFile, []byte(invalidJSON), 0600); err != nil {
		t.Fatal(err)
	}

	err := ExecutePreview(tmpFile)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestExecutePreview_FileNotFound(t *testing.T) {
	err := ExecutePreview("/nonexistent/path.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
