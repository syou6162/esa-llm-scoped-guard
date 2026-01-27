package guard

import (
	"errors"
	"strings"
	"testing"
)

func TestValidatePostInput_CreateNewAndPostNumber(t *testing.T) {
	postNum123 := 123
	postNum0 := 0
	postNumNeg := -1

	tests := []struct {
		name        string
		input       *PostInput
		wantErr     bool
		wantErrCode ValidationErrorCode
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
						{ID: "task-1", Title: "Task 1: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
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
						{ID: "task-1", Title: "Task 1: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeMutuallyExclusive,
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
						{ID: "task-1", Title: "Task 1: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeMissingRequired,
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
						{ID: "task-1", Title: "Task 1: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
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
						{ID: "task-1", Title: "Task 1: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
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
						{ID: "task-1", Title: "Task 1: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeInvalidValue,
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
						{ID: "task-1", Title: "Task 1: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeInvalidValue,
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
			if tt.wantErr {
				var ve *ValidationError
				if !errors.As(err, &ve) {
					t.Errorf("ValidatePostInput() error = %v, expected ValidationError", err)
					return
				}
				if ve.Code() != tt.wantErrCode {
					t.Errorf("ValidationError.Code() = %v, want %v", ve.Code(), tt.wantErrCode)
				}
			}
		})
	}
}

func TestValidatePostInput(t *testing.T) {
	tests := []struct {
		name        string
		input       *PostInput
		wantErr     bool
		wantErrCode ValidationErrorCode
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
							Title:       "Task 1: タスク",
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
			wantErr:     true,
			wantErrCode: ErrCodeFieldEmpty,
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
			wantErr:     true,
			wantErrCode: ErrCodeCategoryEmpty,
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
			wantErr:     true,
			wantErrCode: ErrCodeCategoryInvalidDateSuffix,
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
			if tt.wantErr {
				var ve *ValidationError
				if !errors.As(err, &ve) {
					t.Errorf("ValidatePostInput() error = %v, expected ValidationError", err)
					return
				}
				if ve.Code() != tt.wantErrCode {
					t.Errorf("ValidationError.Code() = %v, want %v", ve.Code(), tt.wantErrCode)
				}
			}
		})
	}
}

// TestValidatePostInput_Body はBody構造体のバリデーションをテストします
func TestValidatePostInput_Body(t *testing.T) {
	tests := []struct {
		name        string
		input       *PostInput
		wantErr     bool
		wantErrCode ValidationErrorCode
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
							Title:       "Task 1: タスク",
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
			wantErr:     true,
			wantErrCode: ErrCodeFieldEmpty,
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
			wantErr:     true,
			wantErrCode: ErrCodeFieldEmpty,
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
			if tt.wantErr {
				var ve *ValidationError
				if !errors.As(err, &ve) {
					t.Errorf("ValidatePostInput() error = %v, expected ValidationError", err)
					return
				}
				if ve.Code() != tt.wantErrCode {
					t.Errorf("ValidationError.Code() = %v, want %v", ve.Code(), tt.wantErrCode)
				}
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
							Title:       "Task 1: タスク",
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
		name        string
		input       *PostInput
		wantErr     bool
		wantErrCode ValidationErrorCode
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
							Title:       "Task 1: タスク",
							Status:      TaskStatusInProgress,
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
							Title:       "Task 1: タスク",
							Status:      TaskStatusInProgress,
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
							Title:       "Task 1: タスク",
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
							Title:       "Task 1: タスク",
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
							Title:       "Task 1: タスク",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "Description",
							GitHubURLs:  []string{"https://gitlab.com/owner/repo/pull/123"},
						},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeFieldInvalidFormat,
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
							Title:       "Task 1: タスク",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "Description",
							GitHubURLs:  []string{"http://github.com/owner/repo/pull/123"},
						},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeFieldInvalidFormat,
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
							Title:       "Task 1: タスク",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "Description",
							GitHubURLs:  []string{"https://api.github.com/repos/owner/repo"},
						},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeFieldInvalidFormat,
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
							Title:       "Task 1: タスク",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "Description",
						},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeFieldInvalidFormat,
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
							Title:       "Task 1: タスク",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "Description",
						},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeFieldInvalidFormat,
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
							Title:       "Task 1: タスク",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "# This is h1",
						},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeFieldInvalidFormat,
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
							Title:       "Task 1: タスク",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "## This is h2",
						},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeFieldInvalidFormat,
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
							Title:       "Task 1: タスク",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "### This is h3",
						},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeFieldInvalidFormat,
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
							Title:       "Task 1: タスク",
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
							Title:       "Task 1: タスク",
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
							Title:       "Task 1: タスク",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "##### This is h5 heading",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "PRリンクあり + not_started（エラー）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1: タスク",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "Description",
							GitHubURLs:  []string{"https://github.com/owner/repo/pull/123"},
						},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeFieldInvalidFormat,
		},
		{
			name: "PRリンクあり + in_progress（OK）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1: タスク",
							Status:      TaskStatusInProgress,
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
			name: "PRリンクあり + in_review（OK）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1: タスク",
							Status:      TaskStatusInReview,
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
			name: "PRリンクあり + completed（OK）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1: タスク",
							Status:      TaskStatusCompleted,
							Summary:     []string{"要約"},
							Description: "Description",
							GitHubURLs:  []string{"https://github.com/owner/repo/pull/123"},
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
			if tt.wantErr {
				var ve *ValidationError
				if !errors.As(err, &ve) {
					t.Errorf("Expected ValidationError, got %T", err)
					return
				}
				if ve.Code() != tt.wantErrCode {
					t.Errorf("ValidationError.Code() = %v, want %v", ve.Code(), tt.wantErrCode)
				}
			}
		})
	}
}

