package main

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// FrontmatterData はフロントマターのYAML構造体
type FrontmatterData struct {
	Name     string   `yaml:"name"`
	Category string   `yaml:"category"`
	Tags     []string `yaml:"tags,omitempty"`
	WIP      bool     `yaml:"wip"`
}

// GenerateFrontmatter はPostInputからフロントマター付きMarkdownを生成します
func GenerateFrontmatter(input *PostInput) (string, error) {
	// フロントマターデータを構築
	data := FrontmatterData{
		Name:     input.Name,
		Category: input.Category,
		Tags:     input.Tags,
		WIP:      false, // 常にfalse（Ship It!状態）
	}

	// YAMLにマーシャル
	yamlBytes, err := yaml.Marshal(&data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal frontmatter: %w", err)
	}

	// フロントマター付きMarkdownを生成
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.Write(yamlBytes)
	sb.WriteString("---\n")
	sb.WriteString(input.BodyMD)

	result := sb.String()

	// 最終ペイロードサイズチェック（1MB制限）
	if len(result) > 1024*1024 {
		return "", fmt.Errorf("final payload size exceeds 1MB (frontmatter + body_md)")
	}

	return result, nil
}
