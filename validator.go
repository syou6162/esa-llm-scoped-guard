package main

import (
	"fmt"
	"strings"
	"unicode"
)

// ValidatePostInput は PostInput の各フィールドを検証します
func ValidatePostInput(input *PostInput) error {
	// post_numberの検証
	if input.PostNumber != nil && *input.PostNumber <= 0 {
		return fmt.Errorf("post_number must be greater than 0")
	}

	// nameの検証（トリミング前）
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if len(input.Name) > 255 {
		return fmt.Errorf("name exceeds 255 bytes")
	}
	if containsControlCharacters(input.Name) {
		return fmt.Errorf("name contains control characters")
	}

	// categoryの検証（トリミング前）
	input.Category = strings.TrimSpace(input.Category)
	if input.Category == "" {
		return fmt.Errorf("category cannot be empty")
	}

	// body_mdの検証（トリミング前）
	input.BodyMD = strings.TrimSpace(input.BodyMD)
	if input.BodyMD == "" {
		return fmt.Errorf("body_md cannot be empty")
	}
	if len(input.BodyMD) > 1024*1024 {
		return fmt.Errorf("body_md exceeds 1MB")
	}
	if strings.HasPrefix(input.BodyMD, "---") {
		return fmt.Errorf("body_md cannot start with --- (frontmatter conflict)")
	}

	// tagsの検証
	if len(input.Tags) > 10 {
		return fmt.Errorf("tags cannot exceed 10")
	}
	for i, tag := range input.Tags {
		input.Tags[i] = strings.TrimSpace(tag)
		if input.Tags[i] == "" {
			return fmt.Errorf("tag cannot be empty")
		}
		if len(input.Tags[i]) > 50 {
			return fmt.Errorf("tag exceeds 50 bytes")
		}
		if containsControlCharacters(input.Tags[i]) {
			return fmt.Errorf("tag contains control characters")
		}
	}

	return nil
}

// containsControlCharacters は文字列に制御文字（改行、タブなど）が含まれているかチェックします
func containsControlCharacters(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}