// TestValidateSummary はSummaryフィールドのバリデーションをテストします
func TestValidateSummary(t *testing.T) {
	tests := []struct {
		name        string
		summary     []string
		wantErr     bool
		wantErrCode ValidationErrorCode
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
			name:        "エラー（0行）",
			summary:     []string{},
			wantErr:     true,
			wantErrCode: ErrCodeFieldInvalidFormat,
		},
		{
			name:        "エラー（4行）",
			summary:     []string{"要約1", "要約2", "要約3", "要約4"},
			wantErr:     true,
			wantErrCode: ErrCodeFieldInvalidFormat,
		},
		{
			name:        "エラー（141字）",
			summary:     []string{strings.Repeat("あ", 141)},
			wantErr:     true,
			wantErrCode: ErrCodeFieldTooLong,
		},
		{
			name:        "エラー（2行目が141字）",
			summary:     []string{"正常な行", strings.Repeat("あ", 141)},
			wantErr:     true,
			wantErrCode: ErrCodeFieldTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSummary(tt.summary)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSummary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				var ve *ValidationError
				if !errors.As(err, &ve) {
					t.Errorf("Expected ValidationError, got %T", err)
					return
				}
				if ve.Code() != tt.wantErrCode {
					t.Errorf("ValidationError.Code() = %v, want %v", ve.Code(), tt.wantErrCode)
				}
			}
		})
	}
}

// TestValidatePostInput_Summary はタスクのSummaryフィールドのバリデーションをテストします
func TestValidatePostInput_Summary(t *testing.T) {
	tests := []struct {
		name        string
		input       *PostInput
		wantErr     bool
		wantErrCode ValidationErrorCode
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
							Title:       "Task 1: タスク",
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
							Title:       "Task 1: タスク",
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
							Title:       "Task 1: タスク",
							Status:      TaskStatusNotStarted,
							Summary:     []string{},
							Description: "Description",
						},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeFieldInvalidFormat,
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
							Title:       "Task 1: タスク",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約1", "要約2", "要約3", "要約4"},
							Description: "Description",
						},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeFieldInvalidFormat,
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
							Title:       "Task 1: タスク",
							Status:      TaskStatusNotStarted,
							Summary:     []string{strings.Repeat("あ", 141)},
							Description: "Description",
						},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeFieldTooLong,
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
			if tt.wantErr {
				var ve *ValidationError
				if !errors.As(err, &ve) {
					t.Errorf("Expected ValidationError, got %T", err)
					return
				}
				if ve.Code() != tt.wantErrCode {
					t.Errorf("ValidationError.Code() = %v, want %v", ve.Code(), tt.wantErrCode)
				}
			}
		})
	}
}

