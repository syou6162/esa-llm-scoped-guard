package guard

import (
	"strings"
	"testing"
)

func TestValidatePostInput(t *testing.T) {
	tests := []struct {
		name    string
		input   *PostInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "有効な入力（新規作成）",
			input: &PostInput{
				Name:     "Test Post",
				Category: "LLM/Tasks",
				BodyMD:   "## Content",
			},
			wantErr: false,
		},
		{
			name: "有効な入力（更新）",
			input: &PostInput{
				PostNumber: intPtr(123),
				Name:       "Test Post",
				Category:   "LLM/Tasks",
				BodyMD:     "## Content",
			},
			wantErr: false,
		},
		{
			name: "有効な入力（日本語カテゴリと日本語body_md）",
			input: &PostInput{
				Name:     "Test Post",
				Category: "LLM/タスク",
				BodyMD:   "## 内容\n\nこれは日本語のコンテンツです。",
			},
			wantErr: false,
		},
		{
			name: "有効な入力（nameに日本語）",
			input: &PostInput{
				Name:     "テスト投稿 重要 確認事項",
				Category: "LLM/Tasks",
				BodyMD:   "## Content",
			},
			wantErr: false,
		},
		{
			name: "nameが空",
			input: &PostInput{
				Name:     "",
				Category: "LLM/Tasks",
				BodyMD:   "## Content",
			},
			wantErr: true,
			errMsg:  "name cannot be empty",
		},
		{
			name: "nameが空白のみ",
			input: &PostInput{
				Name:     "   ",
				Category: "LLM/Tasks",
				BodyMD:   "## Content",
			},
			wantErr: true,
			errMsg:  "name cannot be empty",
		},
		{
			name: "nameが255バイト超過",
			input: &PostInput{
				Name:     strings.Repeat("a", 256),
				Category: "LLM/Tasks",
				BodyMD:   "## Content",
			},
			wantErr: true,
			errMsg:  "name exceeds 255 bytes",
		},
		{
			name: "nameに改行を含む",
			input: &PostInput{
				Name:     "Test\nPost",
				Category: "LLM/Tasks",
				BodyMD:   "## Content",
			},
			wantErr: true,
			errMsg:  "name contains control characters",
		},
		{
			name: "nameに/を含む",
			input: &PostInput{
				Name:     "Test/Post",
				Category: "LLM/Tasks",
				BodyMD:   "## Content",
			},
			wantErr: true,
			errMsg:  "name cannot contain /",
		},
		{
			name: "nameに全角括弧を含む",
			input: &PostInput{
				Name:     "Test（Post）",
				Category: "LLM/Tasks",
				BodyMD:   "## Content",
			},
			wantErr: true,
			errMsg:  "name cannot contain fullwidth parentheses or colon",
		},
		{
			name: "nameに全角コロンを含む",
			input: &PostInput{
				Name:     "Test：Post",
				Category: "LLM/Tasks",
				BodyMD:   "## Content",
			},
			wantErr: true,
			errMsg:  "name cannot contain fullwidth parentheses or colon",
		},
		{
			name: "categoryが空",
			input: &PostInput{
				Name:     "Test Post",
				Category: "",
				BodyMD:   "## Content",
			},
			wantErr: true,
			errMsg:  "category cannot be empty",
		},
		{
			name: "body_mdが空",
			input: &PostInput{
				Name:     "Test Post",
				Category: "LLM/Tasks",
				BodyMD:   "",
			},
			wantErr: true,
			errMsg:  "body_md cannot be empty",
		},
		{
			name: "body_mdが空白のみ",
			input: &PostInput{
				Name:     "Test Post",
				Category: "LLM/Tasks",
				BodyMD:   "   \n  ",
			},
			wantErr: true,
			errMsg:  "body_md cannot be empty",
		},
		{
			name: "body_mdが1MB超過",
			input: &PostInput{
				Name:     "Test Post",
				Category: "LLM/Tasks",
				BodyMD:   strings.Repeat("a", 1024*1024+1),
			},
			wantErr: true,
			errMsg:  "body_md exceeds 1MB",
		},
		{
			name: "body_mdが---で始まる（フロントマター衝突）",
			input: &PostInput{
				Name:     "Test Post",
				Category: "LLM/Tasks",
				BodyMD:   "---\nfrontmatter",
			},
			wantErr: true,
			errMsg:  "body_md cannot start with ---",
		},
		{
			name: "post_numberが0",
			input: &PostInput{
				PostNumber: intPtr(0),
				Name:       "Test Post",
				Category:   "LLM/Tasks",
				BodyMD:     "## Content",
			},
			wantErr: true,
			errMsg:  "post_number must be greater than 0",
		},
		{
			name: "post_numberが負数",
			input: &PostInput{
				PostNumber: intPtr(-1),
				Name:       "Test Post",
				Category:   "LLM/Tasks",
				BodyMD:     "## Content",
			},
			wantErr: true,
			errMsg:  "post_number must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Trim before validation (as done in main.go)
			TrimPostInput(tt.input)
			err := ValidatePostInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePostInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidatePostInput() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}

func TestValidatePostInputSchema(t *testing.T) {
	tests := []struct {
		name    string
		input   *PostInput
		wantErr bool
	}{
		{
			name: "有効な入力",
			input: &PostInput{
				Name:     "Test Post",
				Category: "LLM/Tasks",
				BodyMD:   "## Content",
			},
			wantErr: false,
		},
		{
			name: "post_numberあり",
			input: &PostInput{
				PostNumber: intPtr(123),
				Name:       "Test Post",
				Category:   "LLM/Tasks",
				BodyMD:     "## Content",
			},
			wantErr: false,
		},
		{
			name: "nameが空",
			input: &PostInput{
				Name:     "",
				Category: "LLM/Tasks",
				BodyMD:   "## Content",
			},
			wantErr: true,
		},
		{
			name: "categoryが空",
			input: &PostInput{
				Name:     "Test",
				Category: "",
				BodyMD:   "## Content",
			},
			wantErr: true,
		},
		{
			name: "body_mdが空",
			input: &PostInput{
				Name:     "Test",
				Category: "LLM/Tasks",
				BodyMD:   "",
			},
			wantErr: true,
		},
		{
			name: "post_numberが0",
			input: &PostInput{
				PostNumber: intPtr(0),
				Name:       "Test",
				Category:   "LLM/Tasks",
				BodyMD:     "## Content",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePostInputSchema(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePostInputSchema() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
