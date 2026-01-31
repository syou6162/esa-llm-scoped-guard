package guard

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/syou6162/esa-llm-scoped-guard/internal/esa"
)

type mockEsaClient struct {
	getPostFunc func(number int) (*esa.Post, error)
}

func (m *mockEsaClient) CreatePost(input *esa.PostInput) (*esa.Post, error) {
	return nil, nil
}

func (m *mockEsaClient) UpdatePost(number int, input *esa.PostInput) (*esa.Post, error) {
	return nil, nil
}

func (m *mockEsaClient) GetPost(number int) (*esa.Post, error) {
	return m.getPostFunc(number)
}

func TestExecuteDiff_WithPostNumber(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "update.json")

	updateJSON := `{
		"post_number": 123,
		"name": "Test Post",
		"category": "LLM/Tasks/2026/01/28",
		"body": {
			"background": "New background",
			"tasks": [
				{
					"id": "task-1",
					"title": "Task 1: New task",
					"status": "in_progress",
					"summary": ["New summary"],
					"description": "New description"
				}
			]
		}
	}`

	if err := os.WriteFile(tmpFile, []byte(updateJSON), 0600); err != nil {
		t.Fatal(err)
	}

	mockClient := &mockEsaClient{
		getPostFunc: func(number int) (*esa.Post, error) {
			return &esa.Post{
				Number:   123,
				Name:     "Old Post",
				Category: "LLM/Tasks/2026/01/28",
				BodyMD:   "## サマリー\n- [ ] Old task\n\n## 背景\nOld background",
			}, nil
		},
	}

	allowedCategories := []string{"LLM/Tasks"}

	var output string
	var execErr error
	output = captureStdout(func() {
		execErr = executeDiffWithClient(tmpFile, allowedCategories, mockClient)
	})

	if execErr != nil {
		t.Errorf("expected no error, got %v", execErr)
	}

	if output == "" {
		t.Error("expected diff output, got empty string")
	}

	// unified diff形式（@@ を含む）を検証
	if !strings.Contains(output, "@@") {
		t.Errorf("expected unified diff format with @@, got: %q", output)
	}
}

func TestExecuteDiff_CreateNew(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "new.json")

	newJSON := `{
		"create_new": true,
		"name": "Test Post",
		"category": "LLM/Tasks/2026/01/28",
		"body": {
			"background": "Test background",
			"tasks": [
				{
					"id": "task-1",
					"title": "Task 1: Test",
					"status": "not_started",
					"summary": ["Test summary"],
					"description": "Test description"
				}
			]
		}
	}`

	if err := os.WriteFile(tmpFile, []byte(newJSON), 0600); err != nil {
		t.Fatal(err)
	}

	allowedCategories := []string{"LLM/Tasks"}
	mockClient := &mockEsaClient{}

	// 標準出力をキャプチャ
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := executeDiffWithClient(tmpFile, allowedCategories, mockClient)

	w.Close()
	os.Stdout = oldStdout

	output, _ := io.ReadAll(r)

	if err != nil {
		t.Fatalf("expected no error for create_new, got %v", err)
	}

	// 全行が+で始まる差分が出力されているはず
	if !strings.Contains(string(output), "+## サマリー") {
		t.Error("expected diff output with all lines starting with +")
	}

	// @@ -0,0 +1,N @@ 形式のヘッダーがあるはず
	if !strings.Contains(string(output), "@@ -0,0") {
		t.Error("expected unified diff header with @@ -0,0")
	}
}

func TestExecuteDiff_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.json")

	invalidJSON := `{
		"post_number": 123,
		"name": "Test Post",
		"category": "Invalid/Category"
	}`

	if err := os.WriteFile(tmpFile, []byte(invalidJSON), 0600); err != nil {
		t.Fatal(err)
	}

	allowedCategories := []string{"LLM/Tasks"}
	mockClient := &mockEsaClient{}
	err := executeDiffWithClient(tmpFile, allowedCategories, mockClient)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestExecuteDiff_CategoryNotAllowed(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "update.json")

	updateJSON := `{
		"post_number": 123,
		"name": "Test Post",
		"category": "LLM/Tasks/2026/01/28",
		"body": {
			"background": "New background",
			"tasks": [
				{
					"id": "task-1",
					"title": "Task 1: New task",
					"status": "in_progress",
					"summary": ["New summary"],
					"description": "New description"
				}
			]
		}
	}`

	if err := os.WriteFile(tmpFile, []byte(updateJSON), 0600); err != nil {
		t.Fatal(err)
	}

	// 許可されていないカテゴリの既存記事を返すモック
	mockClient := &mockEsaClient{
		getPostFunc: func(number int) (*esa.Post, error) {
			return &esa.Post{
				Number:   123,
				Name:     "Old Post",
				Category: "Restricted/Category/2026/01/28", // 許可されていないカテゴリ
				BodyMD:   "## サマリー\n- [ ] Old task\n\n## 背景\nOld background",
			}, nil
		},
	}

	allowedCategories := []string{"LLM/Tasks"}
	err := executeDiffWithClient(tmpFile, allowedCategories, mockClient)
	if err == nil {
		t.Fatal("expected error for category not allowed")
	}
	if !strings.Contains(err.Error(), "category") {
		t.Errorf("expected error message to mention category, got: %v", err)
	}
}