func TestValidatePostInput_DependsOn(t *testing.T) {
	tests := []struct {
		name        string
		input       *PostInput
		wantErr     bool
		wantErrCode ValidationErrorCode
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
						{ID: "task-1", Title: "Task 1: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
						{ID: "task-2", Title: "Task 2: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-1"}},
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
						{ID: "task-1", Title: "Task 1: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
						{ID: "task-2", Title: "Task 2: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
						{ID: "task-3", Title: "Task 3: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-1", "task-2"}},
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
						{ID: "task-1", Title: "Task 1: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
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
						{ID: "task-1", Title: "Task 1: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{}},
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
						{ID: "task-1", Title: "Task 1: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-999"}},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeNonExistentRef,
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
						{ID: "task-1", Title: "Task 1: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{""}},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeFieldEmpty,
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
						{ID: "task-1", Title: "Task 1: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-1"}},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeSelfReference,
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
			if tt.wantErr {
				var ve *ValidationError
				if !errors.As(err, &ve) {
					t.Errorf("Expected ValidationError, got %T", err)
					return
				}
				if ve.Code() != tt.wantErrCode {
					t.Errorf("ValidationError.Code() = %v, want %v", ve.Code(), tt.wantErrCode)
				}
			}
		})
	}
}

func TestValidatePostInput_CyclicDependency(t *testing.T) {
	tests := []struct {
		name        string
		input       *PostInput
		wantErr     bool
		wantErrCode ValidationErrorCode
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
						{ID: "task-1", Title: "Task 1: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-2"}},
						{ID: "task-2", Title: "Task 2: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-1"}},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeCircularDependency,
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
						{ID: "task-1", Title: "Task 1: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-3"}},
						{ID: "task-2", Title: "Task 2: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-1"}},
						{ID: "task-3", Title: "Task 3: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-2"}},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeCircularDependency,
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
						{ID: "task-1", Title: "Task 1: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-5"}},
						{ID: "task-2", Title: "Task 2: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-1"}},
						{ID: "task-3", Title: "Task 3: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-2"}},
						{ID: "task-4", Title: "Task 4: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-3"}},
						{ID: "task-5", Title: "Task 5: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-4"}},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeCircularDependency,
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
						{ID: "task-1", Title: "Task 1: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-2"}},
						{ID: "task-2", Title: "Task 2: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-3"}},
						{ID: "task-3", Title: "Task 3: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-2"}},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeCircularDependency,
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
						{ID: "task-1", Title: "Task 1: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-2", "task-3"}},
						{ID: "task-2", Title: "Task 2: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-3"}},
						{ID: "task-3", Title: "Task 3: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-2"}},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeCircularDependency,
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
						{ID: "task-1", Title: "Task 1: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-2"}},
						{ID: "task-2", Title: "Task 2: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-3"}},
						{ID: "task-3", Title: "Task 3: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-4"}},
						{ID: "task-4", Title: "Task 4: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-2"}},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeCircularDependency,
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
						{ID: "task-1", Title: "Task 1: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-2"}},
						{ID: "task-2", Title: "Task 2: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-3"}},
						{ID: "task-3", Title: "Task 3: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
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
						{ID: "task-3", Title: "Task 3: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-2"}},
						{ID: "task-2", Title: "Task 2: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-1"}},
						{ID: "task-1", Title: "Task 1: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
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
						{ID: "task-1", Title: "Task 1: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
						{ID: "task-2", Title: "Task 2: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
						{ID: "task-3", Title: "Task 3: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-1", "task-2"}},
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
						{ID: "task-1", Title: "Task 1: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
						{ID: "task-2", Title: "Task 2: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-1"}},
						{ID: "task-3", Title: "Task 3: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-1"}},
						{ID: "task-4", Title: "Task 4: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-2", "task-3"}},
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
						{ID: "task-1", Title: "Task 1: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-2"}},
						{ID: "task-2", Title: "Task 2: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
						{ID: "task-3", Title: "Task 3: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
						{ID: "task-4", Title: "Task 4: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-5"}},
						{ID: "task-5", Title: "Task 5: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
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
						{ID: "task-1", Title: "Task 1: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
						{ID: "task-2", Title: "Task 2: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
						{ID: "task-3", Title: "Task 3: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
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
						{ID: "task-1", Title: "Task 1: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
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
						{ID: "task-1", Title: "Task 1: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
						{ID: "task-2", Title: "Task 2: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-1"}},
						{ID: "task-3", Title: "Task 3: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-1"}},
						{ID: "task-4", Title: "Task 4: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-1"}},
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
						{ID: "task-1", Title: "Task 1: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
						{ID: "task-2", Title: "Task 2: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-1"}},
						{ID: "task-3", Title: "Task 3: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-1"}},
						{ID: "task-4", Title: "Task 4: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-2", "task-3"}},
						{ID: "task-5", Title: "Task 5: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-2"}},
						{ID: "task-6", Title: "Task 6: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc", DependsOn: []string{"task-4", "task-5"}},
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
			if tt.wantErr {
				var ve *ValidationError
				if !errors.As(err, &ve) {
					t.Errorf("Expected ValidationError, got %T", err)
					return
				}
				if ve.Code() != tt.wantErrCode {
					t.Errorf("ValidationError.Code() = %v, want %v", ve.Code(), tt.wantErrCode)
				}
			}
		})
	}
}

