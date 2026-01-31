package guard

import (
	"fmt"
)

// ExecuteValidate はYAMLの妥当性を検証する。
// 正常時は何も出力せず終了コード0を返す。
func ExecuteValidate(yamlPath string) error {
	input, err := ReadPostInputFromFile(yamlPath)
	if err != nil {
		return fmt.Errorf("failed to read YAML file: %w", err)
	}

	TrimPostInput(input)

	if err := ValidatePostInput(input); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}