func TestExecuteDiff_CategoryChangeAttempt(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "update.json")

	updateJSON := `{
		"post_number": 123,
		"name": "Test Post",
		"category": "LLM/Tasks/2026/01/29",
		"body": {
			"background": "New background",
			"tasks": [
				{
					"id": "task-1",
					"title": "Task 1: New task",
					"status": "in_progress",
					"summary": ["New summary"],
					"description": "New description"
				}
			]
		}
	}`

	if err := os.WriteFile(tmpFile, []byte(updateJSON), 0600); err != nil {
		t.Fatal(err)
	}

	// カテゴリが異なる既存記事を返すモック
	mockClient := &mockEsaClient{
		getPostFunc: func(number int) (*esa.Post, error) {
			return &esa.Post{
				Number:   123,
				Name:     "Old Post",
				Category: "LLM/Tasks/2026/01/28", // 日付が異なる
				BodyMD:   "## サマリー\n- [ ] Old task\n\n## 背景\nOld background",
			}, nil
		},
	}

	allowedCategories := []string{"LLM/Tasks"}
	err := executeDiffWithClient(tmpFile, allowedCategories, mockClient)
	if err == nil {
		t.Fatal("expected error for category change attempt")
	}
	if !strings.Contains(err.Error(), "category change") {
		t.Errorf("expected error message to mention category change, got: %v", err)
	}
}

