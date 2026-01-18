package guard

import (
	"strings"
	"testing"
)

func TestValidatePostInput_CreateNewAndPostNumber(t *testing.T) {
	postNum123 := 123
	postNum0 := 0
	postNumNeg := -1

	tests := []struct {
		name    string
		input   *PostInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "create_new=true, post_number=nil (OK: 新規作成)",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Content",
					Tasks: []Task{
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "create_new=true, post_number=123 (エラー: 両方指定)",
			input: &PostInput{
				CreateNew:  true,
				PostNumber: &postNum123,
				Name:       "Test Post",
				Category:   "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Content",
					Tasks: []Task{
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
					},
				},
			},
			wantErr: true,
			errMsg:  "cannot specify both create_new and post_number",
		},
		{
			name: "create_new=false, post_number=nil (エラー: どちらも未指定)",
			input: &PostInput{
				CreateNew: false,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Content",
					Tasks: []Task{
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
					},
				},
			},
			wantErr: true,
			errMsg:  "must specify either create_new or post_number",
		},
		{
			name: "create_new=false, post_number=123 (OK: 更新)",
			input: &PostInput{
				CreateNew:  false,
				PostNumber: &postNum123,
				Name:       "Test Post",
				Category:   "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Content",
					Tasks: []Task{
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "create_new省略, post_number=123 (OK: 更新)",
			input: &PostInput{
				PostNumber: &postNum123,
				Name:       "Test Post",
				Category:   "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Content",
					Tasks: []Task{
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "create_new=false, post_number=0 (エラー: post_number <= 0)",
			input: &PostInput{
				CreateNew:  false,
				PostNumber: &postNum0,
				Name:       "Test Post",
				Category:   "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Content",
					Tasks: []Task{
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
					},
				},
			},
			wantErr: true,
			errMsg:  "post_number must be greater than 0",
		},
		{
			name: "create_new=false, post_number=-1 (エラー: post_number <= 0)",
			input: &PostInput{
				CreateNew:  false,
				PostNumber: &postNumNeg,
				Name:       "Test Post",
				Category:   "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Content",
					Tasks: []Task{
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
					},
				},
			},
			wantErr: true,
			errMsg:  "post_number must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Content",
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "Description 1",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "nameが空",
			input: &PostInput{
				CreateNew: true,
				Name:      "",
				Category:  "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Content",
				},
			},
			wantErr: true,
			errMsg:  "name cannot be empty",
		},
		{
			name: "categoryが空",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "",
				Body: Body{
					Background: "Content",
				},
			},
			wantErr: true,
			errMsg:  "category cannot be empty",
		},
		{
			name: "categoryが日付形式で終わらない",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks",
				Body: Body{
					Background: "Content",
				},
			},
			wantErr: true,
			errMsg:  "category must end with /yyyy/mm/dd format",
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

// TestValidatePostInput_Body はBody構造体のバリデーションをテストします
func TestValidatePostInput_Body(t *testing.T) {
	tests := []struct {
		name    string
		input   *PostInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "有効な入力（backgroundとtasks）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "This is a background",
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "Description 1",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "backgroundが空",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "",
				},
			},
			wantErr: true,
			errMsg:  "background cannot be empty",
		},
		{
			name: "backgroundが空白のみ",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "   \n  ",
				},
			},
			wantErr: true,
			errMsg:  "background cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

