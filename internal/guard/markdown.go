package guard

import (
	"strings"
)

// GenerateMarkdown はBody構造体からマークダウンを生成します
func GenerateMarkdown(body *Body) string {
	var sb strings.Builder

	// 背景セクション
	sb.WriteString("## 背景\n")

	// 関連リンク（存在する場合のみ、背景の最初に配置）
	if len(body.RelatedLinks) > 0 {
		sb.WriteString("関連リンク:\n")
		for _, link := range body.RelatedLinks {
			sb.WriteString("- ")
			sb.WriteString(link)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	} else {
		// 関連リンクがない場合は空行を追加
		sb.WriteString("\n")
	}

	sb.WriteString(body.Background)

	// タスクセクション
	if len(body.Tasks) > 0 {
		sb.WriteString("\n\n## タスク\n")
		for _, task := range body.Tasks {
			sb.WriteString("\n### ")
			sb.WriteString(task.ID)
			sb.WriteString(": ")
			sb.WriteString(task.Title)
			sb.WriteString("\n")
			sb.WriteString("Status: ")
			sb.WriteString(string(task.Status))
			sb.WriteString("\n\n")
			sb.WriteString(task.Description)

			// GitHub URLsセクション（存在する場合のみ）
			if len(task.GitHubURLs) > 0 {
				sb.WriteString("\n\n")
				if len(task.GitHubURLs) == 1 {
					// 単一URL
					sb.WriteString("Pull Request: ")
					sb.WriteString(task.GitHubURLs[0])
				} else {
					// 複数URL
					sb.WriteString("Pull Requests:\n")
					for _, ghURL := range task.GitHubURLs {
						sb.WriteString("- ")
						sb.WriteString(ghURL)
						sb.WriteString("\n")
					}
				}
			}
		}
	}

	return sb.String()
}
