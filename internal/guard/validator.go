package guard

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/url"
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
	input.Body.Background = strings.TrimSpace(input.Body.Background)

	// Tasksの各フィールドをトリミング
	for i := range input.Body.Tasks {
		input.Body.Tasks[i].ID = strings.TrimSpace(input.Body.Tasks[i].ID)
		input.Body.Tasks[i].Title = strings.TrimSpace(input.Body.Tasks[i].Title)
		input.Body.Tasks[i].Description = strings.TrimSpace(input.Body.Tasks[i].Description)
		for j := range input.Body.Tasks[i].GitHubURLs {
			input.Body.Tasks[i].GitHubURLs[j] = strings.TrimSpace(input.Body.Tasks[i].GitHubURLs[j])
		}
		for j := range input.Body.Tasks[i].DependsOn {
			input.Body.Tasks[i].DependsOn[j] = strings.TrimSpace(input.Body.Tasks[i].DependsOn[j])
		}
	}
}

// ValidatePostInput は PostInput の各フィールドを検証します
func ValidatePostInput(input *PostInput) error {
	// create_newとpost_numberの検証
	if input.CreateNew && input.PostNumber != nil {
		return fmt.Errorf("cannot specify both create_new and post_number")
	}
	if !input.CreateNew && input.PostNumber == nil {
		return fmt.Errorf("must specify either create_new or post_number")
	}
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

	// bodyの検証
	if input.Body.Background == "" {
		return fmt.Errorf("background cannot be empty")
	}
	// backgroundには## 背景より上位の見出し（#, ##）を含めることができない
	if containsHeadingMarkers(input.Body.Background, 2) {
		return fmt.Errorf("background cannot contain heading markers (# or ##)")
	}

	// tasksの検証
	if len(input.Body.Tasks) == 0 {
		return fmt.Errorf("tasks cannot be empty")
	}

	// タスクIDのユニーク性チェック用マップ
	taskIDs := make(map[string]bool)

	for i, task := range input.Body.Tasks {
		if task.ID == "" {
			return fmt.Errorf("task[%d].id cannot be empty", i)
		}
		if task.Title == "" {
			return fmt.Errorf("task[%d].title cannot be empty", i)
		}
		if task.Description == "" {
			return fmt.Errorf("task[%d].description cannot be empty", i)
		}
		// descriptionには### タスクタイトルより上位の見出し（#, ##, ###）を含めることができない
		if containsHeadingMarkers(task.Description, 3) {
			return fmt.Errorf("task[%d].description cannot contain heading markers (# or ## or ###)", i)
		}
		if string(task.Status) == "" {
			return fmt.Errorf("task[%d].status cannot be empty", i)
		}

		// GitHub URLsの検証
		for j, ghURL := range task.GitHubURLs {
			if !isGitHubURL(ghURL) {
				return fmt.Errorf("task[%d].github_urls[%d]: must be a valid GitHub URL (https://github.com/...)", i, j)
			}
		}

		// IDのユニーク性チェック
		if taskIDs[task.ID] {
			return fmt.Errorf("duplicate task ID: %s", task.ID)
		}
		taskIDs[task.ID] = true
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

// isGitHubURL はURLがgithub.comドメインかつHTTPSかを検証します
func isGitHubURL(urlStr string) bool {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	return parsed.Scheme == "https" && parsed.Host == "github.com"
}

// containsHeadingMarkers は文字列に行頭の見出しマーカー（#で始まる行）が含まれているかチェックします
// maxLevel: 許可しない最大レベル（1=#, 2=##, 3=###）
func containsHeadingMarkers(text string, maxLevel int) bool {
	// 行頭（または行頭の空白の後）に1~maxLevel個の#が続く行を検出
	pattern := fmt.Sprintf(`(?m)^\s*#{1,%d}\s`, maxLevel)
	re := regexp.MustCompile(pattern)
	return re.MatchString(text)
}