func TestValidatePostInputSchema(t *testing.T) {
	tests := []struct {
		name    string
		input   *PostInput
		wantErr bool
	}{
		{
			name: "有効な入力",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Content",
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "Description 1",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "nameが空",
			input: &PostInput{
				CreateNew: true,
				Name:      "",
				Category:  "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Content",
				},
			},
			wantErr: true,
		},
		{
			name: "backgroundが空",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test",
				Category:  "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "",
				},
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

// TestValidatePostInput_GitHubURLs はタスクのGitHub URLsのバリデーションをテストします
func TestValidatePostInput_GitHubURLs(t *testing.T) {
	tests := []struct {
		name    string
		input   *PostInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "有効なGitHub URL（単一）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "Description",
							GitHubURLs:  []string{"https://github.com/owner/repo/pull/123"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "有効なGitHub URL（複数）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "Description",
							GitHubURLs: []string{
								"https://github.com/owner/repo/pull/123",
								"https://github.com/owner/repo/issues/456",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "github_urls省略（OK）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "Description",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "空の配列（OK）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "Description",
							GitHubURLs:  []string{},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "無効なドメイン（gitlab.com）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "Description",
							GitHubURLs:  []string{"https://gitlab.com/owner/repo/pull/123"},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "must be a valid GitHub URL",
		},
		{
			name: "HTTPスキーム（許可しない）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "Description",
							GitHubURLs:  []string{"http://github.com/owner/repo/pull/123"},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "must be a valid GitHub URL",
		},
		{
			name: "サブドメイン（許可しない）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "Description",
							GitHubURLs:  []string{"https://api.github.com/repos/owner/repo"},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "must be a valid GitHub URL",
		},
		{
			name: "backgroundに# h1見出し（許可しない）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "# This is h1 heading",
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "Description",
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "background cannot contain heading markers",
		},
		{
			name: "backgroundに## h2見出し（許可しない）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "Some text\n## This is h2 heading\nMore text",
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "Description",
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "background cannot contain heading markers",
		},
		{
			name: "descriptionに# h1見出し（許可しない）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "# This is h1",
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "description cannot contain heading markers",
		},
		{
			name: "descriptionに## h2見出し（許可しない）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "## This is h2",
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "description cannot contain heading markers",
		},
		{
			name: "descriptionに### h3見出し（許可しない）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "### This is h3",
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "description cannot contain heading markers",
		},
		{
			name: "backgroundに#### h4見出し（許可する）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "#### This is h4\nSome content",
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "Description",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "descriptionに#### h4見出し（許可する）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "#### This is h4 heading",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "descriptionに##### h5見出し（許可する）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "##### This is h5 heading",
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

// TestValidateSummary はSummaryフィールドのバリデーションをテストします
func TestValidateSummary(t *testing.T) {
	tests := []struct {
		name    string
		summary []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "有効（1行）",
			summary: []string{"タスクの要約1"},
			wantErr: false,
		},
		{
			name:    "有効（2行）",
			summary: []string{"タスクの要約1", "タスクの要約2"},
			wantErr: false,
		},
		{
			name:    "有効（3行）",
			summary: []string{"タスクの要約1", "タスクの要約2", "タスクの要約3"},
			wantErr: false,
		},
		{
			name:    "有効（140字ちょうど）",
			summary: []string{strings.Repeat("あ", 140)},
			wantErr: false,
		},
		{
			name:    "エラー（0行）",
			summary: []string{},
			wantErr: true,
			errMsg:  "summary must have 1-3 items, got 0",
		},
		{
			name:    "エラー（4行）",
			summary: []string{"要約1", "要約2", "要約3", "要約4"},
			wantErr: true,
			errMsg:  "summary must have 1-3 items, got 4",
		},
		{
			name:    "エラー（141字）",
			summary: []string{strings.Repeat("あ", 141)},
			wantErr: true,
			errMsg:  "summary line 1 exceeds 140 characters",
		},
		{
			name:    "エラー（2行目が141字）",
			summary: []string{"正常な行", strings.Repeat("あ", 141)},
			wantErr: true,
			errMsg:  "summary line 2 exceeds 140 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSummary(tt.summary)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSummary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateSummary() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

// TestValidatePostInput_Summary はタスクのSummaryフィールドのバリデーションをテストします
func TestValidatePostInput_Summary(t *testing.T) {
	tests := []struct {
		name    string
		input   *PostInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "有効なSummary（1行）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"タスクの要約"},
							Description: "Description",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "有効なSummary（3行）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約1", "要約2", "要約3"},
							Description: "Description",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Summary空配列（エラー）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1",
							Status:      TaskStatusNotStarted,
							Summary:     []string{},
							Description: "Description",
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "summary must have 1-3 items",
		},
		{
			name: "Summary4行（エラー）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約1", "要約2", "要約3", "要約4"},
							Description: "Description",
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "summary must have 1-3 items",
		},
		{
			name: "Summary141字（エラー）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1",
							Status:      TaskStatusNotStarted,
							Summary:     []string{strings.Repeat("あ", 141)},
							Description: "Description",
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "summary line 1 exceeds 140 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

func TestValidatePostInput_DependsOn(t *testing.T) {
	tests := []struct {
		name    string
		input   *PostInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "有効な依存関係（単一）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
						{ID: "task-2", Title: "Task 2", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-1"}},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "有効な依存関係（複数）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
						{ID: "task-2", Title: "Task 2", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
						{ID: "task-3", Title: "Task 3", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-1", "task-2"}},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "依存なし（省略）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "依存なし（空配列）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{}},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "存在しないタスクIDを参照",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-999"}},
					},
				},
			},
			wantErr: true,
			errMsg:  "task[0].depends_on references non-existent task ID: task-999",
		},
		{
			name: "空文字のタスクID",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{""}},
					},
				},
			},
			wantErr: true,
			errMsg:  "task[0].depends_on[0]: empty task ID",
		},
		{
			name: "自己参照",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-1"}},
					},
				},
			},
			wantErr: true,
			errMsg:  "task[0].depends_on: self-reference is not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

