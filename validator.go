package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

//go:embed schema/post.schema.json
var schemaJSON string

var compiledSchema *jsonschema.Schema

func init() {
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("post.schema.json", strings.NewReader(schemaJSON)); err != nil {
		panic(fmt.Sprintf("failed to add schema resource: %v", err))
	}
	var err error
	compiledSchema, err = compiler.Compile("post.schema.json")
	if err != nil {
		panic(fmt.Sprintf("failed to compile schema: %v", err))
	}
}

// ValidatePostInputSchema はJSONスキーマに基づいて検証します
func ValidatePostInputSchema(input *PostInput) error {
	// PostInputをJSONに変換してスキーマ検証
	data, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("failed to marshal input: %w", err)
	}

	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("failed to unmarshal input: %w", err)
	}

	if err := compiledSchema.Validate(v); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	return nil
}

// TrimPostInput はPostInputの各フィールドをトリミングします
func TrimPostInput(input *PostInput) {
	input.Name = strings.TrimSpace(input.Name)
	input.Category = strings.TrimSpace(input.Category)
	input.BodyMD = strings.TrimSpace(input.BodyMD)
	for i := range input.Tags {
		input.Tags[i] = strings.TrimSpace(input.Tags[i])
	}
}

// ValidatePostInput は PostInput の各フィールドを検証します
func ValidatePostInput(input *PostInput) error {
	// post_numberの検証
	if input.PostNumber != nil && *input.PostNumber <= 0 {
		return fmt.Errorf("post_number must be greater than 0")
	}

	// nameの検証
	if input.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if len(input.Name) > 255 {
		return fmt.Errorf("name exceeds 255 bytes")
	}
	if containsControlCharacters(input.Name) {
		return fmt.Errorf("name contains control characters")
	}

	// categoryの検証
	if input.Category == "" {
		return fmt.Errorf("category cannot be empty")
	}

	// body_mdの検証
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
	for _, tag := range input.Tags {
		if tag == "" {
			return fmt.Errorf("tag cannot be empty")
		}
		if len(tag) > 50 {
			return fmt.Errorf("tag exceeds 50 bytes")
		}
		if containsControlCharacters(tag) {
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
