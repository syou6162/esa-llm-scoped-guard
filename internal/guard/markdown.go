package guard

// GenerateMarkdown はBody構造体からマークダウンを生成します
func GenerateMarkdown(body *Body) string {
	return "## 背景\n\n" + body.Background
}
