package guard

import (
	"fmt"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/syou6162/esa-llm-scoped-guard/internal/esa"
)

// ExecuteDiff は既存記事との差分を標準出力に出力する。
func ExecuteDiff(jsonPath string, teamName string, accessToken string) error {
	client := esa.NewEsaClient(teamName, accessToken)
	return executeDiffWithClient(jsonPath, client)
}

func executeDiffWithClient(jsonPath string, client esa.EsaClientInterface) error {
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

	newMarkdown := GenerateMarkdown(&input.Body)

	diff := generateUnifiedDiff(existingPost.BodyMD, newMarkdown)
	fmt.Print(diff)

	return nil
}

func generateUnifiedDiff(oldText, newText string) string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(oldText, newText, false)
	return dmp.DiffPrettyText(diffs)
}
