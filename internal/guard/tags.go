package guard

// MergeTags は既存のタグリストに新しいタグを追加します（重複しない場合のみ）
func MergeTags(existingTags []string, newTag string) []string {
	if newTag == "" {
		return existingTags
	}

	// 既に存在するかチェック
	for _, tag := range existingTags {
		if tag == newTag {
			return existingTags
		}
	}

	// 存在しない場合は追加
	return append(existingTags, newTag)
}
