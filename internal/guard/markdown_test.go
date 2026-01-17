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
		{
			name: "タスクあり（1つ）",
			body: &Body{
				Background: "背景説明",
				Tasks: []Task{
					{
						ID:          "task-1",
						Title:       "タスク1のタイトル",
						Status:      TaskStatusNotStarted,
						Description: "タスク1の詳細説明",
					},
				},
			},
			want: "## 背景\n\n背景説明\n\n## タスク\n\n### タスク1のタイトル\nStatus: not_started\n\nタスク1の詳細説明",
		},
		{
			name: "タスクあり（複数、異なるステータス）",
			body: &Body{
				Background: "背景説明",
				Tasks: []Task{
					{
						ID:          "task-1",
						Title:       "タスク1",
						Status:      TaskStatusNotStarted,
						Description: "説明1",
					},
					{
						ID:          "task-2",
						Title:       "タスク2",
						Status:      TaskStatusInProgress,
						Description: "説明2",
					},
					{
						ID:          "task-3",
						Title:       "タスク3",
						Status:      TaskStatusCompleted,
						Description: "説明3",
					},
				},
			},
			want: "## 背景\n\n背景説明\n\n## タスク\n\n### タスク1\nStatus: not_started\n\n説明1\n### タスク2\nStatus: in_progress\n\n説明2\n### タスク3\nStatus: completed\n\n説明3",
		},
		{
			name: "背景 + 関連リンク + タスク",
			body: &Body{
				Background:   "背景説明",
				RelatedLinks: []string{"https://example.com/doc"},
				Tasks: []Task{
					{
						ID:          "task-1",
						Title:       "タスク1",
						Status:      TaskStatusInReview,
						Description: "レビュー中のタスク",
					},
				},
			},
			want: "## 背景\n関連リンク:\n- https://example.com/doc\n\n背景説明\n\n## タスク\n\n### タスク1\nStatus: in_review\n\nレビュー中のタスク",
		},
		{
			name: "タスクにGitHub URL（単一）",
			body: &Body{
				Background: "背景説明",
				Tasks: []Task{
					{
						ID:          "task-1",
						Title:       "タスク1",
						Status:      TaskStatusInProgress,
						Description: "タスク1の詳細説明",
						GitHubURLs:  []string{"https://github.com/owner/repo/pull/123"},
					},
				},
			},
			want: "## 背景\n\n背景説明\n\n## タスク\n\n### タスク1\nStatus: in_progress\n\nタスク1の詳細説明\n\nPull Request: https://github.com/owner/repo/pull/123",
		},
		{
			name: "タスクにGitHub URL（複数）",
			body: &Body{
				Background: "背景説明",
				Tasks: []Task{
					{
						ID:          "task-1",
						Title:       "タスク1",
						Status:      TaskStatusInProgress,
						Description: "タスク1の詳細説明",
						GitHubURLs: []string{
							"https://github.com/owner/repo/pull/123",
							"https://github.com/owner/repo/issues/456",
						},
					},
				},
			},
			want: "## 背景\n\n背景説明\n\n## タスク\n\n### タスク1\nStatus: in_progress\n\nタスク1の詳細説明\n\nPull Requests:\n- https://github.com/owner/repo/pull/123\n- https://github.com/owner/repo/issues/456\n",
		},
		{
			name: "タスクにGitHub URL未指定（省略）",
			body: &Body{
				Background: "背景説明",
				Tasks: []Task{
					{
						ID:          "task-1",
						Title:       "タスク1",
						Status:      TaskStatusNotStarted,
						Description: "タスク1の詳細説明",
					},
				},
			},
			want: "## 背景\n\n背景説明\n\n## タスク\n\n### タスク1\nStatus: not_started\n\nタスク1の詳細説明",
		},
		{
			name: "タスクにGitHub URL空配列",
			body: &Body{
				Background: "背景説明",
				Tasks: []Task{
					{
						ID:          "task-1",
						Title:       "タスク1",
						Status:      TaskStatusCompleted,
						Description: "タスク1の詳細説明",
						GitHubURLs:  []string{},
					},
				},
			},
			want: "## 背景\n\n背景説明\n\n## タスク\n\n### タスク1\nStatus: completed\n\nタスク1の詳細説明",
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
