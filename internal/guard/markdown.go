package guard

import (
	"strings"
)

// GenerateMarkdown はBody構造体からマークダウンを生成します
func GenerateMarkdown(body *Body) string {
	var sb strings.Builder

	if summary := generateSummarySection(body.Tasks); summary != "" {
		sb.WriteString(summary)
	}

	sb.WriteString(generateBackgroundSection(body.Background, body.RelatedLinks))

	if tasks := generateTasksSection(body.Tasks); tasks != "" {
		sb.WriteString(tasks)
	}

	return sb.String()
}

// generateSummarySection はサマリーセクションを生成します（依存関係グラフを含む）
func generateSummarySection(tasks []Task) string {
	if len(tasks) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("## サマリー\n")
	for _, task := range tasks {
		if task.Status == TaskStatusCompleted {
			sb.WriteString("- [x] ")
		} else {
			sb.WriteString("- [ ] ")
		}
		sb.WriteString(task.Title)
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	// 依存関係グラフをサブセクションとして追加
	if mermaid := generateMermaidGraph(tasks); mermaid != "" {
		sb.WriteString("### 依存関係グラフ\n")
		sb.WriteString(mermaid)
	}

	return sb.String()
}

// generateBackgroundSection は背景セクションを生成します
func generateBackgroundSection(background string, relatedLinks []string) string {
	var sb strings.Builder
	sb.WriteString("## 背景\n")

	if len(relatedLinks) > 0 {
		sb.WriteString("関連リンク:\n")
		for _, link := range relatedLinks {
			sb.WriteString("- ")
			sb.WriteString(link)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("\n")
	}

	sb.WriteString(background)
	return sb.String()
}

// generateTasksSection はタスクセクションを生成します
func generateTasksSection(tasks []Task) string {
	if len(tasks) == 0 {
		return ""
	}

	// IDからタイトルへのマッピングを構築
	taskTitles := make(map[string]string)
	for _, t := range tasks {
		taskTitles[t.ID] = t.Title
	}

	// 推移的縮約を適用
	reducedDeps := transitiveReduction(tasks)

	var sb strings.Builder
	sb.WriteString("\n\n## タスク\n")
	for _, task := range tasks {
		sb.WriteString(generateTaskMarkdown(task, taskTitles, reducedDeps))
	}
	return sb.String()
}

// generateTaskMarkdown は1つのタスクのマークダウンを生成します
func generateTaskMarkdown(task Task, taskTitles map[string]string, reducedDeps map[string][]string) string {
	var sb strings.Builder

	sb.WriteString("\n### ")
	sb.WriteString(task.Title)
	sb.WriteString("\n")
	sb.WriteString("- Status: `")
	sb.WriteString(string(task.Status))
	sb.WriteString("`\n")

	// 依存関係の表示（推移的縮約済み）
	deps := reducedDeps[task.ID]
	if len(deps) > 0 {
		sb.WriteString("- Depends on:\n")
		for _, depID := range deps {
			sb.WriteString("  - `")
			sb.WriteString(taskTitles[depID])
			sb.WriteString("`\n")
		}
	}

	if len(task.GitHubURLs) > 0 {
		if len(task.GitHubURLs) == 1 {
			sb.WriteString("- Pull Request: ")
			sb.WriteString(task.GitHubURLs[0])
			sb.WriteString("\n")
		} else {
			sb.WriteString("- Pull Requests:\n")
			for _, ghURL := range task.GitHubURLs {
				sb.WriteString("  - ")
				sb.WriteString(ghURL)
				sb.WriteString("\n")
			}
		}
	}

	// 要約セクションを追加
	sb.WriteString("\n- 要約:\n")
	for _, line := range task.Summary {
		sb.WriteString("  - ")
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	// Descriptionをdetailsで囲む
	sb.WriteString("\n<details><summary>詳細を開く</summary>\n\n")
	sb.WriteString(task.Description)
	sb.WriteString("\n\n</details>\n")
	return sb.String()
}

// transitiveReduction は依存関係の推移的縮約を行います
// 例: A→B, B→C, A→C の場合、A→C は冗長なので除外し A→B のみを返します
func transitiveReduction(tasks []Task) map[string][]string {
	// タスクIDからTaskへのマップを構築
	taskMap := make(map[string]Task)
	for _, task := range tasks {
		taskMap[task.ID] = task
	}

	// 縮約された依存関係を格納
	reduced := make(map[string][]string)

	// 各タスクについて処理
	for _, task := range tasks {
		if len(task.DependsOn) == 0 {
			continue
		}

		// 直接的な依存関係のみを保持
		var directDeps []string

		for _, depID := range task.DependsOn {
			// この依存先が他の依存先経由で到達可能かチェック
			isRedundant := false

			for _, otherDepID := range task.DependsOn {
				if depID == otherDepID {
					continue
				}

				// otherDepID から depID に到達可能かDFSでチェック
				if canReach(otherDepID, depID, taskMap, make(map[string]bool)) {
					isRedundant = true
					break
				}
			}

			// 冗長でなければ直接的な依存関係として保持
			if !isRedundant {
				directDeps = append(directDeps, depID)
			}
		}

		reduced[task.ID] = directDeps
	}

	return reduced
}

// canReach は from から to に依存関係を辿って到達可能かをDFSで判定します
func canReach(from, to string, taskMap map[string]Task, visited map[string]bool) bool {
	if from == to {
		return true
	}

	if visited[from] {
		return false
	}

	visited[from] = true

	task, exists := taskMap[from]
	if !exists {
		return false
	}

	for _, depID := range task.DependsOn {
		if canReach(depID, to, taskMap, visited) {
			return true
		}
	}

	return false
}

// generateMermaidGraph はタスクの依存関係をMermaid形式で出力します
func generateMermaidGraph(tasks []Task) string {
	if len(tasks) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n```mermaid\n")
	sb.WriteString("graph TD\n")

	// ノード定義
	for _, task := range tasks {
		sb.WriteString("    ")
		sb.WriteString(task.ID)
		sb.WriteString("[\"`")
		sb.WriteString(escapeMermaidTitle(task.Title))
		sb.WriteString("`\"]:::")
		sb.WriteString(string(task.Status))
		sb.WriteString("\n")
	}

	// タスク完了ノード
	sb.WriteString("    done([タスク完了]):::goal\n")

	// エッジ定義
	sb.WriteString("\n")

	// 推移的縮約を適用
	reducedDeps := transitiveReduction(tasks)

	// 依存関係のエッジを作成（縮約済み）
	for _, task := range tasks {
		deps := reducedDeps[task.ID]
		for _, depID := range deps {
			sb.WriteString("    ")
			sb.WriteString(depID)
			sb.WriteString(" --> ")
			sb.WriteString(task.ID)
			sb.WriteString("\n")
		}
	}

	// リーフノード（他から依存されていないタスク）から「タスク完了」へのエッジ
	leafNodes := findLeafNodes(tasks)
	for _, leafID := range leafNodes {
		sb.WriteString("    ")
		sb.WriteString(leafID)
		sb.WriteString(" --> done\n")
	}

	// スタイル定義
	sb.WriteString("\n")
	sb.WriteString("    classDef completed fill:#90EE90\n")
	sb.WriteString("    classDef in_progress fill:#FFD700\n")
	sb.WriteString("    classDef in_review fill:#FFA500\n")
	sb.WriteString("    classDef not_started fill:#D3D3D3\n")
	sb.WriteString("    classDef goal fill:#87CEEB,stroke:#4169E1,stroke-width:3px\n")

	sb.WriteString("```\n")
	return sb.String()
}

// escapeMermaidTitle はMermaidのノードラベル用にタイトルをエスケープします
func escapeMermaidTitle(title string) string {
	// バッククォートをエスケープ
	return strings.ReplaceAll(title, "`", "&#96;")
}

// findLeafNodes は他のタスクから依存されていないタスク（リーフノード）を見つけます
func findLeafNodes(tasks []Task) []string {
	// 全タスクIDのセットを作成
	allIDs := make(map[string]bool)
	for _, task := range tasks {
		allIDs[task.ID] = true
	}

	// 依存先として参照されているIDを除外
	for _, task := range tasks {
		for _, depID := range task.DependsOn {
			delete(allIDs, depID)
		}
	}

	// 残ったIDがリーフノード
	leafNodes := make([]string, 0, len(allIDs))
	for _, task := range tasks {
		if allIDs[task.ID] {
			leafNodes = append(leafNodes, task.ID)
		}
	}

	return leafNodes
}
