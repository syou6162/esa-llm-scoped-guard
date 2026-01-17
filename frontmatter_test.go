package main

import (
	"strings"
	"testing"
)

func TestGenerateFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		input    *PostInput
		wantErr  bool
		contains []string
	}{
		{
			name: "基本的なフロントマター",
			input: &PostInput{
				Name:     "Test Post",
				Category: "LLM/Tasks",
				BodyMD:   "## Content",
			},
			wantErr: false,
			contains: []string{
				"---",
				"name: Test Post",
				"category: LLM/Tasks",
				"wip: false",
				"---",
				"## Content",
			},
		},
		{
			name: "タグ付きフロントマター",
			input: &PostInput{
				Name:     "Test Post",
				Category: "LLM/Tasks",
				Tags:     []string{"tag1", "tag2"},
				BodyMD:   "## Content",
			},
			wantErr: false,
			contains: []string{
				"tags:",
				"- tag1",
				"- tag2",
			},
		},
		{
			name: "特殊文字を含むname",
			input: &PostInput{
				Name:     "Test: Post",
				Category: "LLM/Tasks",
				BodyMD:   "## Content",
			},
			wantErr: false,
			contains: []string{
				"name:",
				"Test: Post",
			},
		},
		{
			name: "特殊文字を含むtag",
			input: &PostInput{
				Name:     "Test Post",
				Category: "LLM/Tasks",
				Tags:     []string{"tag:1", "tag-2"},
				BodyMD:   "## Content",
			},
			wantErr: false,
			contains: []string{
				"tags:",
				"- tag:1",
				"- tag-2",
			},
		},
		{
			name: "最終ペイロードサイズチェック（1MB超過）",
			input: &PostInput{
				Name:     "Test Post",
				Category: "LLM/Tasks",
				BodyMD:   strings.Repeat("a", 1024*1024-50), // フロントマター追加後に1MB超過
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateFrontmatter(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateFrontmatter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("GenerateFrontmatter() result does not contain %q\nGot:\n%s", want, result)
				}
			}
		})
	}
}
