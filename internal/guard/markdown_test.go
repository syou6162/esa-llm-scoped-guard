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
				{ID: "task-1", Title: "タスク1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "説明"},
			},
			want: "## サマリー\n- [ ] タスク1\n\n### 依存関係グラフ\n\n```mermaid\ngraph TD\n    task-1[\"タスク1\"]:::not_started\n    done([タスク完了]):::goal\n\n    task-1 --> done\n\n    classDef completed fill:#90EE90\n    classDef in_progress fill:#FFD700\n    classDef in_review fill:#FFA500\n    classDef not_started fill:#D3D3D3\n    classDef goal fill:#87CEEB,stroke:#4169E1,stroke-width:3px\n```\n",
		},
		{
			name: "タスク1つ（completed）",
			tasks: []Task{
				{ID: "task-1", Title: "タスク1", Status: TaskStatusCompleted, Summary: []string{"要約"}, Description: "説明"},
			},
			want: "## サマリー\n- [x] タスク1\n\n### 依存関係グラフ\n\n```mermaid\ngraph TD\n    task-1[\"タスク1\"]:::completed\n    done([タスク完了]):::goal\n\n    task-1 --> done\n\n    classDef completed fill:#90EE90\n    classDef in_progress fill:#FFD700\n    classDef in_review fill:#FFA500\n    classDef not_started fill:#D3D3D3\n    classDef goal fill:#87CEEB,stroke:#4169E1,stroke-width:3px\n```\n",
		},
		{
			name: "複数タスク（異なるステータス）",
			tasks: []Task{
				{ID: "task-1", Title: "タスク1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "説明"},
				{ID: "task-2", Title: "タスク2", Status: TaskStatusInProgress, Summary: []string{"要約"}, Description: "説明"},
				{ID: "task-3", Title: "タスク3", Status: TaskStatusCompleted, Summary: []string{"要約"}, Description: "説明"},
			},
			want: "## サマリー\n- [ ] タスク1\n- [ ] タスク2\n- [x] タスク3\n\n### 依存関係グラフ\n\n```mermaid\ngraph TD\n    task-1[\"タスク1\"]:::not_started\n    task-2[\"タスク2\"]:::in_progress\n    task-3[\"タスク3\"]:::completed\n    done([タスク完了]):::goal\n\n    task-1 --> done\n    task-2 --> done\n    task-3 --> done\n\n    classDef completed fill:#90EE90\n    classDef in_progress fill:#FFD700\n    classDef in_review fill:#FFA500\n    classDef not_started fill:#D3D3D3\n    classDef goal fill:#87CEEB,stroke:#4169E1,stroke-width:3px\n```\n",
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

func TestGenerateInstructionsSection(t *testing.T) {
	tests := []struct {
		name         string
		instructions []string
		want         string
	}{
		{
			name:         "指示なし",
			instructions: []string{},
			want:         "",
		},
		{
			name:         "指示1つ",
			instructions: []string{"t_wada式のTDDで開発する"},
			want:         "\n\n## 開発指針\n- t_wada式のTDDで開発する\n",
		},
		{
			name: "指示複数",
			instructions: []string{
				"t_wada式のTDDで開発する",
				"各フェーズ完了時に小まめにコミットする",
				"テストを書く前にTODOリストを作成する",
			},
			want: "\n\n## 開発指針\n- t_wada式のTDDで開発する\n- 各フェーズ完了時に小まめにコミットする\n- テストを書く前にTODOリストを作成する\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateInstructionsSection(tt.instructions)
			if got != tt.want {
				t.Errorf("generateInstructionsSection() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateTaskMarkdown(t *testing.T) {
	tests := []struct {
		name        string
		task        Task
		taskTitles  map[string]string
		reducedDeps map[string][]string
		want        string
	}{
		{
			name: "基本的なタスク",
			task: Task{
				Title:       "タスク1",
				Status:      TaskStatusNotStarted,
				Summary:     []string{"タスク1の要約"},
				Description: "タスク1の詳細説明",
			},
			taskTitles:  map[string]string{},
			reducedDeps: map[string][]string{},
			want:        "\n### タスク1\n- Status: `not_started`\n\n- 要約:\n  - タスク1の要約\n\n<details><summary>詳細を開く</summary>\n\nタスク1の詳細説明\n\n</details>\n",
		},
		{
			name: "GitHub URL（単一）",
			task: Task{
				Title:       "タスク1",
				Status:      TaskStatusInProgress,
				Summary:     []string{"タスク1の要約"},
				Description: "タスク1の詳細説明",
				GitHubURLs:  []string{"https://github.com/owner/repo/pull/123"},
			},
			taskTitles:  map[string]string{},
			reducedDeps: map[string][]string{},
			want:        "\n### タスク1\n- Status: `in_progress`\n- Pull Request: https://github.com/owner/repo/pull/123\n\n- 要約:\n  - タスク1の要約\n\n<details><summary>詳細を開く</summary>\n\nタスク1の詳細説明\n\n</details>\n",
		},
		{
			name: "GitHub URL（複数）",
			task: Task{
				Title:       "タスク1",
				Status:      TaskStatusInProgress,
				Summary:     []string{"タスク1の要約"},
				Description: "タスク1の詳細説明",
				GitHubURLs: []string{
					"https://github.com/owner/repo/pull/123",
					"https://github.com/owner/repo/issues/456",
				},
			},
			taskTitles:  map[string]string{},
			reducedDeps: map[string][]string{},
			want:        "\n### タスク1\n- Status: `in_progress`\n- Pull Requests:\n  - https://github.com/owner/repo/pull/123\n  - https://github.com/owner/repo/issues/456\n\n- 要約:\n  - タスク1の要約\n\n<details><summary>詳細を開く</summary>\n\nタスク1の詳細説明\n\n</details>\n",
		},
		{
			name: "GitHub URL空配列",
			task: Task{
				Title:       "タスク1",
				Status:      TaskStatusCompleted,
				Summary:     []string{"タスク1の要約"},
				Description: "タスク1の詳細説明",
				GitHubURLs:  []string{},
			},
			taskTitles:  map[string]string{},
			reducedDeps: map[string][]string{},
			want:        "\n### タスク1\n- Status: `completed`\n\n- 要約:\n  - タスク1の要約\n\n<details><summary>詳細を開く</summary>\n\nタスク1の詳細説明\n\n</details>\n",
		},
		{
			name: "依存関係（単一）",
			task: Task{
				ID:          "task-2",
				Title:       "タスク2",
				Status:      TaskStatusNotStarted,
				Summary:     []string{"タスク2の要約"},
				Description: "タスク2の詳細説明",
				DependsOn:   []string{"task-1"},
			},
			taskTitles: map[string]string{
				"task-1": "タスク1",
				"task-2": "タスク2",
			},
			reducedDeps: map[string][]string{
				"task-2": {"task-1"},
			},
			want: "\n### タスク2\n- Status: `not_started`\n- Depends on:\n  - `タスク1`\n\n- 要約:\n  - タスク2の要約\n\n<details><summary>詳細を開く</summary>\n\nタスク2の詳細説明\n\n</details>\n",
		},
		{
			name: "依存関係（複数）",
			task: Task{
				ID:          "task-3",
				Title:       "タスク3",
				Status:      TaskStatusNotStarted,
				Summary:     []string{"タスク3の要約"},
				Description: "タスク3の詳細説明",
				DependsOn:   []string{"task-1", "task-2"},
			},
			taskTitles: map[string]string{
				"task-1": "タスク1",
				"task-2": "タスク2",
				"task-3": "タスク3",
			},
			reducedDeps: map[string][]string{
				"task-3": {"task-1", "task-2"},
			},
			want: "\n### タスク3\n- Status: `not_started`\n- Depends on:\n  - `タスク1`\n  - `タスク2`\n\n- 要約:\n  - タスク3の要約\n\n<details><summary>詳細を開く</summary>\n\nタスク3の詳細説明\n\n</details>\n",
		},
		{
			name: "依存関係 + GitHub URL",
			task: Task{
				ID:          "task-2",
				Title:       "タスク2",
				Status:      TaskStatusInProgress,
				Summary:     []string{"タスク2の要約"},
				Description: "タスク2の詳細説明",
				DependsOn:   []string{"task-1"},
				GitHubURLs:  []string{"https://github.com/owner/repo/pull/123"},
			},
			taskTitles: map[string]string{
				"task-1": "タスク1",
				"task-2": "タスク2",
			},
			reducedDeps: map[string][]string{
				"task-2": {"task-1"},
			},
			want: "\n### タスク2\n- Status: `in_progress`\n- Depends on:\n  - `タスク1`\n- Pull Request: https://github.com/owner/repo/pull/123\n\n- 要約:\n  - タスク2の要約\n\n<details><summary>詳細を開く</summary>\n\nタスク2の詳細説明\n\n</details>\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateTaskMarkdown(tt.task, tt.taskTitles, tt.reducedDeps)
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
					Summary:     []string{"要約1"},
					Description: "タスク1の詳細説明",
				},
			},
			want: "\n\n## タスク\n\n### タスク1のタイトル\n- Status: `not_started`\n\n- 要約:\n  - 要約1\n\n<details><summary>詳細を開く</summary>\n\nタスク1の詳細説明\n\n</details>\n",
		},
		{
			name: "複数タスク",
			tasks: []Task{
				{
					Title:       "タスク1",
					Status:      TaskStatusNotStarted,
					Summary:     []string{"要約1"},
					Description: "説明1",
				},
				{
					Title:       "タスク2",
					Status:      TaskStatusInProgress,
					Summary:     []string{"要約2"},
					Description: "説明2",
				},
			},
			want: "\n\n## タスク\n\n### タスク1\n- Status: `not_started`\n\n- 要約:\n  - 要約1\n\n<details><summary>詳細を開く</summary>\n\n説明1\n\n</details>\n\n### タスク2\n- Status: `in_progress`\n\n- 要約:\n  - 要約2\n\n<details><summary>詳細を開く</summary>\n\n説明2\n\n</details>\n",
		},
		{
			name: "依存関係を含むタスク",
			tasks: []Task{
				{
					ID:          "task-1",
					Title:       "タスク1",
					Status:      TaskStatusNotStarted,
					Summary:     []string{"要約1"},
					Description: "説明1",
				},
				{
					ID:          "task-2",
					Title:       "タスク2",
					Status:      TaskStatusInProgress,
					Summary:     []string{"要約2"},
					Description: "説明2",
					DependsOn:   []string{"task-1"},
				},
			},
			want: "\n\n## タスク\n\n### タスク1\n- Status: `not_started`\n\n- 要約:\n  - 要約1\n\n<details><summary>詳細を開く</summary>\n\n説明1\n\n</details>\n\n### タスク2\n- Status: `in_progress`\n- Depends on:\n  - `タスク1`\n\n- 要約:\n  - 要約2\n\n<details><summary>詳細を開く</summary>\n\n説明2\n\n</details>\n",
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
						ID:          "task-1",
						Title:       "タスク1",
						Status:      TaskStatusNotStarted,
						Summary:     []string{"要約1"},
						Description: "説明1",
					},
					{
						ID:          "task-2",
						Title:       "タスク2",
						Status:      TaskStatusCompleted,
						Summary:     []string{"要約2"},
						Description: "説明2",
					},
				},
			},
			want: "## サマリー\n- [ ] タスク1\n- [x] タスク2\n\n### 依存関係グラフ\n\n```mermaid\ngraph TD\n    task-1[\"タスク1\"]:::not_started\n    task-2[\"タスク2\"]:::completed\n    done([タスク完了]):::goal\n\n    task-1 --> done\n    task-2 --> done\n\n    classDef completed fill:#90EE90\n    classDef in_progress fill:#FFD700\n    classDef in_review fill:#FFA500\n    classDef not_started fill:#D3D3D3\n    classDef goal fill:#87CEEB,stroke:#4169E1,stroke-width:3px\n```\n## 背景\n\n背景説明\n\n## タスク\n\n### タスク1\n- Status: `not_started`\n\n- 要約:\n  - 要約1\n\n<details><summary>詳細を開く</summary>\n\n説明1\n\n</details>\n\n### タスク2\n- Status: `completed`\n\n- 要約:\n  - 要約2\n\n<details><summary>詳細を開く</summary>\n\n説明2\n\n</details>\n",
		},
		{
			name: "全要素を含む",
			body: &Body{
				Background:   "背景説明",
				RelatedLinks: []string{"https://example.com/doc"},
				Tasks: []Task{
					{
						ID:          "task-1",
						Title:       "タスク1",
						Status:      TaskStatusInReview,
						Summary:     []string{"レビュー中の要約"},
						Description: "レビュー中のタスク",
						GitHubURLs:  []string{"https://github.com/owner/repo/pull/123"},
					},
				},
			},
			want: "## サマリー\n- [ ] タスク1\n\n### 依存関係グラフ\n\n```mermaid\ngraph TD\n    task-1[\"タスク1\"]:::in_review\n    done([タスク完了]):::goal\n\n    task-1 --> done\n\n    classDef completed fill:#90EE90\n    classDef in_progress fill:#FFD700\n    classDef in_review fill:#FFA500\n    classDef not_started fill:#D3D3D3\n    classDef goal fill:#87CEEB,stroke:#4169E1,stroke-width:3px\n```\n## 背景\n関連リンク:\n- https://example.com/doc\n\n背景説明\n\n## タスク\n\n### タスク1\n- Status: `in_review`\n- Pull Request: https://github.com/owner/repo/pull/123\n\n- 要約:\n  - レビュー中の要約\n\n<details><summary>詳細を開く</summary>\n\nレビュー中のタスク\n\n</details>\n",
		},
		{
			name: "instructionsを含む",
			body: &Body{
				Background:   "背景説明",
				Instructions: []string{"t_wada式のTDDで開発する", "小まめにコミットする"},
				Tasks: []Task{
					{
						ID:          "task-1",
						Title:       "タスク1",
						Status:      TaskStatusNotStarted,
						Summary:     []string{"要約1"},
						Description: "説明1",
					},
				},
			},
			want: "## サマリー\n- [ ] タスク1\n\n### 依存関係グラフ\n\n```mermaid\ngraph TD\n    task-1[\"タスク1\"]:::not_started\n    done([タスク完了]):::goal\n\n    task-1 --> done\n\n    classDef completed fill:#90EE90\n    classDef in_progress fill:#FFD700\n    classDef in_review fill:#FFA500\n    classDef not_started fill:#D3D3D3\n    classDef goal fill:#87CEEB,stroke:#4169E1,stroke-width:3px\n```\n## 背景\n\n背景説明\n\n## 開発指針\n- t_wada式のTDDで開発する\n- 小まめにコミットする\n\n\n## タスク\n\n### タスク1\n- Status: `not_started`\n\n- 要約:\n  - 要約1\n\n<details><summary>詳細を開く</summary>\n\n説明1\n\n</details>\n",
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

func TestGenerateMermaidGraph(t *testing.T) {
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
			name: "依存関係なし（1タスク）",
			tasks: []Task{
				{ID: "task-1", Title: "タスク1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "説明"},
			},
			want: "\n```mermaid\ngraph TD\n    task-1[\"タスク1\"]:::not_started\n    done([タスク完了]):::goal\n\n    task-1 --> done\n\n    classDef completed fill:#90EE90\n    classDef in_progress fill:#FFD700\n    classDef in_review fill:#FFA500\n    classDef not_started fill:#D3D3D3\n    classDef goal fill:#87CEEB,stroke:#4169E1,stroke-width:3px\n```\n",
		},
		{
			name: "依存関係なし（複数タスク）",
			tasks: []Task{
				{ID: "task-1", Title: "タスク1", Status: TaskStatusCompleted, Summary: []string{"要約"}, Description: "説明"},
				{ID: "task-2", Title: "タスク2", Status: TaskStatusInProgress, Summary: []string{"要約"}, Description: "説明"},
			},
			want: "\n```mermaid\ngraph TD\n    task-1[\"タスク1\"]:::completed\n    task-2[\"タスク2\"]:::in_progress\n    done([タスク完了]):::goal\n\n    task-1 --> done\n    task-2 --> done\n\n    classDef completed fill:#90EE90\n    classDef in_progress fill:#FFD700\n    classDef in_review fill:#FFA500\n    classDef not_started fill:#D3D3D3\n    classDef goal fill:#87CEEB,stroke:#4169E1,stroke-width:3px\n```\n",
		},
		{
			name: "シンプルな依存関係（A → B）",
			tasks: []Task{
				{ID: "task-1", Title: "要件定義", Status: TaskStatusCompleted, Summary: []string{"要約"}, Description: "説明"},
				{ID: "task-2", Title: "設計", Status: TaskStatusInProgress, Summary: []string{"要約"}, Description: "説明", DependsOn: []string{"task-1"}},
			},
			want: "\n```mermaid\ngraph TD\n    task-1[\"要件定義\"]:::completed\n    task-2[\"設計\"]:::in_progress\n    done([タスク完了]):::goal\n\n    task-1 --> task-2\n    task-2 --> done\n\n    classDef completed fill:#90EE90\n    classDef in_progress fill:#FFD700\n    classDef in_review fill:#FFA500\n    classDef not_started fill:#D3D3D3\n    classDef goal fill:#87CEEB,stroke:#4169E1,stroke-width:3px\n```\n",
		},
		{
			name: "複数依存先",
			tasks: []Task{
				{ID: "task-1", Title: "タスク1", Status: TaskStatusCompleted, Summary: []string{"要約"}, Description: "説明"},
				{ID: "task-2", Title: "タスク2", Status: TaskStatusCompleted, Summary: []string{"要約"}, Description: "説明"},
				{ID: "task-3", Title: "タスク3", Status: TaskStatusInProgress, Summary: []string{"要約"}, Description: "説明", DependsOn: []string{"task-1", "task-2"}},
			},
			want: "\n```mermaid\ngraph TD\n    task-1[\"タスク1\"]:::completed\n    task-2[\"タスク2\"]:::completed\n    task-3[\"タスク3\"]:::in_progress\n    done([タスク完了]):::goal\n\n    task-1 --> task-3\n    task-2 --> task-3\n    task-3 --> done\n\n    classDef completed fill:#90EE90\n    classDef in_progress fill:#FFD700\n    classDef in_review fill:#FFA500\n    classDef not_started fill:#D3D3D3\n    classDef goal fill:#87CEEB,stroke:#4169E1,stroke-width:3px\n```\n",
		},
		{
			name: "長いチェーン依存",
			tasks: []Task{
				{ID: "task-1", Title: "A", Status: TaskStatusCompleted, Summary: []string{"要約"}, Description: "説明"},
				{ID: "task-2", Title: "B", Status: TaskStatusCompleted, Summary: []string{"要約"}, Description: "説明", DependsOn: []string{"task-1"}},
				{ID: "task-3", Title: "C", Status: TaskStatusInProgress, Summary: []string{"要約"}, Description: "説明", DependsOn: []string{"task-2"}},
				{ID: "task-4", Title: "D", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "説明", DependsOn: []string{"task-3"}},
			},
			want: "\n```mermaid\ngraph TD\n    task-1[\"A\"]:::completed\n    task-2[\"B\"]:::completed\n    task-3[\"C\"]:::in_progress\n    task-4[\"D\"]:::not_started\n    done([タスク完了]):::goal\n\n    task-1 --> task-2\n    task-2 --> task-3\n    task-3 --> task-4\n    task-4 --> done\n\n    classDef completed fill:#90EE90\n    classDef in_progress fill:#FFD700\n    classDef in_review fill:#FFA500\n    classDef not_started fill:#D3D3D3\n    classDef goal fill:#87CEEB,stroke:#4169E1,stroke-width:3px\n```\n",
		},
		{
			name: "全ステータス確認",
			tasks: []Task{
				{ID: "task-1", Title: "完了タスク", Status: TaskStatusCompleted, Summary: []string{"要約"}, Description: "説明"},
				{ID: "task-2", Title: "進行中タスク", Status: TaskStatusInProgress, Summary: []string{"要約"}, Description: "説明"},
				{ID: "task-3", Title: "レビュー中タスク", Status: TaskStatusInReview, Summary: []string{"要約"}, Description: "説明"},
				{ID: "task-4", Title: "未着手タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "説明"},
			},
			want: "\n```mermaid\ngraph TD\n    task-1[\"完了タスク\"]:::completed\n    task-2[\"進行中タスク\"]:::in_progress\n    task-3[\"レビュー中タスク\"]:::in_review\n    task-4[\"未着手タスク\"]:::not_started\n    done([タスク完了]):::goal\n\n    task-1 --> done\n    task-2 --> done\n    task-3 --> done\n    task-4 --> done\n\n    classDef completed fill:#90EE90\n    classDef in_progress fill:#FFD700\n    classDef in_review fill:#FFA500\n    classDef not_started fill:#D3D3D3\n    classDef goal fill:#87CEEB,stroke:#4169E1,stroke-width:3px\n```\n",
		},
		{
			name: "タイトルにダブルクォート含む",
			tasks: []Task{
				{ID: "task-1", Title: "タスク\"1\"", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "説明"},
			},
			want: "\n```mermaid\ngraph TD\n    task-1[\"タスク&quot;1&quot;\"]:::not_started\n    done([タスク完了]):::goal\n\n    task-1 --> done\n\n    classDef completed fill:#90EE90\n    classDef in_progress fill:#FFD700\n    classDef in_review fill:#FFA500\n    classDef not_started fill:#D3D3D3\n    classDef goal fill:#87CEEB,stroke:#4169E1,stroke-width:3px\n```\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateMermaidGraph(tt.tasks)
			if got != tt.want {
				t.Errorf("generateMermaidGraph() = %q, want %q", got, tt.want)
			}
		})
	}
}