// TestValidateInstructions はInstructionsフィールドのバリデーションをテストします
func TestValidateInstructions(t *testing.T) {
	tests := []struct {
		name         string
		instructions []string
		wantErr      bool
		wantErrCode  ValidationErrorCode
	}{
		{
			name:         "有効（0項目）",
			instructions: []string{},
			wantErr:      false,
		},
		{
			name:         "有効（1項目）",
			instructions: []string{"t_wada式のTDDで開発する"},
			wantErr:      false,
		},
		{
			name: "有効（10項目）",
			instructions: []string{
				"指示1", "指示2", "指示3", "指示4", "指示5",
				"指示6", "指示7", "指示8", "指示9", "指示10",
			},
			wantErr: false,
		},
		{
			name: "エラー（11項目）",
			instructions: []string{
				"指示1", "指示2", "指示3", "指示4", "指示5",
				"指示6", "指示7", "指示8", "指示9", "指示10", "指示11",
			},
			wantErr:     true,
			wantErrCode: ErrCodeFieldInvalidFormat,
		},
		{
			name:         "有効（500文字ちょうど）",
			instructions: []string{strings.Repeat("あ", 500)},
			wantErr:      false,
		},
		{
			name:         "エラー（501文字）",
			instructions: []string{strings.Repeat("あ", 501)},
			wantErr:      true,
			wantErrCode:  ErrCodeFieldTooLong,
		},
		{
			name:         "エラー（# h1見出し）",
			instructions: []string{"# This is h1"},
			wantErr:      true,
			wantErrCode:  ErrCodeFieldInvalidFormat,
		},
		{
			name:         "エラー（## h2見出し）",
			instructions: []string{"## This is h2"},
			wantErr:      true,
			wantErrCode:  ErrCodeFieldInvalidFormat,
		},
		{
			name:         "有効（### h3見出し以下は許可）",
			instructions: []string{"### This is h3"},
			wantErr:      false,
		},
		{
			name:         "有効（#### h4見出し）",
			instructions: []string{"#### This is h4"},
			wantErr:      false,
		},
		{
			name:         "エラー（先頭に - ）",
			instructions: []string{"- リスト項目"},
			wantErr:      true,
			wantErrCode:  ErrCodeFieldInvalidFormat,
		},
		{
			name:         "エラー（先頭に * ）",
			instructions: []string{"* リスト項目"},
			wantErr:      true,
			wantErrCode:  ErrCodeFieldInvalidFormat,
		},
		{
			name:         "エラー（先頭に + ）",
			instructions: []string{"+ リスト項目"},
			wantErr:      true,
			wantErrCode:  ErrCodeFieldInvalidFormat,
		},
		{
			name:         "エラー（先頭に 1. ）",
			instructions: []string{"1. リスト項目"},
			wantErr:      true,
			wantErrCode:  ErrCodeFieldInvalidFormat,
		},
		{
			name:         "エラー（先頭に 2. ）",
			instructions: []string{"2. リスト項目"},
			wantErr:      true,
			wantErrCode:  ErrCodeFieldInvalidFormat,
		},
		{
			name:         "有効（途中に - があるのはOK）",
			instructions: []string{"これは - を含む文章"},
			wantErr:      false,
		},
		{
			name:         "エラー（空白+リストマーカー）",
			instructions: []string{"  - リスト項目"},
			wantErr:      true,
			wantErrCode:  ErrCodeFieldInvalidFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInstructions(tt.instructions)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateInstructions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				var ve *ValidationError
				if !errors.As(err, &ve) {
					t.Errorf("Expected ValidationError, got %T", err)
					return
				}
				if ve.Code() != tt.wantErrCode {
					t.Errorf("ValidationError.Code() = %v, want %v", ve.Code(), tt.wantErrCode)
				}
			}
		})
	}
}

