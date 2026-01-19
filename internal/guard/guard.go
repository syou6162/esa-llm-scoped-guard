package guard

import (
	"fmt"
	"strings"
)

// NormalizeCategory はカテゴリを正規化し、検証します。
// パストラバーサル（..）、空セグメント、先頭/末尾スラッシュを拒否します。
func NormalizeCategory(category string) (string, error) {
	if category == "" {
		return "", NewValidationError(ErrCodeCategoryEmpty, "category cannot be empty")
	}

	// 先頭/末尾スラッシュチェック
	if strings.HasPrefix(category, "/") {
		return "", NewValidationError(ErrCodeCategoryInvalidPath, fmt.Sprintf("category cannot start with /: %s", category)).
			WithField("category")
	}
	if strings.HasSuffix(category, "/") {
		return "", NewValidationError(ErrCodeCategoryInvalidPath, fmt.Sprintf("category cannot end with /: %s", category)).
			WithField("category")
	}

	// パスを/で分割
	segments := strings.Split(category, "/")

	// 各セグメントを検証
	for _, seg := range segments {
		if seg == "" {
			return "", NewValidationError(ErrCodeCategoryInvalidPath, fmt.Sprintf("category contains empty segment: %s", category)).
				WithField("category")
		}
		if seg == "." || seg == ".." {
			return "", NewValidationError(ErrCodeCategoryInvalidPath, fmt.Sprintf("category contains . or ..: %s", category)).
				WithField("category")
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
			return false, NewValidationError(ErrCodeCategoryInvalidPath, fmt.Sprintf("invalid allowed category %s: %v", allowed, err)).
				Wrap(err)
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

// ValidateUpdateRequest は更新リクエストの妥当性を検証します。
// 既存記事のカテゴリが許可範囲内か、カテゴリ変更が試みられていないかをチェックします。
func ValidateUpdateRequest(existingCategory, newCategory string, allowedCategories []string) error {
	// 既存カテゴリが許可範囲内か確認
	allowedExisting, err := IsAllowedCategory(existingCategory, allowedCategories)
	if err != nil {
		return NewValidationError(ErrCodeCategoryNotAllowed, fmt.Sprintf("existing category validation failed: %v", err)).
			Wrap(err)
	}
	if !allowedExisting {
		return NewValidationError(ErrCodeCategoryNotAllowed, fmt.Sprintf("existing post category %s is not allowed", existingCategory)).
			WithField("category")
	}

	// カテゴリホッピング防止（既存カテゴリ == 入力カテゴリ）
	if existingCategory != newCategory {
		return NewValidationError(ErrCodeCategoryChangeNotAllowed, fmt.Sprintf("category change is not allowed (existing: %s, new: %s)", existingCategory, newCategory)).
			WithField("category")
	}

	return nil
}
