package guard

import (
	"testing"
)

func TestGenerateMarkdown(t *testing.T) {
	tests := []struct {
		name string
		body *Body
		want string
	}{
		{
			name: "基本的な背景",
			body: &Body{
				Background: "This is a background",
			},
			want: "## 背景\n\nThis is a background",
		},
		{
			name: "複数行の背景",
			body: &Body{
				Background: "Line 1\nLine 2\nLine 3",
			},
			want: "## 背景\n\nLine 1\nLine 2\nLine 3",
		},
		{
			name: "日本語の背景",
			body: &Body{
				Background: "これはタスクの背景説明です。",
			},
			want: "## 背景\n\nこれはタスクの背景説明です。",
		},
		{
			name: "関連リンクあり（1つ）",
			body: &Body{
				Background:   "背景説明",
				RelatedLinks: []string{"https://example.com/doc"},
			},
			want: "## 背景\n関連リンク:\n- https://example.com/doc\n\n背景説明",
		},
		{
			name: "関連リンクあり（複数）",
			body: &Body{
				Background: "背景説明",
				RelatedLinks: []string{
					"https://example.com/doc1",
					"https://example.com/doc2",
					"https://github.com/user/repo",
				},
			},
			want: "## 背景\n関連リンク:\n- https://example.com/doc1\n- https://example.com/doc2\n- https://github.com/user/repo\n\n背景説明",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateMarkdown(tt.body)
			if got != tt.want {
				t.Errorf("GenerateMarkdown() = %q, want %q", got, tt.want)
			}
		})
	}
}
