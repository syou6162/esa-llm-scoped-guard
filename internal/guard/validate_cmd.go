package guard

import (
	"fmt"
)

// ExecuteValidate はJSONの妥当性を検証する。
// 正常時は何も出力せず終了コード0を返す。
func ExecuteValidate(jsonPath string) error {
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

	return nil
}
