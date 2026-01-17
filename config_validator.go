package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/syou6162/esa-llm-scoped-guard/internal/guard"
)

// ValidateConfigFile は設定ファイルのセキュリティ検証を行います
func ValidateConfigFile(path string) error {
	// ファイル情報を取得
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat config file: %w", err)
	}

	// 通常ファイルかチェック
	if !fileInfo.Mode().IsRegular() {
		return fmt.Errorf("config file is not a regular file: %s", path)
	}

	// ファイル権限をチェック（group/world-writableでないこと）
	if fileInfo.Mode().Perm()&0022 != 0 {
		return fmt.Errorf("config file is group or world writable: %s", path)
	}

	// ファイルの所有者をチェック（現在のユーザーが所有していること）
	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("failed to get file ownership info")
	}
	if stat.Uid != uint32(os.Getuid()) {
		return fmt.Errorf("config file is not owned by current user")
	}

	// 設定ディレクトリの権限もチェック
	dirInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		return fmt.Errorf("failed to stat config directory: %w", err)
	}
	if dirInfo.Mode().Perm()&0022 != 0 {
		return fmt.Errorf("config directory is group or world writable")
	}

	return nil
}

// ValidateConfig は設定の妥当性を検証します
func ValidateConfig(config *Config) error {
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
