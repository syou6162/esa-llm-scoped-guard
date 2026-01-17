package guard

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

//go:embed schema/post.schema.json
var schemaJSON string

var (
	compiledSchema     *jsonschema.Schema
	schemaCompileError error
	schemaOnce         sync.Once
)

// compileSchema はJSONスキーマを一度だけコンパイルします
func compileSchema() {
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("post.schema.json", strings.NewReader(schemaJSON)); err != nil {
		schemaCompileError = fmt.Errorf("failed to add schema resource: %w", err)
		return
	}
	var err error
	compiledSchema, err = compiler.Compile("post.schema.json")
	if err != nil {
		schemaCompileError = fmt.Errorf("failed to compile schema: %w", err)
		return
	}
}

// ValidatePostInputSchema はJSONスキーマに基づいて検証します
func ValidatePostInputSchema(input *PostInput) error {
	// スキーマを一度だけコンパイル
	schemaOnce.Do(compileSchema)
	if schemaCompileError != nil {
		return schemaCompileError
	}

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
	if strings.Contains(input.Name, "/") {
		return fmt.Errorf("name cannot contain /")
	}
	if strings.ContainsAny(input.Name, "（）：") {
		return fmt.Errorf("name cannot contain fullwidth parentheses or colon")
	}

	// categoryの検証
	if input.Category == "" {
		return fmt.Errorf("category cannot be empty")
	}
	if !hasValidDateSuffix(input.Category) {
		return fmt.Errorf("category must end with /yyyy/mm/dd format")
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

var dateSuffixRegex = regexp.MustCompile(`/(\d{4})/(\d{2})/(\d{2})$`)

// hasValidDateSuffix はcategoryが/yyyy/mm/dd形式で終わっているかチェックします
func hasValidDateSuffix(category string) bool {
	matches := dateSuffixRegex.FindStringSubmatch(category)
	if matches == nil {
		return false
	}

	year, _ := strconv.Atoi(matches[1])
	month, _ := strconv.Atoi(matches[2])
	day, _ := strconv.Atoi(matches[3])

	// 年: 2000-2099
	if year < 2000 || year > 2099 {
		return false
	}

	// 月: 1-12
	if month < 1 || month > 12 {
		return false
	}

	// 日: 1-31
	if day < 1 || day > 31 {
		return false
	}

	return true
}
