package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/syou6162/esa-llm-scoped-guard/internal/guard"
	"gopkg.in/yaml.v3"
)

// Config はアプリケーション設定
type Config struct {
	Esa struct {
		TeamName string `yaml:"team_name"`
	} `yaml:"esa"`
	AllowedCategories []string `yaml:"allowed_categories"`
}

// LoadConfig は設定ファイルを読み込みます
func LoadConfig(path string) (*Config, error) {
	// symlinkを解決
	realPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve symlink: %w", err)
	}

	// ファイル情報を取得
	fileInfo, err := os.Stat(realPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat config file: %w", err)
	}

	// 通常ファイルかチェック
	if !fileInfo.Mode().IsRegular() {
		return nil, fmt.Errorf("config file is not a regular file: %s", realPath)
	}

	// ファイル権限をチェック（group/world-writableでないこと）
	if fileInfo.Mode().Perm()&0022 != 0 {
		return nil, fmt.Errorf("config file is group or world writable: %s", realPath)
	}

	// ファイルの所有者をチェック（現在のユーザーが所有していること）
	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, fmt.Errorf("failed to get file ownership info")
	}
	if stat.Uid != uint32(os.Getuid()) {
		return nil, fmt.Errorf("config file is not owned by current user")
	}

	// 設定ディレクトリの権限もチェック
	dirInfo, err := os.Stat(filepath.Dir(realPath))
	if err != nil {
		return nil, fmt.Errorf("failed to stat config directory: %w", err)
	}
	if dirInfo.Mode().Perm()&0022 != 0 {
		return nil, fmt.Errorf("config directory is group or world writable")
	}

	// ファイルを読み込み
	data, err := os.ReadFile(realPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// YAMLをパース
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 設定を検証
	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// validateConfig は設定の妥当性を検証します
func validateConfig(config *Config) error {
	// team_nameの検証
	if config.Esa.TeamName == "" {
		return fmt.Errorf("team_name cannot be empty")
	}

	// team_nameは [A-Za-z0-9_-] のみ許可
	for _, c := range config.Esa.TeamName {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' || c == '-') {
			return fmt.Errorf("team_name contains invalid characters (only A-Z, a-z, 0-9, _, - allowed): %s", config.Esa.TeamName)
		}
	}

	// allowed_categoriesの検証（fail closed）
	if len(config.AllowedCategories) == 0 {
		return fmt.Errorf("allowed_categories cannot be empty (fail closed)")
	}

	// 各カテゴリを正規化・検証
	for i, category := range config.AllowedCategories {
		normalized, err := guard.NormalizeCategory(category)
		if err != nil {
			return fmt.Errorf("invalid allowed category %s: %w", category, err)
		}
		config.AllowedCategories[i] = normalized
	}

	return nil
}
