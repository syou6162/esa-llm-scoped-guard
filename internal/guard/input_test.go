package guard

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadPostInputFromFile(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		wantErr     bool
		wantErrCode ValidationErrorCode
		validate    func(*testing.T, *PostInput)
	}{
		{
			name: "有効なYAML",
			yamlContent: `name: Test Post
category: LLM/Tasks
body:
  background: Task background
  tasks: []`,
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
			yamlContent: `name: 日本語テスト
category: Claude Code/開発日誌
body:
  background: タスクの背景
  tasks: []`,
			wantErr: false,
			validate: func(t *testing.T, input *PostInput) {
				if input.Category != "Claude Code/開発日誌" {
					t.Errorf("Category = %v, want Claude Code/開発日誌", input.Category)
				}
			},
		},
		{
			name:        "不正なYAML",
			yamlContent: `name: [unclosed`,
			wantErr:     true,
			wantErrCode: ErrCodeYAMLInvalid,
		},
		{
			name: "未知のフィールド",
			yamlContent: `name: Test
category: LLM/Tasks
body:
  background: Content
  tasks: []
unknown_field: value`,
			wantErr:     true,
			wantErrCode: ErrCodeYAMLInvalid,
		},
		{
			name: "複数のYAMLドキュメント",
			yamlContent: `name: Test
category: LLM/Tasks
body:
  background: Content
  tasks: []
---
name: Test2
category: LLM/Tasks
body:
  background: Content2
  tasks: []`,
			wantErr:     true,
			wantErrCode: ErrCodeYAMLInvalid,
		},
		{
			name: "trailing data",
			yamlContent: `name: Test
category: LLM/Tasks
body:
  background: Content
  tasks: []
invalid`,
			wantErr:     true,
			wantErrCode: ErrCodeYAMLInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			yamlPath := filepath.Join(tmpDir, "test.yaml")

			if err := os.WriteFile(yamlPath, []byte(tt.yamlContent), 0600); err != nil {
				t.Fatalf("Failed to write test YAML: %v", err)
			}

			input, err := ReadPostInputFromFile(yamlPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadPostInputFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				var ve *ValidationError
				if !errors.As(err, &ve) {
					t.Errorf("Expected ValidationError, got %T", err)
					return
				}
				if ve.Code() != tt.wantErrCode {
					t.Errorf("ValidationError.Code() = %v, want %v", ve.Code(), tt.wantErrCode)
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
		return
	}
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Errorf("Expected ValidationError, got %T", err)
		return
	}
	if ve.Code() != ErrCodeFileSizeExceeded {
		t.Errorf("ValidationError.Code() = %v, want %v", ve.Code(), ErrCodeFileSizeExceeded)
	}
}

func TestReadPostInputFromFile_NonRegularFile(t *testing.T) {
	// ディレクトリを指定
	tmpDir := t.TempDir()

	_, err := ReadPostInputFromFile(tmpDir)
	if err == nil {
		t.Error("Expected error for non-regular file, got nil")
		return
	}
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Errorf("Expected ValidationError, got %T", err)
		return
	}
	if ve.Code() != ErrCodeNotRegularFile {
		t.Errorf("ValidationError.Code() = %v, want %v", ve.Code(), ErrCodeNotRegularFile)
	}
}