func TestValidatePostInput_CyclicDependency(t *testing.T) {
	tests := []struct {
		name    string
		input   *PostInput
		wantErr bool
		errMsg  string
	}{
		// エラーになるべきケース
		{
			name: "2タスク循環（A → B → A）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-2"}},
						{ID: "task-2", Title: "Task 2", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-1"}},
					},
				},
			},
			wantErr: true,
			errMsg:  "circular dependency detected",
		},
		{
			name: "3タスク循環（A → B → C → A）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-3"}},
						{ID: "task-2", Title: "Task 2", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-1"}},
						{ID: "task-3", Title: "Task 3", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-2"}},
					},
				},
			},
			wantErr: true,
			errMsg:  "circular dependency detected",
		},
		{
			name: "長い循環（A → B → C → D → E → A）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-5"}},
						{ID: "task-2", Title: "Task 2", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-1"}},
						{ID: "task-3", Title: "Task 3", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-2"}},
						{ID: "task-4", Title: "Task 4", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-3"}},
						{ID: "task-5", Title: "Task 5", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-4"}},
					},
				},
			},
			wantErr: true,
			errMsg:  "circular dependency detected",
		},
		{
			name: "部分的循環（A → B, B → C → B）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-2"}},
						{ID: "task-2", Title: "Task 2", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-3"}},
						{ID: "task-3", Title: "Task 3", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-2"}},
					},
				},
			},
			wantErr: true,
			errMsg:  "circular dependency detected",
		},
		{
			name: "複数依存先の1つが循環（A → [B, C], B → C → B）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-2", "task-3"}},
						{ID: "task-2", Title: "Task 2", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-3"}},
						{ID: "task-3", Title: "Task 3", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-2"}},
					},
				},
			},
			wantErr: true,
			errMsg:  "circular dependency detected",
		},
		{
			name: "間接的な循環（A → B → C → D → B）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-2"}},
						{ID: "task-2", Title: "Task 2", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-3"}},
						{ID: "task-3", Title: "Task 3", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-4"}},
						{ID: "task-4", Title: "Task 4", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-2"}},
					},
				},
			},
			wantErr: true,
			errMsg:  "circular dependency detected",
		},
		// OKになるべきケース
		{
			name: "循環なし（チェーン A → B → C）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-2"}},
						{ID: "task-2", Title: "Task 2", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-3"}},
						{ID: "task-3", Title: "Task 3", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "循環なし（逆順チェーン C → B → A）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{ID: "task-3", Title: "Task 3", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-2"}},
						{ID: "task-2", Title: "Task 2", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-1"}},
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "循環なし（複数依存 C depends on [A, B]）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
						{ID: "task-2", Title: "Task 2", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
						{ID: "task-3", Title: "Task 3", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-1", "task-2"}},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "循環なし（ダイヤモンド D → [B, C], B → A, C → A）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
						{ID: "task-2", Title: "Task 2", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-1"}},
						{ID: "task-3", Title: "Task 3", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-1"}},
						{ID: "task-4", Title: "Task 4", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-2", "task-3"}},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "循環なし（独立したタスクあり A → B, C単独, D → E）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-2"}},
						{ID: "task-2", Title: "Task 2", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
						{ID: "task-3", Title: "Task 3", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
						{ID: "task-4", Title: "Task 4", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-5"}},
						{ID: "task-5", Title: "Task 5", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "循環なし（依存なしタスクのみ）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
						{ID: "task-2", Title: "Task 2", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
						{ID: "task-3", Title: "Task 3", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "循環なし（1タスクのみ、依存なし）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "循環なし（全タスクが1つに依存 B → A, C → A, D → A）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
						{ID: "task-2", Title: "Task 2", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-1"}},
						{ID: "task-3", Title: "Task 3", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-1"}},
						{ID: "task-4", Title: "Task 4", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-1"}},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "循環なし（複雑なDAG）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
						{ID: "task-2", Title: "Task 2", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-1"}},
						{ID: "task-3", Title: "Task 3", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-1"}},
						{ID: "task-4", Title: "Task 4", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-2", "task-3"}},
						{ID: "task-5", Title: "Task 5", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-2"}},
						{ID: "task-6", Title: "Task 6", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-4", "task-5"}},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
