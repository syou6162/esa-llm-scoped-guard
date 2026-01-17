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
				Category: "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "## Content",
				},
			},
			wantErr: false,
		},
		{
			name: "nameが空",
			input: &PostInput{
				Name:     "",
				Category: "LLM/Tasks/2025/01/18",
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
				Name:     "Test Post",
				Category: "",
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
				Name:     "Test Post",
				Category: "LLM/Tasks",
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
			name: "有効な入力（backgroundのみ）",
			input: &PostInput{
				Name:     "Test Post",
				Category: "LLM/Tasks/2026/01/18",
				Body: Body{
					Background: "This is a background",
				},
			},
			wantErr: false,
		},
		{
			name: "backgroundが空",
			input: &PostInput{
				Name:     "Test Post",
				Category: "LLM/Tasks/2026/01/18",
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
				Name:     "Test Post",
				Category: "LLM/Tasks/2026/01/18",
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
				Name:     "Test Post",
				Category: "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "## Content",
				},
			},
			wantErr: false,
		},
		{
			name: "nameが空",
			input: &PostInput{
				Name:     "",
				Category: "LLM/Tasks/2025/01/18",
				Body: Body{
					Background: "## Content",
				},
			},
			wantErr: true,
		},
		{
			name: "backgroundが空",
			input: &PostInput{
				Name:     "Test",
				Category: "LLM/Tasks/2025/01/18",
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
