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
	// 行単位の差分を生成
	dmp := diffmatchpatch.New()
	a, b, lineArray := dmp.DiffLinesToChars(oldText, newText)
	diffs := dmp.DiffMain(a, b, false)
	diffs = dmp.DiffCharsToLines(diffs, lineArray)
	diffs = dmp.DiffCleanupSemantic(diffs)

	// 差分がない場合は空文字列を返す
	hasChanges := false
	for _, diff := range diffs {
		if diff.Type != diffmatchpatch.DiffEqual {
			hasChanges = true
			break
		}
	}
	if !hasChanges {
		return ""
	}

	// unified diff形式で出力
	now := time.Now().Format("2006-01-02 15:04:05 -0700")
	var result strings.Builder
	result.WriteString("--- old\t" + now + "\n")
	result.WriteString("+++ new\t" + now + "\n")

	const contextLines = 3
	oldLineNum := 1
	newLineNum := 1
	var hunkLines []string
	var hunkOldStart, hunkNewStart int
	var hunkOldCount, hunkNewCount int
	contextAfter := 0

	flushHunk := func() {
		if len(hunkLines) > 0 {
			result.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n", hunkOldStart, hunkOldCount, hunkNewStart, hunkNewCount))
			for _, line := range hunkLines {
				result.WriteString(line)
			}
			hunkLines = nil
			hunkOldCount = 0
			hunkNewCount = 0
			contextAfter = 0
		}
	}

	for i, diff := range diffs {
		lines := strings.Split(diff.Text, "\n")
		// 最後の空行を削除
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}

		switch diff.Type {
		case diffmatchpatch.DiffEqual:
			for j, line := range lines {
				// ハンク開始前または変更直後のコンテキスト行
				if len(hunkLines) == 0 {
					// 新しいハンク開始: 最大contextLines行前から
					if j < len(lines)-contextLines && i < len(diffs)-1 {
						// スキップする行でも行番号を進める
						oldLineNum++
						newLineNum++
						continue
					}
					hunkOldStart = oldLineNum
					hunkNewStart = newLineNum
				} else if contextAfter >= contextLines {
					// contextLines行以上の等価行が続いたらハンクを分割
					flushHunk()
					// 次のハンクのために位置を調整
					if j < len(lines)-contextLines {
						oldLineNum++
						newLineNum++
						continue
					}
					// 末尾の場合は新規ハンクを開始しない
					if i == len(diffs)-1 {
						oldLineNum++
						newLineNum++
						continue
					}
					hunkOldStart = oldLineNum
					hunkNewStart = newLineNum
					contextAfter = 0
				}

				hunkLines = append(hunkLines, " "+line+"\n")
				hunkOldCount++
				hunkNewCount++
				oldLineNum++
				newLineNum++
				contextAfter++
			}
		case diffmatchpatch.DiffDelete:
			if len(hunkLines) == 0 {
				// 新しいハンク開始
				hunkOldStart = oldLineNum
				hunkNewStart = newLineNum
			}
			contextAfter = 0
			for _, line := range lines {
				hunkLines = append(hunkLines, "-"+line+"\n")
				hunkOldCount++
				oldLineNum++
			}
		case diffmatchpatch.DiffInsert:
			if len(hunkLines) == 0 {
				// 新しいハンク開始
				hunkOldStart = oldLineNum
				hunkNewStart = newLineNum
			}
			contextAfter = 0
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
