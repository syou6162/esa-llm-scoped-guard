package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/syou6162/esa-llm-scoped-guard/internal/guard"
)

// TestIntegrationEndToEnd は全体フローの統合テストです
func TestIntegrationEndToEnd(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir := t.TempDir()

	// テスト用のJSONファイルを作成
	jsonPath := filepath.Join(tmpDir, "test.json")
	jsonContent := `{
		"name": "Test Post",
		"category": "LLM/Tasks",
		"body_md": "## Test Content"
	}`

	if err := os.WriteFile(jsonPath, []byte(jsonContent), 0600); err != nil {
		t.Fatalf("Failed to write test JSON: %v", err)
	}

	// JSONファイルの読み込み
	input, err := readJSONFile(jsonPath)
	if err != nil {
		t.Fatalf("readJSONFile() error = %v", err)
	}

	// スキーマバリデーション
	if err := ValidatePostInputSchema(input); err != nil {
		t.Fatalf("ValidatePostInputSchema() error = %v", err)
	}

	// 詳細なバリデーション
	if err := ValidatePostInput(input); err != nil {
		t.Fatalf("ValidatePostInput() error = %v", err)
	}

	// カテゴリが正規化されることを確認
	normalizedCategory, err := guard.NormalizeCategory(input.Category)
	if err != nil {
		t.Fatalf("guard.NormalizeCategory() error = %v", err)
	}

	if normalizedCategory != "LLM/Tasks" {
		t.Errorf("Expected normalized category 'LLM/Tasks', got '%s'", normalizedCategory)
	}
}

// TestIntegrationJapaneseCategory は日本語カテゴリの統合テストです
func TestIntegrationJapaneseCategory(t *testing.T) {
	tmpDir := t.TempDir()

	// 日本語カテゴリを含むJSONファイルを作成
	jsonPath := filepath.Join(tmpDir, "japanese.json")
	jsonContent := `{
		"name": "日本語テスト",
		"category": "Claude Code/開発日誌",
		"body_md": "## 日本語カテゴリ\n\nこれは日本語カテゴリのテストです。"
	}`

	if err := os.WriteFile(jsonPath, []byte(jsonContent), 0600); err != nil {
		t.Fatalf("Failed to write test JSON: %v", err)
	}

	// JSONファイルの読み込み
	input, err := readJSONFile(jsonPath)
	if err != nil {
		t.Fatalf("readJSONFile() error = %v", err)
	}

	// スキーマバリデーション
	if err := ValidatePostInputSchema(input); err != nil {
		t.Fatalf("ValidatePostInputSchema() error = %v", err)
	}

	// 詳細なバリデーション
	if err := ValidatePostInput(input); err != nil {
		t.Fatalf("ValidatePostInput() error = %v", err)
	}

	// カテゴリが正規化されることを確認
	normalizedCategory, err := guard.NormalizeCategory(input.Category)
	if err != nil {
		t.Fatalf("guard.NormalizeCategory() error = %v", err)
	}

	if normalizedCategory != "Claude Code/開発日誌" {
		t.Errorf("Expected normalized category 'Claude Code/開発日誌', got '%s'", normalizedCategory)
	}
}

// TestIntegrationInvalidJSON は不正なJSONの統合テストです
func TestIntegrationInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		jsonContent string
		wantErr     string
	}{
		{
			name:        "空のname",
			jsonContent: `{"name": "", "category": "LLM/Tasks", "body_md": "content"}`,
			wantErr:     "schema validation failed",
		},
		{
			name:        "カテゴリなし",
			jsonContent: `{"name": "Test", "body_md": "content"}`,
			wantErr:     "schema validation failed",
		},
		{
			name:        "body_mdなし",
			jsonContent: `{"name": "Test", "category": "LLM/Tasks"}`,
			wantErr:     "schema validation failed",
		},
		{
			name:        "不正なpost_number",
			jsonContent: `{"post_number": 0, "name": "Test", "category": "LLM/Tasks", "body_md": "content"}`,
			wantErr:     "schema validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonPath := filepath.Join(tmpDir, tt.name+".json")
			if err := os.WriteFile(jsonPath, []byte(tt.jsonContent), 0600); err != nil {
				t.Fatalf("Failed to write test JSON: %v", err)
			}

			input, err := readJSONFile(jsonPath)
			if err != nil {
				// readJSONFileでエラーが出る場合もある
				return
			}

			// スキーマバリデーション
			err = ValidatePostInputSchema(input)
			if err == nil {
				t.Errorf("Expected schema validation error, got nil")
			}
		})
	}
}
