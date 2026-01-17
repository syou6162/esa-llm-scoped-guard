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
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Description: "Desc"},
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
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Description: "Desc"},
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
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Description: "Desc"},
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
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Description: "Desc"},
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
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Description: "Desc"},
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
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Description: "Desc"},
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
						{ID: "task-1", Title: "Task 1", Status: TaskStatusNotStarted, Description: "Desc"},
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
					Background: "## Content",
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1",
							Status:      TaskStatusNotStarted,
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
					Background: "## Content",
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
					Background: "## Content",
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
					Background: "## Content",
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
					Background: "## Content",
					Tasks: []Task{
						{
							ID:          "task-1",
							Title:       "Task 1",
							Status:      TaskStatusNotStarted,
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
					Background: "## Content",
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
