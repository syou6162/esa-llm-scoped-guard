package guard

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExecuteValidate_ValidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "valid.yaml")

	validYAML := `create_new: true
name: Test Post
category: LLM/Tasks/2026/01/28
body:
  background: Test background
  tasks:
    - id: task-1
      title: "Task 1: Test"
      status: not_started
      summary:
        - Test summary
      description: Test description
`

	if err := os.WriteFile(tmpFile, []byte(validYAML), 0600); err != nil {
		t.Fatal(err)
	}

	err := ExecuteValidate(tmpFile)
	if err != nil {
		t.Errorf("expected no error for valid YAML, got %v", err)
	}
}

func TestExecuteValidate_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.yaml")

	invalidJSON := `name: Test Post
category: Invalid/Category
`

	if err := os.WriteFile(tmpFile, []byte(invalidJSON), 0600); err != nil {
		t.Fatal(err)
	}

	err := ExecuteValidate(tmpFile)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestExecuteValidate_FileNotFound(t *testing.T) {
	err := ExecuteValidate("/nonexistent/path.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestExecuteValidate_MalformedJSON(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "malformed.yaml")

	malformedJSON := `invalid: [yaml`

	if err := os.WriteFile(tmpFile, []byte(malformedJSON), 0600); err != nil {
		t.Fatal(err)
	}

	err := ExecuteValidate(tmpFile)
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}
