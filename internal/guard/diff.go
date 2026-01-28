package guard

import (
	"bytes"
	"fmt"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/syou6162/esa-llm-scoped-guard/internal/esa"
)

// ExecuteDiff は既存記事との差分を標準出力に出力する。
func ExecuteDiff(jsonPath string, teamName string, allowedCategories []string, accessToken string) error {
	client := esa.NewEsaClient(teamName, accessToken)
	return executeDiffWithClient(jsonPath, allowedCategories, client)
}

func executeDiffWithClient(jsonPath string, allowedCategories []string, client esa.EsaClientInterface) error {
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

	if input.CreateNew {
		return fmt.Errorf("diff command requires post_number (cannot diff new posts)")
	}

	existingPost, err := client.GetPost(*input.PostNumber)
	if err != nil {
		return fmt.Errorf("failed to get existing post: %w", err)
	}

	// セキュリティチェック: 既存記事のカテゴリが許可範囲内か検証
	if err := ValidateUpdateRequest(existingPost.Category, input.Category, allowedCategories); err != nil {
		return fmt.Errorf("category validation failed: %w", err)
	}

	newMarkdown := GenerateMarkdown(&input.Body)

	diff := generateUnifiedDiff(existingPost.BodyMD, newMarkdown)
	fmt.Print(diff)

	return nil
}

func generateUnifiedDiff(oldText, newText string) string {
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(oldText),
		B:        difflib.SplitLines(newText),
		FromFile: "old",
		ToFile:   "new",
		Context:  3,
	}

	var buf bytes.Buffer
	if err := difflib.WriteUnifiedDiff(&buf, diff); err != nil {
		// エラー時は空文字列を返すが、実際にはWriteUnifiedDiffはほぼ失敗しない
		return ""
	}
	return buf.String()
}
