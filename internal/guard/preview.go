package guard

import (
	"fmt"
)

// ExecutePreview は生成されるMarkdownを標準出力に出力する。
func ExecutePreview(jsonPath string) error {
	input, err := ReadPostInputFromFile(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to read JSON file: %w", err)
	}

	TrimPostInput(input)

	if err := ValidatePostInputSchema(input); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	if err := ValidatePostInput(input); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	markdown := GenerateMarkdown(&input.Body)
	fmt.Print(markdown)

	return nil
}