// TestValidatePostInput_Instructions はInstructionsフィールドのバリデーションを統合的にテストします
func TestValidatePostInput_Instructions(t *testing.T) {
	tests := []struct {
		name        string
		input       *PostInput
		wantErr     bool
		wantErrCode ValidationErrorCode
	}{
		{
			name: "有効（instructionsなし - omitempty）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "Background",
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1: タスク",
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
			name: "有効（instructions空配列）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background:   "Background",
					Instructions: []string{},
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1: タスク",
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
			name: "有効（instructions 1項目）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background:   "Background",
					Instructions: []string{"t_wada式のTDDで開発する"},
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1: タスク",
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
			name: "有効（instructions 10項目）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "Background",
					Instructions: []string{
						"指示1", "指示2", "指示3", "指示4", "指示5",
						"指示6", "指示7", "指示8", "指示9", "指示10",
					},
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1: タスク",
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
			name: "エラー（instructions 11項目）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "Background",
					Instructions: []string{
						"指示1", "指示2", "指示3", "指示4", "指示5",
						"指示6", "指示7", "指示8", "指示9", "指示10", "指示11",
					},
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1: タスク",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "Description",
						},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeFieldInvalidFormat,
		},
		{
			name: "エラー（501文字）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background:   "Background",
					Instructions: []string{strings.Repeat("あ", 501)},
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1: タスク",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "Description",
						},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeFieldTooLong,
		},
		{
			name: "エラー（見出しマーカー #）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background:   "Background",
					Instructions: []string{"# This is h1"},
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1: タスク",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "Description",
						},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeFieldInvalidFormat,
		},
		{
			name: "エラー（見出しマーカー ##）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background:   "Background",
					Instructions: []string{"## This is h2"},
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1: タスク",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "Description",
						},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeFieldInvalidFormat,
		},
		{
			name: "エラー（リストマーカー - ）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background:   "Background",
					Instructions: []string{"- リスト項目"},
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1: タスク",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "Description",
						},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeFieldInvalidFormat,
		},
		{
			name: "エラー（リストマーカー * ）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background:   "Background",
					Instructions: []string{"* リスト項目"},
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1: タスク",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "Description",
						},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeFieldInvalidFormat,
		},
		{
			name: "エラー（リストマーカー + ）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background:   "Background",
					Instructions: []string{"+ リスト項目"},
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1: タスク",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "Description",
						},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeFieldInvalidFormat,
		},
		{
			name: "エラー（番号付きリストマーカー 1. ）",
			input: &PostInput{
				CreateNew: true,
				Name:      "Test Post",
				Category:  "LLM/Tasks/2026/01/18",
				Body: Body{
					Background:   "Background",
					Instructions: []string{"1. リスト項目"},
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1: タスク",
							Status:      TaskStatusNotStarted,
							Summary:     []string{"要約"},
							Description: "Description",
						},
					},
				},
			},
			wantErr:     true,
			wantErrCode: ErrCodeFieldInvalidFormat,
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
			if tt.wantErr {
				var ve *ValidationError
				if !errors.As(err, &ve) {
					t.Errorf("Expected ValidationError, got %T", err)
					return
				}
				if ve.Code() != tt.wantErrCode {
					t.Errorf("ValidationError.Code() = %v, want %v", ve.Code(), tt.wantErrCode)
				}
			}
		})
	}
}

