package main

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
				Tags:       []string{"tag1", "tag2"},
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
			name: "tagsが10個超過",
			input: &PostInput{
				Name:     "Test Post",
				Category: "LLM/Tasks",
				Tags:     []string{"t1", "t2", "t3", "t4", "t5", "t6", "t7", "t8", "t9", "t10", "t11"},
				BodyMD:   "## Content",
			},
			wantErr: true,
			errMsg:  "tags cannot exceed 10",
		},
		{
			name: "tagが空",
			input: &PostInput{
				Name:     "Test Post",
				Category: "LLM/Tasks",
				Tags:     []string{"tag1", ""},
				BodyMD:   "## Content",
			},
			wantErr: true,
			errMsg:  "tag cannot be empty",
		},
		{
			name: "tagが空白のみ",
			input: &PostInput{
				Name:     "Test Post",
				Category: "LLM/Tasks",
				Tags:     []string{"tag1", "  "},
				BodyMD:   "## Content",
			},
			wantErr: true,
			errMsg:  "tag cannot be empty",
		},
		{
			name: "tagが50バイト超過",
			input: &PostInput{
				Name:     "Test Post",
				Category: "LLM/Tasks",
				Tags:     []string{strings.Repeat("a", 51)},
				BodyMD:   "## Content",
			},
			wantErr: true,
			errMsg:  "tag exceeds 50 bytes",
		},
		{
			name: "tagに改行を含む",
			input: &PostInput{
				Name:     "Test Post",
				Category: "LLM/Tasks",
				Tags:     []string{"tag1\ntag2"},
				BodyMD:   "## Content",
			},
			wantErr: true,
			errMsg:  "tag contains control characters",
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
