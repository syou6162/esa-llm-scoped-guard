package guard

import (
	"fmt"
	"strings"
)

// NormalizeCategory はカテゴリを正規化し、検証します。
// パストラバーサル（..）、空セグメント、先頭/末尾スラッシュを拒否します。
func NormalizeCategory(category string) (string, error) {
	if category == "" {
		return "", fmt.Errorf("category cannot be empty")
	}

	// 先頭/末尾スラッシュチェック
	if strings.HasPrefix(category, "/") {
		return "", fmt.Errorf("category cannot start with /: %s", category)
	}
	if strings.HasSuffix(category, "/") {
		return "", fmt.Errorf("category cannot end with /: %s", category)
	}

	// パスを/で分割
	segments := strings.Split(category, "/")

	// 各セグメントを検証
	for _, seg := range segments {
		if seg == "" {
			return "", fmt.Errorf("category contains empty segment: %s", category)
		}
		if seg == "." || seg == ".." {
			return "", fmt.Errorf("category contains . or ..: %s", category)
		}
	}

	// 正規化後のカテゴリを返す（現状は入力と同じ）
	return category, nil
}

// IsAllowedCategory はカテゴリが許可カテゴリに含まれているかチェックします。
// 境界チェックにより、"LLM/Tasks" が "LLM/Tasks-evil" にマッチしないことを保証します。
func IsAllowedCategory(category string, allowedCategories []string) (bool, error) {
	// カテゴリを正規化
	normalized, err := NormalizeCategory(category)
	if err != nil {
		return false, err
	}

	// 許可カテゴリと比較
	for _, allowed := range allowedCategories {
		// 許可カテゴリも正規化（設定ファイルから読み込まれた値も検証）
		normalizedAllowed, err := NormalizeCategory(allowed)
		if err != nil {
			return false, fmt.Errorf("invalid allowed category %s: %w", allowed, err)
		}

		// 完全一致
		if normalized == normalizedAllowed {
			return true, nil
		}

		// サブカテゴリチェック（境界チェック付き）
		// "LLM/Tasks" が "LLM/Tasks/sub" にマッチするが、"LLM/Tasks-evil" にはマッチしない
		if strings.HasPrefix(normalized, normalizedAllowed+"/") {
			return true, nil
		}
	}

	return false, nil
}
