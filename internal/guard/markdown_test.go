package guard

import (
	"testing"
)

func TestGenerateSummarySection(t *testing.T) {
	tests := []struct {
		name  string
		tasks []Task
		want  string
	}{
		{
			name:  "タスクなし",
			tasks: []Task{},
			want:  "",
		},
		{
			name: "タスク1つ（not_started）",
			tasks: []Task{
				{Title: "タスク1", Status: TaskStatusNotStarted},
			},
			want: "## サマリー\n- [ ] タスク1\n\n",
		},
		{
			name: "タスク1つ（completed）",
			tasks: []Task{
				{Title: "タスク1", Status: TaskStatusCompleted},
			},
			want: "## サマリー\n- [x] タスク1\n\n",
		},
		{
			name: "複数タスク（異なるステータス）",
			tasks: []Task{
				{Title: "タスク1", Status: TaskStatusNotStarted},
				{Title: "タスク2", Status: TaskStatusInProgress},
				{Title: "タスク3", Status: TaskStatusCompleted},
			},
			want: "## サマリー\n- [ ] タスク1\n- [ ] タスク2\n- [x] タスク3\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateSummarySection(tt.tasks)
			if got != tt.want {
				t.Errorf("generateSummarySection() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateBackgroundSection(t *testing.T) {
	tests := []struct {
		name         string
		background   string
		relatedLinks []string
		want         string
	}{
		{
			name:       "基本的な背景",
			background: "This is a background",
			want:       "## 背景\n\nThis is a background",
		},
		{
			name:       "複数行の背景",
			background: "Line 1\nLine 2\nLine 3",
			want:       "## 背景\n\nLine 1\nLine 2\nLine 3",
		},
		{
			name:       "日本語の背景",
			background: "これはタスクの背景説明です。",
			want:       "## 背景\n\nこれはタスクの背景説明です。",
		},
		{
			name:         "関連リンクあり（1つ）",
			background:   "背景説明",
			relatedLinks: []string{"https://example.com/doc"},
			want:         "## 背景\n関連リンク:\n- https://example.com/doc\n\n背景説明",
		},
		{
			name:       "関連リンクあり（複数）",
			background: "背景説明",
			relatedLinks: []string{
				"https://example.com/doc1",
				"https://example.com/doc2",
				"https://github.com/user/repo",
			},
			want: "## 背景\n関連リンク:\n- https://example.com/doc1\n- https://example.com/doc2\n- https://github.com/user/repo\n\n背景説明",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateBackgroundSection(tt.background, tt.relatedLinks)
			if got != tt.want {
				t.Errorf("generateBackgroundSection() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateTaskMarkdown(t *testing.T) {
	tests := []struct {
		name string
		task Task
		want string
	}{
		{
			name: "基本的なタスク",
			task: Task{
				Title:       "タスク1",
				Status:      TaskStatusNotStarted,
				Description: "タスク1の詳細説明",
			},
			want: "\n### タスク1\n- Status: `not_started`\n\nタスク1の詳細説明",
		},
		{
			name: "GitHub URL（単一）",
			task: Task{
				Title:       "タスク1",
				Status:      TaskStatusInProgress,
				Description: "タスク1の詳細説明",
				GitHubURLs:  []string{"https://github.com/owner/repo/pull/123"},
			},
			want: "\n### タスク1\n- Status: `in_progress`\n- Pull Request: https://github.com/owner/repo/pull/123\n\nタスク1の詳細説明",
		},
		{
			name: "GitHub URL（複数）",
			task: Task{
				Title:       "タスク1",
				Status:      TaskStatusInProgress,
				Description: "タスク1の詳細説明",
				GitHubURLs: []string{
					"https://github.com/owner/repo/pull/123",
					"https://github.com/owner/repo/issues/456",
				},
			},
			want: "\n### タスク1\n- Status: `in_progress`\n- Pull Requests:\n  - https://github.com/owner/repo/pull/123\n  - https://github.com/owner/repo/issues/456\n\nタスク1の詳細説明",
		},
		{
			name: "GitHub URL空配列",
			task: Task{
				Title:       "タスク1",
				Status:      TaskStatusCompleted,
				Description: "タスク1の詳細説明",
				GitHubURLs:  []string{},
			},
			want: "\n### タスク1\n- Status: `completed`\n\nタスク1の詳細説明",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateTaskMarkdown(tt.task)
			if got != tt.want {
				t.Errorf("generateTaskMarkdown() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateTasksSection(t *testing.T) {
	tests := []struct {
		name  string
		tasks []Task
		want  string
	}{
		{
			name:  "タスクなし",
			tasks: []Task{},
			want:  "",
		},
		{
			name: "タスク1つ",
			tasks: []Task{
				{
					Title:       "タスク1のタイトル",
					Status:      TaskStatusNotStarted,
					Description: "タスク1の詳細説明",
				},
			},
			want: "\n\n## タスク\n\n### タスク1のタイトル\n- Status: `not_started`\n\nタスク1の詳細説明",
		},
		{
			name: "複数タスク",
			tasks: []Task{
				{
					Title:       "タスク1",
					Status:      TaskStatusNotStarted,
					Description: "説明1",
				},
				{
					Title:       "タスク2",
					Status:      TaskStatusInProgress,
					Description: "説明2",
				},
			},
			want: "\n\n## タスク\n\n### タスク1\n- Status: `not_started`\n\n説明1\n### タスク2\n- Status: `in_progress`\n\n説明2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateTasksSection(tt.tasks)
			if got != tt.want {
				t.Errorf("generateTasksSection() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateMarkdown(t *testing.T) {
	tests := []struct {
		name string
		body *Body
		want string
	}{
		{
			name: "背景のみ",
			body: &Body{
				Background: "This is a background",
			},
			want: "## 背景\n\nThis is a background",
		},
		{
			name: "背景 + 関連リンク",
			body: &Body{
				Background:   "背景説明",
				RelatedLinks: []string{"https://example.com/doc"},
			},
			want: "## 背景\n関連リンク:\n- https://example.com/doc\n\n背景説明",
		},
		{
			name: "サマリー + 背景 + タスク",
			body: &Body{
				Background: "背景説明",
				Tasks: []Task{
					{
						Title:       "タスク1",
						Status:      TaskStatusNotStarted,
						Description: "説明1",
					},
					{
						Title:       "タスク2",
						Status:      TaskStatusCompleted,
						Description: "説明2",
					},
				},
			},
			want: "## サマリー\n- [ ] タスク1\n- [x] タスク2\n\n## 背景\n\n背景説明\n\n## タスク\n\n### タスク1\n- Status: `not_started`\n\n説明1\n### タスク2\n- Status: `completed`\n\n説明2",
		},
		{
			name: "全要素を含む",
			body: &Body{
				Background:   "背景説明",
				RelatedLinks: []string{"https://example.com/doc"},
				Tasks: []Task{
					{
						Title:       "タスク1",
						Status:      TaskStatusInReview,
						Description: "レビュー中のタスク",
						GitHubURLs:  []string{"https://github.com/owner/repo/pull/123"},
					},
				},
			},
			want: "## サマリー\n- [ ] タスク1\n\n## 背景\n関連リンク:\n- https://example.com/doc\n\n背景説明\n\n## タスク\n\n### タスク1\n- Status: `in_review`\n- Pull Request: https://github.com/owner/repo/pull/123\n\nレビュー中のタスク",
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
