package guard

import (
	_ "embed"
	"encoding/json"
	"errors"
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

// visitState はDFSでのノード訪問状態を表す
type visitState int

const (
	stateUnvisited visitState = iota // 未訪問
	stateVisiting                    // 処理中（現在のDFSパス上）
	stateVisited                     // 処理完了
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
		return NewValidationError(ErrCodeInvalidValue, fmt.Sprintf("schema validation failed: %v", err)).Wrap(err)
	}

	return nil
}

// detectCyclicDependency はDFSを使用して循環依存を検出します
func detectCyclicDependency(tasks []Task) (bool, []string) {
	// タスクIDから依存先へのマッピングを構築
	graph := make(map[string][]string)
	for _, task := range tasks {
		graph[task.ID] = task.DependsOn
	}

	state := make(map[string]visitState)
	var cyclePath []string

	var dfs func(id string, path []string) bool
	dfs = func(id string, path []string) bool {
		if state[id] == stateVisiting {
			// 処理中のノードに再訪 = 循環検出
			cyclePath = append(path, id)
			return true
		}
		if state[id] == stateVisited {
			// 既に処理完了済み
			return false
		}

		state[id] = stateVisiting
		for _, depID := range graph[id] {
			if dfs(depID, append(path, id)) {
				return true
			}
		}
		state[id] = stateVisited
		return false
	}

	for _, task := range tasks {
		if state[task.ID] == stateUnvisited {
			if dfs(task.ID, nil) {
				return true, cyclePath
			}
		}
	}

	return false, nil
}

