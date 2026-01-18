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

// generateSummarySection はサマリーセクションを生成します
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

	var sb strings.Builder
	sb.WriteString("\n\n## タスク\n")
	for _, task := range tasks {
		sb.WriteString(generateTaskMarkdown(task, taskTitles))
	}
	return sb.String()
}

// generateTaskMarkdown は1つのタスクのマークダウンを生成します
func generateTaskMarkdown(task Task, taskTitles map[string]string) string {
	var sb strings.Builder

	sb.WriteString("\n### ")
	sb.WriteString(task.Title)
	sb.WriteString("\n")
	sb.WriteString("- Status: `")
	sb.WriteString(string(task.Status))
	sb.WriteString("`\n")

	// 依存関係の表示
	if len(task.DependsOn) > 0 {
		sb.WriteString("- Depends on:\n")
		for _, depID := range task.DependsOn {
			sb.WriteString("  - ")
			sb.WriteString(taskTitles[depID])
			sb.WriteString("\n")
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
