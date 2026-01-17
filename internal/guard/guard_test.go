package guard

import (
	"testing"
)

func TestIsAllowedCategory(t *testing.T) {
	tests := []struct {
		name              string
		allowedCategories []string
		category          string
		want              bool
		wantErr           bool
	}{
		{
			name:              "完全一致",
			allowedCategories: []string{"LLM/Tasks"},
			category:          "LLM/Tasks",
			want:              true,
		},
		{
			name:              "サブカテゴリ",
			allowedCategories: []string{"LLM/Tasks"},
			category:          "LLM/Tasks/2025-01",
			want:              true,
		},
		{
			name:              "複数の許可カテゴリ（最初にマッチ）",
			allowedCategories: []string{"LLM/Tasks", "Draft/AI-Generated"},
			category:          "LLM/Tasks/2025-01",
			want:              true,
		},
		{
			name:              "複数の許可カテゴリ（2番目にマッチ）",
			allowedCategories: []string{"LLM/Tasks", "Draft/AI-Generated"},
			category:          "Draft/AI-Generated/Test",
			want:              true,
		},
		{
			name:              "許可外カテゴリ",
			allowedCategories: []string{"LLM/Tasks"},
			category:          "LLM/Other",
			want:              false,
		},
		{
			name:              "境界チェック: Tasks-evilはTasksにマッチしない",
			allowedCategories: []string{"LLM/Tasks"},
			category:          "LLM/Tasks-evil",
			want:              false,
		},
		{
			name:              "境界チェック: TasksがTasks/subにマッチしない",
			allowedCategories: []string{"LLM/Tasks/sub"},
			category:          "LLM/Tasks",
			want:              false,
		},
		{
			name:              "空カテゴリ",
			allowedCategories: []string{"LLM/Tasks"},
			category:          "",
			want:              false,
			wantErr:           true,
		},
		{
			name:              "..を含むカテゴリ",
			allowedCategories: []string{"LLM/Tasks"},
			category:          "LLM/../Tasks",
			want:              false,
			wantErr:           true,
		},
		{
			name:              ".を含むカテゴリ",
			allowedCategories: []string{"LLM/Tasks"},
			category:          "LLM/./Tasks",
			want:              false,
			wantErr:           true,
		},
		{
			name:              "末尾スラッシュ",
			allowedCategories: []string{"LLM/Tasks"},
			category:          "LLM/Tasks/",
			want:              false,
			wantErr:           true,
		},
		{
			name:              "先頭スラッシュ",
			allowedCategories: []string{"LLM/Tasks"},
			category:          "/LLM/Tasks",
			want:              false,
			wantErr:           true,
		},
		{
			name:              "連続スラッシュ",
			allowedCategories: []string{"LLM/Tasks"},
			category:          "LLM//Tasks",
			want:              false,
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsAllowedCategory(tt.category, tt.allowedCategories)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsAllowedCategory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsAllowedCategory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeCategory(t *testing.T) {
	tests := []struct {
		name     string
		category string
		want     string
		wantErr  bool
	}{
		{
			name:     "通常のカテゴリ",
			category: "LLM/Tasks/2025-01",
			want:     "LLM/Tasks/2025-01",
		},
		{
			name:     "ASCII文字のみ",
			category: "ABC/def-123_456",
			want:     "ABC/def-123_456",
		},
		{
			name:     "空文字列",
			category: "",
			wantErr:  true,
		},
		{
			name:     "..を含む",
			category: "LLM/../Tasks",
			wantErr:  true,
		},
		{
			name:     ".を含む",
			category: "LLM/./Tasks",
			wantErr:  true,
		},
		{
			name:     "末尾スラッシュ",
			category: "LLM/Tasks/",
			wantErr:  true,
		},
		{
			name:     "先頭スラッシュ",
			category: "/LLM/Tasks",
			wantErr:  true,
		},
		{
			name:     "連続スラッシュ",
			category: "LLM//Tasks",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeCategory(tt.category)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeCategory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("NormalizeCategory() = %v, want %v", got, tt.want)
			}
		})
	}
}