// TrimPostInput はPostInputの各フィールドをトリミングします
func TrimPostInput(input *PostInput) {
	input.Name = strings.TrimSpace(input.Name)
	input.Category = strings.TrimSpace(input.Category)
	input.Body.Background = strings.TrimSpace(input.Body.Background)

	// Instructionsの各要素をトリミング
	for i := range input.Body.Instructions {
		input.Body.Instructions[i] = strings.TrimSpace(input.Body.Instructions[i])
	}

	// Tasksの各フィールドをトリミング
	for i := range input.Body.Tasks {
		input.Body.Tasks[i].ID = strings.TrimSpace(input.Body.Tasks[i].ID)
		input.Body.Tasks[i].Title = strings.TrimSpace(input.Body.Tasks[i].Title)
		input.Body.Tasks[i].Description = strings.TrimSpace(input.Body.Tasks[i].Description)
		for j := range input.Body.Tasks[i].Summary {
			input.Body.Tasks[i].Summary[j] = strings.TrimSpace(input.Body.Tasks[i].Summary[j])
		}
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
		return NewValidationError(ErrCodeMutuallyExclusive, "cannot specify both create_new and post_number")
	}
	if !input.CreateNew && input.PostNumber == nil {
		return NewValidationError(ErrCodeMissingRequired, "must specify either create_new or post_number")
	}
	if input.PostNumber != nil && *input.PostNumber <= 0 {
		return NewValidationError(ErrCodeInvalidValue, "post_number must be greater than 0").WithField("post_number")
	}

	// nameの検証
	if input.Name == "" {
		return NewValidationError(ErrCodeFieldEmpty, "name cannot be empty").WithField("name")
	}
	if len(input.Name) > 255 {
		return NewValidationError(ErrCodeFieldTooLong, "name exceeds 255 bytes").WithField("name")
	}
	if containsControlCharacters(input.Name) {
		return NewValidationError(ErrCodeFieldInvalidChars, "name contains control characters").WithField("name")
	}
	if strings.Contains(input.Name, "/") {
		return NewValidationError(ErrCodeFieldInvalidChars, "name cannot contain /").WithField("name")
	}
	if strings.ContainsAny(input.Name, "（）：") {
		return NewValidationError(ErrCodeFieldInvalidChars, "name cannot contain fullwidth parentheses or colon").WithField("name")
	}

	// categoryの検証
	if input.Category == "" {
		return NewValidationError(ErrCodeCategoryEmpty, "category cannot be empty").WithField("category")
	}
	if !hasValidDateSuffix(input.Category) {
		return NewValidationError(ErrCodeCategoryInvalidDateSuffix, "category must end with /yyyy/mm/dd format").WithField("category")
	}

	// bodyの検証
	if input.Body.Background == "" {
		return NewValidationError(ErrCodeFieldEmpty, "background cannot be empty").WithField("background")
	}
	// backgroundには## 背景より上位の見出し（#, ##）を含めることができない
	if containsHeadingMarkers(input.Body.Background, 2) {
		return NewValidationError(ErrCodeFieldInvalidFormat, "background cannot contain heading markers (# or ##)").WithField("background")
	}

	// instructionsの検証
	if err := ValidateInstructions(input.Body.Instructions); err != nil {
		var ve *ValidationError
		if errors.As(err, &ve) {
			// ValidationErrorの場合はコードを保持
			return NewValidationError(ve.Code(), fmt.Sprintf("instructions: %v", err)).
				WithField("instructions").Wrap(err)
		}
		// それ以外のエラーはFieldInvalidFormatとして扱う
		return NewValidationError(ErrCodeFieldInvalidFormat, fmt.Sprintf("instructions: %v", err)).
			WithField("instructions").Wrap(err)
	}

	// tasksの検証
	if len(input.Body.Tasks) == 0 {
		return NewValidationError(ErrCodeFieldEmpty, "tasks cannot be empty").WithField("tasks")
	}

	// タスクIDのユニーク性チェック用マップ
	taskIDs := make(map[string]bool)

	for i, task := range input.Body.Tasks {
		if task.ID == "" {
			return NewValidationError(ErrCodeFieldEmpty, fmt.Sprintf("task[%d].id cannot be empty", i)).
				WithField("task.id").WithIndex(i)
		}
		if task.Title == "" {
			return NewValidationError(ErrCodeFieldEmpty, fmt.Sprintf("task[%d].title cannot be empty", i)).
				WithField("task.title").WithIndex(i)
		}
		if task.Description == "" {
			return NewValidationError(ErrCodeFieldEmpty, fmt.Sprintf("task[%d].description cannot be empty", i)).
				WithField("task.description").WithIndex(i)
		}
		// descriptionには### タスクタイトルより上位の見出し（#, ##, ###）を含めることができない
		if containsHeadingMarkers(task.Description, 3) {
			return NewValidationError(ErrCodeFieldInvalidFormat, fmt.Sprintf("task[%d].description cannot contain heading markers (# or ## or ###)", i)).
				WithField("task.description").WithIndex(i)
		}
		if string(task.Status) == "" {
			return NewValidationError(ErrCodeFieldEmpty, fmt.Sprintf("task[%d].status cannot be empty", i)).
				WithField("task.status").WithIndex(i)
		}

		// Summaryの検証
		if err := ValidateSummary(task.Summary); err != nil {
			var ve *ValidationError
			if errors.As(err, &ve) {
				// ValidationErrorの場合はコードを保持してインデックスを追加
				return NewValidationError(ve.Code(), fmt.Sprintf("task[%d].summary: %v", i, err)).
					WithField("task.summary").WithIndex(i).Wrap(err)
			}
			// それ以外のエラーはFieldInvalidFormatとして扱う
			return NewValidationError(ErrCodeFieldInvalidFormat, fmt.Sprintf("task[%d].summary: %v", i, err)).
				WithField("task.summary").WithIndex(i).Wrap(err)
		}

		// GitHub URLsの検証
		for j, ghURL := range task.GitHubURLs {
			if !isGitHubURL(ghURL) {
				return NewValidationError(ErrCodeFieldInvalidFormat, fmt.Sprintf("task[%d].github_urls[%d]: must be a valid GitHub URL (https://github.com/...)", i, j)).
					WithField("task.github_urls").WithIndex(i)
			}
		}

		// ステータスとGitHub URLsの整合性チェック
		if len(task.GitHubURLs) > 0 && task.Status == TaskStatusNotStarted {
			return NewValidationError(ErrCodeFieldInvalidFormat, fmt.Sprintf("task[%d]: status is 'not_started' but has GitHub URLs (should be 'in_progress' or later)", i)).
				WithField("task.status").WithIndex(i)
		}

		// IDのユニーク性チェック
		if taskIDs[task.ID] {
			return NewValidationError(ErrCodeDuplicateID, fmt.Sprintf("duplicate task ID: %s", task.ID)).
				WithField("task.id")
		}
		taskIDs[task.ID] = true
	}

	// タスク番号の形式と連続性を検証
	if err := ValidateTaskNumberSequence(input.Body.Tasks); err != nil {
		return err
	}

	// 依存関係の検証
	for i, task := range input.Body.Tasks {
		for j, depID := range task.DependsOn {
			if depID == "" {
				return NewValidationError(ErrCodeFieldEmpty, fmt.Sprintf("task[%d].depends_on[%d]: empty task ID", i, j)).
					WithField("task.depends_on").WithIndex(i)
			}
			if depID == task.ID {
				return NewValidationError(ErrCodeSelfReference, fmt.Sprintf("task[%d].depends_on: self-reference is not allowed", i)).
					WithField("task.depends_on").WithIndex(i)
			}
			if !taskIDs[depID] {
				return NewValidationError(ErrCodeNonExistentRef, fmt.Sprintf("task[%d].depends_on references non-existent task ID: %s", i, depID)).
					WithField("task.depends_on").WithIndex(i)
			}
		}
	}

	// 循環依存チェック
	if hasCycle, cyclePath := detectCyclicDependency(input.Body.Tasks); hasCycle {
		return NewValidationError(ErrCodeCircularDependency, fmt.Sprintf("circular dependency detected: %s", strings.Join(cyclePath, " -> ")))
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

var numberedListMarkerRegex = regexp.MustCompile(`^\d+\.\s`)

// taskTitlePrefixRegex はタスクタイトルのプレフィックス形式を検証します
// 形式: "Task N: タスク名" (Nは数字、タスク名に改行を含まない)
// 先頭ゼロや0は ValidateTaskTitleFormat 内で検証
// [^\r\n]+ で改行（CR/LF）を禁止
var taskTitlePrefixRegex = regexp.MustCompile(`^Task (\d+): ([^\r\n]+)$`)

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

// ValidateSummary はSummaryフィールドを検証します
// - 最低1行、最大3行
// - 各行140字以内
func ValidateSummary(summary []string) error {
	if len(summary) < 1 || len(summary) > 3 {
		return NewValidationError(ErrCodeFieldInvalidFormat, fmt.Sprintf("summary must have 1-3 items, got %d", len(summary)))
	}
	for i, line := range summary {
		if len([]rune(line)) > 140 {
			return NewValidationError(ErrCodeFieldTooLong, fmt.Sprintf("summary line %d exceeds 140 characters", i+1))
		}
	}
	return nil
}

// ValidateInstructions はInstructionsフィールドを検証します
// - 最大10項目（0項目もOK）
// - 各項目500文字以内
// - 見出しマーカー（#, ##）禁止
// - リストマーカー（-, *, +, 数字+.）で始まる項目を禁止
func ValidateInstructions(instructions []string) error {
	if len(instructions) > 10 {
		return NewValidationError(ErrCodeFieldInvalidFormat, fmt.Sprintf("instructions must have at most 10 items, got %d", len(instructions)))
	}
	for i, item := range instructions {
		if len([]rune(item)) > 500 {
			return NewValidationError(ErrCodeFieldTooLong, fmt.Sprintf("instructions item %d exceeds 500 characters", i+1))
		}
		// 見出しマーカーチェック
		if containsHeadingMarkers(item, 2) {
			return NewValidationError(ErrCodeFieldInvalidFormat, fmt.Sprintf("instructions item %d cannot contain heading markers (# or ##)", i+1))
		}
		// リストマーカーチェック（行頭の -, *, +, 数字+. を禁止）
		trimmed := strings.TrimSpace(item)
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "+ ") {
			return NewValidationError(ErrCodeFieldInvalidFormat, fmt.Sprintf("instructions item %d cannot start with list markers (-, *, +)", i+1))
		}
		// 数字 + . のパターンチェック（例: "1. ", "2. "）
		if numberedListMarkerRegex.MatchString(trimmed) {
			return NewValidationError(ErrCodeFieldInvalidFormat, fmt.Sprintf("instructions item %d cannot start with numbered list markers (e.g., '1. ')", i+1))
		}
	}
	return nil
}

// ValidateTaskTitleFormat は単一タスクのタイトル形式を検証します
// 形式: "Task N: タスク名"（Nは1から始まる正の整数、先頭ゼロ禁止）
// 戻り値: タスク番号, タスク名, エラー
func ValidateTaskTitleFormat(title string, index int) (int, string, error) {
	matches := taskTitlePrefixRegex.FindStringSubmatch(title)
	if matches == nil {
		// 形式が不正な場合、具体的な問題を診断
		return 0, "", diagnoseTaskTitleError(title, index)
	}

	numberStr := matches[1]

	// 先頭ゼロチェック（01, 001, 010 などを弾く）
	if len(numberStr) > 1 && numberStr[0] == '0' {
		return 0, "", NewValidationError(ErrCodeTaskTitleInvalidPrefix,
			fmt.Sprintf("task[%d].title: task number cannot have leading zero (got: '%s')", index, numberStr)).
			WithField("task.title").WithIndex(index)
	}

	taskNumber, err := strconv.Atoi(numberStr)
	if err != nil {
		// 数値変換エラー（整数が大きすぎる場合など）
		return 0, "", NewValidationError(ErrCodeTaskTitleInvalidPrefix,
			fmt.Sprintf("task[%d].title: task number '%s' is too large", index, numberStr)).
			WithField("task.title").WithIndex(index)
	}

	// 番号0チェック（タスク番号は1から始まる）
	if taskNumber == 0 {
		return 0, "", NewValidationError(ErrCodeTaskTitleInvalidPrefix,
			fmt.Sprintf("task[%d].title: task number must start from 1, not 0", index)).
			WithField("task.title").WithIndex(index)
	}

	taskName := matches[2]
	if strings.TrimSpace(taskName) == "" {
		return 0, "", NewValidationError(ErrCodeTaskTitleInvalidPrefix,
			fmt.Sprintf("task[%d].title: task name cannot be empty (format: 'Task N: タスク名')", index)).
			WithField("task.title").WithIndex(index)
	}

	return taskNumber, taskName, nil
}

// diagnoseTaskTitleError はタイトル形式のエラーを診断し、具体的な修正案を提示します
func diagnoseTaskTitleError(title string, index int) error {
	// ケース1: "Task" で始まるかチェック（大文字小文字含む）
	lowerTitle := strings.ToLower(title)
	if !strings.HasPrefix(lowerTitle, "task") {
		// "Task" すら含まない場合はプレフィックスなし
		suggestion := fmt.Sprintf("Task %d: %s", index+1, title)
		return NewValidationError(ErrCodeTaskTitleInvalidPrefix,
			fmt.Sprintf("task[%d].title: must start with 'Task N: ' prefix (got: '%s', suggestion: '%s')",
				index, title, suggestion)).
			WithField("task.title").WithIndex(index)
	}

	// ケース2: "Task" の後にスペースがない (e.g., "Task1:")
	if strings.HasPrefix(title, "Task") && !strings.HasPrefix(title, "Task ") {
		return NewValidationError(ErrCodeTaskTitleInvalidPrefix,
			fmt.Sprintf("task[%d].title: must have a space after 'Task' (got: '%s')",
				index, title)).
			WithField("task.title").WithIndex(index)
	}

	// ケース3: 大文字小文字が不正 (TASK, task など)
	if !strings.HasPrefix(title, "Task ") && strings.HasPrefix(lowerTitle, "task ") {
		return NewValidationError(ErrCodeTaskTitleInvalidPrefix,
			fmt.Sprintf("task[%d].title: 'Task' must be capitalized exactly as 'Task' (got: '%s')",
				index, title)).
			WithField("task.title").WithIndex(index)
	}

	// ケース4: 番号部分の問題を診断
	afterPrefix := strings.TrimPrefix(title, "Task ")
	colonIndex := strings.Index(afterPrefix, ":")

	if colonIndex == -1 {
		// コロンがない
		return NewValidationError(ErrCodeTaskTitleInvalidPrefix,
			fmt.Sprintf("task[%d].title: missing ':' after task number (format: 'Task N: タスク名', got: '%s')",
				index, title)).
			WithField("task.title").WithIndex(index)
	}

	numberPart := strings.TrimSpace(afterPrefix[:colonIndex])

	// 番号が空
	if numberPart == "" {
		return NewValidationError(ErrCodeTaskTitleInvalidPrefix,
			fmt.Sprintf("task[%d].title: task number is missing (format: 'Task N: タスク名', got: '%s')",
				index, title)).
			WithField("task.title").WithIndex(index)
	}

	// 小数点を含む (e.g., "2.1")
	if strings.Contains(numberPart, ".") {
		return NewValidationError(ErrCodeTaskTitleInvalidPrefix,
			fmt.Sprintf("task[%d].title: task number must be integer, not decimal (got: '%s')",
				index, numberPart)).
			WithField("task.title").WithIndex(index)
	}

	// ハイフンを含む (e.g., "6-7")
	if strings.Contains(numberPart, "-") {
		return NewValidationError(ErrCodeTaskTitleInvalidPrefix,
			fmt.Sprintf("task[%d].title: task number must be single integer, not range (got: '%s')",
				index, numberPart)).
			WithField("task.title").WithIndex(index)
	}

	// アルファベットを含む (e.g., "2A", "A2")
	for _, r := range numberPart {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			return NewValidationError(ErrCodeTaskTitleInvalidPrefix,
				fmt.Sprintf("task[%d].title: task number must be integer only, no letters (got: '%s')",
					index, numberPart)).
				WithField("task.title").WithIndex(index)
		}
	}

	// 先頭ゼロ (e.g., "01", "001")
	// 番号0は連続性チェックで弾くため、先頭ゼロは1文字目が0で2文字以上の場合のみ
	if len(numberPart) > 1 && numberPart[0] == '0' {
		return NewValidationError(ErrCodeTaskTitleInvalidPrefix,
			fmt.Sprintf("task[%d].title: task number cannot have leading zero (got: '%s')",
				index, numberPart)).
			WithField("task.title").WithIndex(index)
	}

	// 番号の後にスペースがない
	if colonIndex+1 < len(afterPrefix) && afterPrefix[colonIndex+1] != ' ' {
		return NewValidationError(ErrCodeTaskTitleInvalidPrefix,
			fmt.Sprintf("task[%d].title: must have a space after ':' (got: '%s')",
				index, title)).
			WithField("task.title").WithIndex(index)
	}

	// タスク名が空または空白のみ
	taskNamePart := ""
	if colonIndex+1 < len(afterPrefix) {
		taskNamePart = afterPrefix[colonIndex+1:]
	}
	if strings.TrimSpace(taskNamePart) == "" {
		suggestion := fmt.Sprintf("Task %s: <タスク名>", numberPart)
		return NewValidationError(ErrCodeTaskTitleInvalidPrefix,
			fmt.Sprintf("task[%d].title: task name cannot be empty (got: '%s', suggestion: '%s')",
				index, title, suggestion)).
			WithField("task.title").WithIndex(index)
	}

	// タスク名に改行が含まれる
	if strings.ContainsAny(title, "\r\n") {
		return NewValidationError(ErrCodeTaskTitleInvalidPrefix,
			fmt.Sprintf("task[%d].title: task name cannot contain newline characters",
				index)).
			WithField("task.title").WithIndex(index)
	}

	// その他の形式エラー
	return NewValidationError(ErrCodeTaskTitleInvalidPrefix,
		fmt.Sprintf("task[%d].title: invalid format (got: '%s', expected: 'Task N: タスク名')",
			index, title)).
		WithField("task.title").WithIndex(index)
}

