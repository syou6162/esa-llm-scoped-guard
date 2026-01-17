package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadPostInputFromFile(t *testing.T) {
	tests := []struct {
		name        string
		jsonContent string
		wantErr     bool
		errMsg      string
		validate    func(*testing.T, *PostInput)
	}{
		{
			name: "有効なJSON",
			jsonContent: `{
				"name": "Test Post",
				"category": "LLM/Tasks",
				"body": {
					"background": "Task background"
				}
			}`,
			wantErr: false,
			validate: func(t *testing.T, input *PostInput) {
				if input.Name != "Test Post" {
					t.Errorf("Name = %v, want Test Post", input.Name)
				}
				if input.Category != "LLM/Tasks" {
					t.Errorf("Category = %v, want LLM/Tasks", input.Category)
				}
				if input.Body.Background != "Task background" {
					t.Errorf("Body.Background = %v, want Task background", input.Body.Background)
				}
			},
		},
		{
			name: "日本語カテゴリ",
			jsonContent: `{
				"name": "日本語テスト",
				"category": "Claude Code/開発日誌",
				"body": {
					"background": "タスクの背景"
				}
			}`,
			wantErr: false,
			validate: func(t *testing.T, input *PostInput) {
				if input.Category != "Claude Code/開発日誌" {
					t.Errorf("Category = %v, want Claude Code/開発日誌", input.Category)
				}
			},
		},
		{
			name:        "不正なJSON",
			jsonContent: `{"name": "Test"`,
			wantErr:     true,
			errMsg:      "failed to parse JSON",
		},
		{
			name: "未知のフィールド",
			jsonContent: `{
				"name": "Test",
				"category": "LLM/Tasks",
				"body": {
					"background": "Content"
				},
				"unknown_field": "value"
			}`,
			wantErr: true,
			errMsg:  "failed to parse JSON",
		},
		{
			name: "複数のJSONオブジェクト",
			jsonContent: `{
				"name": "Test",
				"category": "LLM/Tasks",
				"body": {
					"background": "Content"
				}
			}
			{
				"name": "Test2",
				"category": "LLM/Tasks",
				"body": {
					"background": "Content2"
				}
			}`,
			wantErr: true,
			errMsg:  "JSON file contains multiple values",
		},
		{
			name: "trailing data",
			jsonContent: `{
				"name": "Test",
				"category": "LLM/Tasks",
				"body": {
					"background": "Content"
				}
			} extra data`,
			wantErr: true,
			errMsg:  "JSON file contains multiple values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			jsonPath := filepath.Join(tmpDir, "test.json")

			if err := os.WriteFile(jsonPath, []byte(tt.jsonContent), 0600); err != nil {
				t.Fatalf("Failed to write test JSON: %v", err)
			}

			input, err := ReadPostInputFromFile(jsonPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadPostInputFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ReadPostInputFromFile() error = %v, want error containing %q", err, tt.errMsg)
				}
				return
			}

			if tt.validate != nil {
				tt.validate(t, input)
			}
		})
	}
}

func TestReadPostInputFromFile_FileSize(t *testing.T) {
	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "large.json")

	// 10MB超過のファイル
	largeContent := `{"name": "Test", "category": "LLM/Tasks", "body": {"background": "` + strings.Repeat("a", 10*1024*1024) + `"}}`
	if err := os.WriteFile(jsonPath, []byte(largeContent), 0600); err != nil {
		t.Fatalf("Failed to write test JSON: %v", err)
	}

	_, err := ReadPostInputFromFile(jsonPath)
	if err == nil {
		t.Error("Expected error for file size exceeding 10MB, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "file size exceeds 10MB") {
		t.Errorf("Expected 'file size exceeds 10MB' error, got %v", err)
	}
}

func TestReadPostInputFromFile_NonRegularFile(t *testing.T) {
	// ディレクトリを指定
	tmpDir := t.TempDir()

	_, err := ReadPostInputFromFile(tmpDir)
	if err == nil {
		t.Error("Expected error for non-regular file, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "not a regular file") {
		t.Errorf("Expected 'not a regular file' error, got %v", err)
	}
}
