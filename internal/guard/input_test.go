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
		jsonContent string
		wantErr     bool
		wantErrCode ValidationErrorCode
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
			wantErrCode: ErrCodeJSONInvalid,
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
			wantErr:     true,
			wantErrCode: ErrCodeJSONInvalid,
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
			wantErr:     true,
			wantErrCode: ErrCodeJSONInvalid,
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
			wantErr:     true,
			wantErrCode: ErrCodeJSONInvalid,
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
				var ve *ValidationError
				if errors.As(err, &ve) {
					if ve.Code() != tt.wantErrCode {
						t.Errorf("ValidationError.Code() = %v, want %v", ve.Code(), tt.wantErrCode)
					}
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