// ValidateTaskNumberSequence はタスク番号が1から厳密に連続しているかを検証します
// タスクは配列の順番通りに1, 2, 3, 4... と並んでいる必要があります
func ValidateTaskNumberSequence(tasks []Task) error {
	if len(tasks) == 0 {
		return nil
	}

	// 各タスクの番号を取得し、重複チェックと順序チェックを行う
	taskNumbers := make([]int, len(tasks)) // 各タスクの番号を保存
	seen := make(map[int]int)              // taskNumber -> taskIndex (重複検出用)

	for i, task := range tasks {
		taskNum, _, err := ValidateTaskTitleFormat(task.Title, i)
		if err != nil {
			return err // 形式エラーは先に返す
		}

		// 重複チェック
		if existingIndex, exists := seen[taskNum]; exists {
			return NewValidationError(ErrCodeTaskNumberDuplicate,
				fmt.Sprintf("duplicate task number %d found at task[%d] and task[%d]",
					taskNum, existingIndex, i)).
				WithField("task.title").WithIndex(i)
		}
		seen[taskNum] = i
		taskNumbers[i] = taskNum // 番号を保存
	}

	// 順序チェック（保存した番号を再利用）
	for i, taskNum := range taskNumbers {
		// 期待される番号は i+1（0-indexed なので）
		expectedNum := i + 1
		if taskNum != expectedNum {
			return NewValidationError(ErrCodeTaskNumberNotSequential,
				fmt.Sprintf("task[%d].title: expected Task %d but got Task %d (tasks must be numbered sequentially: 1, 2, 3, ...)",
					i, expectedNum, taskNum)).
				WithField("task.title").WithIndex(i)
		}
	}

	return nil
}
