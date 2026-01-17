package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateConfigFile(t *testing.T) {
	tests := []struct {
		name      string
		setupFile func(t *testing.T) string
		wantErr   bool
		errMsg    string
	}{
		{
			name: "有効な設定ファイル",
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yaml")
				if err := os.WriteFile(configPath, []byte("test"), 0600); err != nil {
					t.Fatalf("Failed to write config file: %v", err)
				}
				return configPath
			},
			wantErr: false,
		},
		{
			name: "ファイルがgroup-writable",
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yaml")
				if err := os.WriteFile(configPath, []byte("test"), 0600); err != nil {
					t.Fatalf("Failed to write config file: %v", err)
				}
				if err := os.Chmod(configPath, 0620); err != nil {
					t.Fatalf("Failed to chmod config file: %v", err)
				}
				return configPath
			},
			wantErr: true,
			errMsg:  "group or world writable",
		},
		{
			name: "ファイルがworld-writable",
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yaml")
				if err := os.WriteFile(configPath, []byte("test"), 0600); err != nil {
					t.Fatalf("Failed to write config file: %v", err)
				}
				if err := os.Chmod(configPath, 0602); err != nil {
					t.Fatalf("Failed to chmod config file: %v", err)
				}
				return configPath
			},
			wantErr: true,
			errMsg:  "group or world writable",
		},
		{
			name: "ディレクトリがgroup-writable",
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				if err := os.Chmod(tmpDir, 0720); err != nil {
					t.Fatalf("Failed to chmod directory: %v", err)
				}
				configPath := filepath.Join(tmpDir, "config.yaml")
				if err := os.WriteFile(configPath, []byte("test"), 0600); err != nil {
					t.Fatalf("Failed to write config file: %v", err)
				}
				return configPath
			},
			wantErr: true,
			errMsg:  "directory is group or world writable",
		},
		{
			name: "ファイルが存在しない",
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				return filepath.Join(tmpDir, "nonexistent.yaml")
			},
			wantErr: true,
			errMsg:  "failed to stat config file",
		},
		{
			name: "ディレクトリを指定",
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				return tmpDir
			},
			wantErr: true,
			errMsg:  "not a regular file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setupFile(t)
			err := ValidateConfigFile(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfigFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateConfigFile() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "有効な設定",
			config: &Config{
				Esa: struct {
					TeamName string `yaml:"team_name"`
				}{
					TeamName: "my-team",
				},
				AllowedCategories: []string{"LLM/Tasks"},
			},
			wantErr: false,
		},
		{
			name: "team_nameが空",
			config: &Config{
				Esa: struct {
					TeamName string `yaml:"team_name"`
				}{
					TeamName: "",
				},
				AllowedCategories: []string{"LLM/Tasks"},
			},
			wantErr: true,
			errMsg:  "team_name cannot be empty",
		},
		{
			name: "team_nameに不正な文字",
			config: &Config{
				Esa: struct {
					TeamName string `yaml:"team_name"`
				}{
					TeamName: "my/team",
				},
				AllowedCategories: []string{"LLM/Tasks"},
			},
			wantErr: true,
			errMsg:  "team_name contains invalid characters",
		},
		{
			name: "allowed_categoriesが空",
			config: &Config{
				Esa: struct {
					TeamName string `yaml:"team_name"`
				}{
					TeamName: "my-team",
				},
				AllowedCategories: []string{},
			},
			wantErr: true,
			errMsg:  "allowed_categories cannot be empty",
		},
		{
			name: "allowed_categoriesに不正なカテゴリ",
			config: &Config{
				Esa: struct {
					TeamName string `yaml:"team_name"`
				}{
					TeamName: "my-team",
				},
				AllowedCategories: []string{"LLM/../Tasks"},
			},
			wantErr: true,
			errMsg:  "invalid allowed category",
		},
		{
			name: "複数の有効なカテゴリ",
			config: &Config{
				Esa: struct {
					TeamName string `yaml:"team_name"`
				}{
					TeamName: "my-team",
				},
				AllowedCategories: []string{"LLM/Tasks", "Draft/AI-Generated"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateConfig() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}