func TestExecuteDiff_IdenticalContent(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")

	// 既存記事と同じ内容のJSON
	inputJSON := `{
		"post_number": 123,
		"name": "Test Post",
		"category": "LLM/Tasks/2026/01/28",
		"body": {
			"background": "Test background",
			"tasks": [
				{
					"id": "task-1",
					"title": "Task 1: Test task",
					"status": "not_started",
					"summary": ["Task summary"],
					"description": "Task description"
				}
			]
		}
	}`

	if err := os.WriteFile(tmpFile, []byte(inputJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// JSONファイルから入力を読み込んでMarkdownを生成（JSON埋め込みあり）
	input, err := ReadPostInputFromFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	expectedMarkdown, err := GenerateMarkdownWithJSON(input)
	if err != nil {
		t.Fatal(err)
	}

	// 同一内容を返すモック（GenerateMarkdownWithJSONの出力をそのまま使用）
	mockClient := &mockEsaClient{
		getPostFunc: func(number int) (*esa.Post, error) {
			return &esa.Post{
				Number:   123,
				Name:     "Test Post",
				Category: "LLM/Tasks/2026/01/28",
				BodyMD:   expectedMarkdown,
			}, nil
		},
	}

	allowedCategories := []string{"LLM/Tasks"}

	// 標準出力をキャプチャ
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = executeDiffWithClient(tmpFile, allowedCategories, mockClient)

	w.Close()
	os.Stdout = oldStdout

	output, _ := io.ReadAll(r)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 同一内容の場合は空出力であるべき
	if len(output) > 0 {
		t.Errorf("expected empty output for identical content, got:\n%s", string(output))
	}
}

func TestExecuteDiff_InlineChange(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")

	inputJSON := `{
		"post_number": 123,
		"name": "Test Post",
		"category": "LLM/Tasks/2026/01/28",
		"body": {
			"background": "Updated background",
			"tasks": [
				{
					"id": "task-1",
					"title": "Task 1: Test task",
					"status": "not_started",
					"summary": ["Task summary"],
					"description": "Task description"
				}
			]
		}
	}`

	if err := os.WriteFile(tmpFile, []byte(inputJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// 1行内で単語が変わるケース（backgroundの1行だけ変更）
	mockClient := &mockEsaClient{
		getPostFunc: func(number int) (*esa.Post, error) {
			return &esa.Post{
				Number:   123,
				Name:     "Test Post",
				Category: "LLM/Tasks/2026/01/28",
				BodyMD: `## サマリー
- [ ] Task 1: Test task

### 依存関係グラフ

` + "```mermaid" + `
graph TD
    task-1["Task 1: Test task"]:::not_started
    done([タスク完了]):::goal

    task-1 --> done

    classDef completed fill:#90EE90
    classDef in_progress fill:#FFD700
    classDef in_review fill:#FFA500
    classDef not_started fill:#D3D3D3
    classDef goal fill:#87CEEB,stroke:#4169E1,stroke-width:3px
` + "```" + `
## 背景

Original background

## タスク

### Task 1: Test task
- Status: ` + "`not_started`" + `

- 要約:
  - Task summary

<details><summary>詳細を開く</summary>

Task description

</details>
`,
			}, nil
		},
	}

	allowedCategories := []string{"LLM/Tasks"}

	// 標準出力をキャプチャ
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := executeDiffWithClient(tmpFile, allowedCategories, mockClient)

	w.Close()
	os.Stdout = oldStdout

	output, _ := io.ReadAll(r)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	outputStr := string(output)

	// unified diff形式のヘッダーを確認
	if !strings.Contains(outputStr, "---") || !strings.Contains(outputStr, "+++") {
		t.Error("expected unified diff headers (--- and +++)")
	}

	// ハンク形式を確認
	if !strings.Contains(outputStr, "@@") {
		t.Error("expected hunk marker (@@)")
	}

	// 行単位の変更を確認
	if !strings.Contains(outputStr, "-Original background") {
		t.Error("expected deletion line with 'Original background'")
	}
	if !strings.Contains(outputStr, "+Updated background") {
		t.Error("expected addition line with 'Updated background'")
	}
}

func TestExecuteDiff_MultipleHunks(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")

	// 複数の離れた場所に変更があるケース
	inputJSON := `{
		"post_number": 123,
		"name": "Test Post",
		"category": "LLM/Tasks/2026/01/28",
		"body": {
			"background": "Updated background",
			"tasks": [
				{
					"id": "task-1",
					"title": "Task 1: Updated task",
					"status": "not_started",
					"summary": ["Updated summary"],
					"description": "Task description"
				}
			]
		}
	}`

	if err := os.WriteFile(tmpFile, []byte(inputJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// backgroundとtitleが変更されるケース（離れた2箇所の変更）
	mockClient := &mockEsaClient{
		getPostFunc: func(number int) (*esa.Post, error) {
			return &esa.Post{
				Number:   123,
				Name:     "Test Post",
				Category: "LLM/Tasks/2026/01/28",
				BodyMD: `## サマリー
- [ ] Task 1: Original task

### 依存関係グラフ

` + "```mermaid" + `
graph TD
    task-1["Task 1: Original task"]:::not_started
    done([タスク完了]):::goal

    task-1 --> done

    classDef completed fill:#90EE90
    classDef in_progress fill:#FFD700
    classDef in_review fill:#FFA500
    classDef not_started fill:#D3D3D3
    classDef goal fill:#87CEEB,stroke:#4169E1,stroke-width:3px
` + "```" + `
## 背景

Original background

## タスク

### Task 1: Original task
- Status: ` + "`not_started`" + `

- 要約:
  - Original summary

<details><summary>詳細を開く</summary>

Task description

</details>
`,
			}, nil
		},
	}

	allowedCategories := []string{"LLM/Tasks"}

	// 標準出力をキャプチャ
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := executeDiffWithClient(tmpFile, allowedCategories, mockClient)

	w.Close()
	os.Stdout = oldStdout

	output, _ := io.ReadAll(r)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	outputStr := string(output)

	// 実際の差分内容をデバッグ出力
	t.Logf("Diff output:\n%s", outputStr)

	// 複数のハンクマーカーを確認（@@ が2つで1ハンク、2ハンクなら4つ以上）
	hunkCount := strings.Count(outputStr, "@@")
	if hunkCount < 4 {
		t.Errorf("expected at least 2 hunks (4+ @@ markers), got %d", hunkCount)
	}

	// 各変更が含まれていることを確認
	if !strings.Contains(outputStr, "-Original background") {
		t.Error("expected deletion line with 'Original background'")
	}
	if !strings.Contains(outputStr, "+Updated background") {
		t.Error("expected addition line with 'Updated background'")
	}

	// ハンク開始行番号が正しいことを確認（行番号は1から始まる）
	// 最初のハンクは1行目付近から始まるはず
	if !strings.Contains(outputStr, "@@ -1,") && !strings.Contains(outputStr, "@@ -2,") && !strings.Contains(outputStr, "@@ -3,") {
		t.Error("expected first hunk to start near line 1-3")
	}

	// 2つ目以降のハンクの行番号も検証
	// 複数ハンクがある場合、2つ目のハンクは10行目以降から始まるはず
	lines := strings.Split(outputStr, "\n")
	hunkHeaders := []string{}
	for _, line := range lines {
		if strings.HasPrefix(line, "@@") {
			hunkHeaders = append(hunkHeaders, line)
		}
	}
	if len(hunkHeaders) >= 2 {
		// 2つ目のハンクヘッダーを検証
		secondHunk := hunkHeaders[1]
		// 2つ目のハンクは1行目より後から始まるはず
		if strings.Contains(secondHunk, "@@ -1,") {
			t.Errorf("second hunk should not start at line 1: %s", secondHunk)
		}
	}
}