func TestValidateTaskTitleFormat(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		index       int
		wantNumber  int
		wantName    string
		wantErr     bool
		wantErrCode ValidationErrorCode
	}{
		// 正常系
		{"有効な形式（Task 1: タスク名）", "Task 1: タスク名", 0, 1, "タスク名", false, ""},
		{"有効な形式（番号2）", "Task 2: Second task", 1, 2, "Second task", false, ""},
		{"有効な形式（長いタスク名）", "Task 10: Very long task name with spaces", 9, 10, "Very long task name with spaces", false, ""},
		{"有効な形式（番号99）", "Task 99: タスク", 98, 99, "タスク", false, ""},

		// 異常系: プレフィックスなし
		{"プレフィックスなし", "タスク名のみ", 0, 0, "", true, ErrCodeTaskTitleInvalidPrefix},
		{"タスクIDっぽいが違う", "task-1: タスク名", 0, 0, "", true, ErrCodeTaskTitleInvalidPrefix},

		// 異常系: 大文字小文字エラー
		{"全大文字TASK", "TASK 1: タスク名", 0, 0, "", true, ErrCodeTaskTitleInvalidPrefix},
		{"全小文字task", "task 1: タスク名", 0, 0, "", true, ErrCodeTaskTitleInvalidPrefix},

		// 異常系: スペース不正
		{"Taskと番号の間にスペースなし", "Task1: タスク名", 0, 0, "", true, ErrCodeTaskTitleInvalidPrefix},
		{"番号の後にスペースなし", "Task 1:タスク名", 0, 0, "", true, ErrCodeTaskTitleInvalidPrefix},
		{"スペース2個", "Task  1: タスク名", 0, 0, "", true, ErrCodeTaskTitleInvalidPrefix},

		// 異常系: 番号形式エラー
		{"小数番号", "Task 2.1: タスク名", 0, 0, "", true, ErrCodeTaskTitleInvalidPrefix},
		{"範囲番号", "Task 6-7: タスク名", 0, 0, "", true, ErrCodeTaskTitleInvalidPrefix},
		{"アルファベット付き番号", "Task 2A: タスク名", 0, 0, "", true, ErrCodeTaskTitleInvalidPrefix},
		{"アルファベット付き番号2", "Task A2: タスク名", 0, 0, "", true, ErrCodeTaskTitleInvalidPrefix},
		{"番号なし", "Task : タスク名", 0, 0, "", true, ErrCodeTaskTitleInvalidPrefix},
		{"先頭ゼロ1桁", "Task 01: タスク名", 0, 0, "", true, ErrCodeTaskTitleInvalidPrefix},
		{"先頭ゼロ2桁", "Task 001: タスク名", 0, 0, "", true, ErrCodeTaskTitleInvalidPrefix},
		{"先頭ゼロ（10）", "Task 010: タスク名", 0, 0, "", true, ErrCodeTaskTitleInvalidPrefix},
		{"番号0", "Task 0: タスク名", 0, 0, "", true, ErrCodeTaskTitleInvalidPrefix},

		// 異常系: タスク名が空
		{"タスク名が空", "Task 1: ", 0, 0, "", true, ErrCodeTaskTitleInvalidPrefix},
		{"タスク名がスペースのみ", "Task 1:    ", 0, 0, "", true, ErrCodeTaskTitleInvalidPrefix},

		// 異常系: タスク名に改行
		{"タスク名に改行（\\n）", "Task 1: タスク\n名", 0, 0, "", true, ErrCodeTaskTitleInvalidPrefix},
		{"タスク名に改行（末尾）", "Task 1: タスク名\n", 0, 0, "", true, ErrCodeTaskTitleInvalidPrefix},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			num, name, err := ValidateTaskTitleFormat(tt.title, tt.index)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTaskTitleFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				var ve *ValidationError
				if !errors.As(err, &ve) {
					t.Errorf("Expected ValidationError, got %T", err)
					return
				}
				if ve.Code() != tt.wantErrCode {
					t.Errorf("ValidationError.Code() = %v, want %v", ve.Code(), tt.wantErrCode)
				}
				// エラーメッセージを確認（デバッグ用）
				t.Logf("Error message: %s", ve.Message())
			} else {
				if num != tt.wantNumber {
					t.Errorf("task number = %d, want %d", num, tt.wantNumber)
				}
				if name != tt.wantName {
					t.Errorf("task name = %s, want %s", name, tt.wantName)
				}
			}
		})
	}
}

