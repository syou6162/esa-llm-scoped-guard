package guard

import (
	"fmt"
	"strings"
	"time"

	"github.com/sergi/go-diff/diffmatchpatch"
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

	// サイズチェック: 既存記事の本文が大きすぎる場合は拒否（DoS対策）
	const maxBodySize = 10 * 1024 * 1024 // 10MB
	if len(existingPost.BodyMD) > maxBodySize {
		return fmt.Errorf("existing post body too large (%d bytes, max %d bytes)", len(existingPost.BodyMD), maxBodySize)
	}

	// セキュリティチェック: 既存記事のカテゴリが許可範囲内か検証
	if err := ValidateUpdateRequest(existingPost.Category, input.Category, allowedCategories); err != nil {
		return fmt.Errorf("category validation failed: %w", err)
	}

	newMarkdown := GenerateMarkdown(&input.Body)

	// サイズチェック: 新しいMarkdownが大きすぎる場合は拒否（DoS対策）
	if len(newMarkdown) > maxBodySize {
		return fmt.Errorf("new markdown too large (%d bytes, max %d bytes)", len(newMarkdown), maxBodySize)
	}

	diff := generateUnifiedDiff(existingPost.BodyMD, newMarkdown)
	fmt.Print(diff)

	return nil
}

func generateUnifiedDiff(oldText, newText string) string {
	// 標準的なunified diff形式を生成
	now := time.Now().Format("2006-01-02 15:04:05 -0700")

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(oldText, newText, false)
	diffs = dmp.DiffCleanupSemantic(diffs)

	var result strings.Builder
	result.WriteString("--- old\t" + now + "\n")
	result.WriteString("+++ new\t" + now + "\n")

	oldLineNum := 1
	newLineNum := 1
	hunkOldStart := 1
	hunkNewStart := 1
	hunkOldCount := 0
	hunkNewCount := 0
	var hunkLines []string

	flushHunk := func() {
		if len(hunkLines) > 0 {
			result.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n", hunkOldStart, hunkOldCount, hunkNewStart, hunkNewCount))
			for _, line := range hunkLines {
				result.WriteString(line)
			}
			hunkLines = nil
		}
	}

	for _, diff := range diffs {
		lines := strings.Split(diff.Text, "\n")
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}

		switch diff.Type {
		case diffmatchpatch.DiffEqual:
			for i, line := range lines {
				if i == 0 && len(hunkLines) == 0 {
					hunkOldStart = oldLineNum
					hunkNewStart = newLineNum
				}
				hunkLines = append(hunkLines, " "+line+"\n")
				hunkOldCount++
				hunkNewCount++
				oldLineNum++
				newLineNum++
			}
		case diffmatchpatch.DiffDelete:
			if len(hunkLines) == 0 {
				hunkOldStart = oldLineNum
				hunkNewStart = newLineNum
			}
			for _, line := range lines {
				hunkLines = append(hunkLines, "-"+line+"\n")
				hunkOldCount++
				oldLineNum++
			}
		case diffmatchpatch.DiffInsert:
			if len(hunkLines) == 0 {
				hunkOldStart = oldLineNum
				hunkNewStart = newLineNum
			}
			for _, line := range lines {
				hunkLines = append(hunkLines, "+"+line+"\n")
				hunkNewCount++
				newLineNum++
			}
		}
	}

	flushHunk()
	return result.String()
}
