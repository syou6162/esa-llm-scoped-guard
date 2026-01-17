package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadAndValidateConfig(t *testing.T) {
	tests := []struct {
		name       string
		configYAML string
		wantErr    bool
		errMsg     string
		checkFunc  func(*Config) error
	}{
		{
			name: "有効な設定ファイル",
			configYAML: `esa:
  team_name: "my-team"

allowed_categories:
  - "LLM/Tasks"
  - "Draft/AI-Generated"
`,
			wantErr: false,
			checkFunc: func(c *Config) error {
				if c.Esa.TeamName != "my-team" {
					t.Errorf("TeamName = %v, want my-team", c.Esa.TeamName)
				}
				if len(c.AllowedCategories) != 2 {
					t.Errorf("len(AllowedCategories) = %v, want 2", len(c.AllowedCategories))
				}
				return nil
			},
		},
		{
			name: "allowed_categoriesが空",
			configYAML: `esa:
  team_name: "my-team"

allowed_categories: []
`,
			wantErr: true,
			errMsg:  "allowed_categories cannot be empty",
		},
		{
			name: "allowed_categoriesが存在しない",
			configYAML: `esa:
  team_name: "my-team"
`,
			wantErr: true,
			errMsg:  "allowed_categories cannot be empty",
		},
		{
			name: "team_nameが空",
			configYAML: `esa:
  team_name: ""

allowed_categories:
  - "LLM/Tasks"
`,
			wantErr: true,
			errMsg:  "team_name cannot be empty",
		},
		{
			name: "team_nameに不正な文字",
			configYAML: `esa:
  team_name: "my/team"

allowed_categories:
  - "LLM/Tasks"
`,
			wantErr: true,
			errMsg:  "team_name contains invalid characters",
		},
		{
			name: "allowed_categoriesに不正なカテゴリ",
			configYAML: `esa:
  team_name: "my-team"

allowed_categories:
  - "LLM/../Tasks"
`,
			wantErr: true,
			errMsg:  "invalid allowed category",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 一時ディレクトリを作成
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			// 設定ファイルを書き込み
			err := os.WriteFile(configPath, []byte(tt.configYAML), 0600)
			if err != nil {
				t.Fatalf("Failed to write config file: %v", err)
			}

			// 設定を読み込み
			config, err := LoadAndValidateConfig(configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadAndValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("LoadAndValidateConfig() error = %v, want error containing %q", err, tt.errMsg)
				}
				return
			}

			if tt.checkFunc != nil {
				if err := tt.checkFunc(config); err != nil {
					t.Errorf("checkFunc() error = %v", err)
				}
			}
		})
	}
}
