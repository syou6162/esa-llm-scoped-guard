package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config はアプリケーション設定
type Config struct {
	Esa struct {
		TeamName string `yaml:"team_name"`
	} `yaml:"esa"`
	AllowedCategories []string `yaml:"allowed_categories"`
}

// LoadAndValidateConfig は設定ファイルを読み込み、検証します
func LoadAndValidateConfig(path string) (*Config, error) {
	// symlinkを解決
	realPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve symlink: %w", err)
	}

	// ファイルのセキュリティ検証
	if err := ValidateConfigFile(realPath); err != nil {
		return nil, err
	}

	// 設定ファイルを読み込み
	config, err := loadConfig(realPath)
	if err != nil {
		return nil, err
	}

	// 設定内容を検証
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

// loadConfig は設定ファイルを読み込みます（内部用）
func loadConfig(path string) (*Config, error) {
	// ファイルを開く
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	// サイズ制限付きで読み込み（10MB+1バイト読んで超過を検出）
	const maxSize = 10 * 1024 * 1024
	limitedReader := io.LimitReader(file, maxSize+1)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// サイズ超過チェック
	if len(data) > maxSize {
		return nil, fmt.Errorf("config file size exceeds 10MB")
	}

	// YAMLをパース
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}