func TestValidateTaskNumberSequence(t *testing.T) {
	tests := []struct {
		name        string
		tasks       []Task
		wantErr     bool
		wantErrCode ValidationErrorCode
	}{
		// 正常系
		{
			name: "1タスクで番号1",
			tasks: []Task{
				{ID: "task-1", Title: "Task 1: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
			},
			wantErr: false,
		},
		{
			name: "3タスクで連続番号（1,2,3）",
			tasks: []Task{
				{ID: "task-1", Title: "Task 1: 最初", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
				{ID: "task-2", Title: "Task 2: 次", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
				{ID: "task-3", Title: "Task 3: 最後", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
			},
			wantErr: false,
		},
		{
			name: "順番が入れ替わっていてもOK（3,1,2）",
			tasks: []Task{
				{ID: "task-3", Title: "Task 3: 3番目", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
				{ID: "task-1", Title: "Task 1: 1番目", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
				{ID: "task-2", Title: "Task 2: 2番目", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
			},
			wantErr: false,
		},

		// 異常系: タイトル形式エラー（ValidateTaskTitleFormatで検出されるべきエラー）
		{
			name: "形式不正なタイトル",
			tasks: []Task{
				{ID: "task-1", Title: "タスク名のみ", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
			},
			wantErr:     true,
			wantErrCode: ErrCodeTaskTitleInvalidPrefix,
		},

		// 異常系: 番号0から始まる（形式エラーとして検出される）
		{
			name: "0から始まる",
			tasks: []Task{
				{ID: "task-0", Title: "Task 0: ゼロ", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
			},
			wantErr:     true,
			wantErrCode: ErrCodeTaskTitleInvalidPrefix,
		},

		// 異常系: 2から始まる（1が欠落）
		{
			name: "2から始まる（1が欠落）",
			tasks: []Task{
				{ID: "task-2", Title: "Task 2: タスク", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
			},
			wantErr:     true,
			wantErrCode: ErrCodeTaskNumberNotSequential,
		},

		// 異常系: 番号が飛ぶ（1,3）
		{
			name: "番号が飛ぶ（1,3）",
			tasks: []Task{
				{ID: "task-1", Title: "Task 1: 最初", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
				{ID: "task-3", Title: "Task 3: 次？", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
			},
			wantErr:     true,
			wantErrCode: ErrCodeTaskNumberNotSequential,
		},

		// 異常系: 番号が重複（1,1）
		{
			name: "番号が重複（1,1）",
			tasks: []Task{
				{ID: "task-1", Title: "Task 1: 最初", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
				{ID: "task-1-dup", Title: "Task 1: 重複", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
			},
			wantErr:     true,
			wantErrCode: ErrCodeTaskNumberDuplicate,
		},

		// 異常系: 番号が範囲外（1,2,10 だが3タスクしかない）
		{
			name: "番号が範囲外（1,2,10）",
			tasks: []Task{
				{ID: "task-1", Title: "Task 1: 最初", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
				{ID: "task-2", Title: "Task 2: 次", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
				{ID: "task-10", Title: "Task 10: 飛んでる", Status: TaskStatusNotStarted, Summary: []string{"要約"}, Description: "Desc"},
			},
			wantErr:     true,
			wantErrCode: ErrCodeTaskNumberNotSequential,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTaskNumberSequence(tt.tasks)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTaskNumberSequence() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				var ve *ValidationError
				if !errors.As(err, &ve) {
					t.Errorf("Expected ValidationError, got %T", err)
					return
				}
				if ve.Code() != tt.wantErrCode {
					t.Errorf("ValidationError.Code() = %v, want %v", ve.Code(), tt.wantErrCode)
				}
				t.Logf("Error message: %s", ve.Message())
			}
		})
	}
}
