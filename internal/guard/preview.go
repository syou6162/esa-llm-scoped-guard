package guard

import (
	"fmt"
)

// ExecutePreview は生成されるMarkdownを標準出力に出力する。
func ExecutePreview(yamlPath string) error {
	input, err := ReadPostInputFromFile(yamlPath)
	if err != nil {
		return fmt.Errorf("failed to read YAML file: %w", err)
	}

	TrimPostInput(input)

	if err := ValidatePostInput(input); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Generate markdown with embedded YAML (same format as post)
	markdown, err := GenerateMarkdownWithYAML(input)
	if err != nil {
		return fmt.Errorf("failed to generate markdown: %w", err)
	}
	fmt.Print(markdown)

	return nil
}
