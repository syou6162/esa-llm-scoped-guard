package guard

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/syou6162/esa-llm-scoped-guard/internal/esa"
)

// mockEsaClientForExecute is a mock implementation of EsaClientInterface for execute tests
type mockEsaClientForExecute struct {
	createPostFunc func(*esa.PostInput) (*esa.Post, error)
	updatePostFunc func(int, *esa.PostInput) (*esa.Post, error)
	getPostFunc    func(int) (*esa.Post, error)
}

func (m *mockEsaClientForExecute) CreatePost(input *esa.PostInput) (*esa.Post, error) {
	if m.createPostFunc != nil {
		return m.createPostFunc(input)
	}
	return &esa.Post{Number: 123}, nil
}

func (m *mockEsaClientForExecute) UpdatePost(number int, input *esa.PostInput) (*esa.Post, error) {
	if m.updatePostFunc != nil {
		return m.updatePostFunc(number, input)
	}
	return &esa.Post{Number: number}, nil
}

func (m *mockEsaClientForExecute) GetPost(number int) (*esa.Post, error) {
	if m.getPostFunc != nil {
		return m.getPostFunc(number)
	}
	return &esa.Post{
		Number:   number,
		Category: "Claude Code/開発日誌/2026/01/28",
		Tags:     []string{},
	}, nil
}

// TestExecutePost_CreateNewUpdatesYAML tests that YAML file is automatically updated after successful post with create_new
func TestExecutePost_CreateNewUpdatesYAML(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.yaml")

	// 新規作成用のYAMLファイルを作成（create_new: true）
	inputYAML := `create_new: true
name: Test Post
category: Claude Code/開発日誌/2026/01/28
body:
  background: Test background
  tasks:
    - id: task-1
      title: "Task 1: Test task"
      status: not_started
      summary:
        - Task summary
      description: Task description
`

	if err := os.WriteFile(tmpFile, []byte(inputYAML), 0644); err != nil {
		t.Fatal(err)
	}

	// モッククライアント（新規作成を返す）
	mockClient := &mockEsaClientForExecute{
		createPostFunc: func(input *esa.PostInput) (*esa.Post, error) {
			return &esa.Post{Number: 999}, nil
		},
	}

	allowedCategories := []string{"Claude Code/開発日誌"}

	// ExecutePost実行（内部でYAML更新が行われるはず）
	err := executePostWithClient(tmpFile, allowedCategories, mockClient, "テスト用メッセージ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// YAMLファイルが自動更新されているか確認
	updatedInput, err := ReadPostInputFromFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read updated YAML: %v", err)
	}

	// create_newがfalseになっている
	if updatedInput.CreateNew {
		t.Error("expected CreateNew to be false after successful post")
	}

	// post_numberが設定されている
	if updatedInput.PostNumber == nil {
		t.Error("expected PostNumber to be set after successful post")
	} else if *updatedInput.PostNumber != 999 {
		t.Errorf("expected PostNumber to be 999, got %d", *updatedInput.PostNumber)
	}
}

// TestExecutePost_UpdateDoesNotChangeYAML tests that YAML file is not modified when updating existing post
func TestExecutePost_UpdateDoesNotChangeYAML(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.yaml")

	// 更新用のYAMLファイルを作成（post_number: 123）
	inputYAML := `post_number: 123
name: Test Post
category: Claude Code/開発日誌/2026/01/28
body:
  background: Test background
  tasks:
    - id: task-1
      title: "Task 1: Test task"
      status: not_started
      summary:
        - Task summary
      description: Task description
`

	if err := os.WriteFile(tmpFile, []byte(inputYAML), 0644); err != nil {
		t.Fatal(err)
	}

	// モッククライアント（更新を返す）
	mockClient := &mockEsaClientForExecute{
		getPostFunc: func(number int) (*esa.Post, error) {
			return &esa.Post{
				Number:   123,
				Category: "Claude Code/開発日誌/2026/01/28",
				Tags:     []string{},
			}, nil
		},
		updatePostFunc: func(number int, input *esa.PostInput) (*esa.Post, error) {
			return &esa.Post{Number: 123}, nil
		},
	}

	allowedCategories := []string{"Claude Code/開発日誌"}

	// ExecutePost実行（更新なのでYAMLは変更されないはず）
	err := executePostWithClient(tmpFile, allowedCategories, mockClient, "テスト用メッセージ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// YAMLファイルが変更されていないことを確認
	updatedInput, err := ReadPostInputFromFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read YAML: %v", err)
	}

	// create_newはfalseのまま
	if updatedInput.CreateNew {
		t.Error("expected CreateNew to remain false")
	}

	// post_numberは123のまま
	if updatedInput.PostNumber == nil {
		t.Error("expected PostNumber to remain set")
	} else if *updatedInput.PostNumber != 123 {
		t.Errorf("expected PostNumber to remain 123, got %d", *updatedInput.PostNumber)
	}
}

// TestExecutePost_CreateFailureDoesNotChangeYAML tests that YAML file is not modified when post creation fails
func TestExecutePost_CreateFailureDoesNotChangeYAML(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.yaml")

	// 新規作成用のYAMLファイルを作成（create_new: true）
	inputYAML := `create_new: true
name: Test Post
category: Claude Code/開発日誌/2026/01/28
body:
  background: Test background
  tasks:
    - id: task-1
      title: "Task 1: Test task"
      status: not_started
      summary:
        - Task summary
      description: Task description
`

	if err := os.WriteFile(tmpFile, []byte(inputYAML), 0644); err != nil {
		t.Fatal(err)
	}

	// モッククライアント（新規作成が失敗する）
	mockClient := &mockEsaClientForExecute{
		createPostFunc: func(input *esa.PostInput) (*esa.Post, error) {
			return nil, fmt.Errorf("API error")
		},
	}

	allowedCategories := []string{"Claude Code/開発日誌"}

	// ExecutePost実行（失敗するのでYAMLは変更されないはず）
	err := executePostWithClient(tmpFile, allowedCategories, mockClient, "テスト用メッセージ")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// YAMLファイルが変更されていないことを確認
	updatedInput, err := ReadPostInputFromFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read YAML: %v", err)
	}

	// create_newはtrueのまま
	if !updatedInput.CreateNew {
		t.Error("expected CreateNew to remain true after failed post")
	}

	// post_numberは設定されていない
	if updatedInput.PostNumber != nil {
		t.Errorf("expected PostNumber to remain nil, got %d", *updatedInput.PostNumber)
	}
}

// TestExecutePost_EmbeddsYAMLInMarkdown tests that GenerateMarkdownWithYAML correctly embeds YAML
func TestExecutePost_EmbeddsYAMLInMarkdown(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.yaml")

	postNumber := 123
	// 更新用のYAMLファイルを作成（post_number: 123）
	inputYAML := `post_number: 123
name: Test Post
category: Claude Code/開発日誌/2026/01/28
body:
  background: Test background
  related_links:
    - https://example.com
  instructions:
    - Instruction 1
  tasks:
    - id: task-1
      title: "Task 1: Test task"
      status: not_started
      summary:
        - Task summary
      description: Task description
`

	if err := os.WriteFile(tmpFile, []byte(inputYAML), 0644); err != nil {
		t.Fatal(err)
	}

	var capturedBodyMD string

	// モッククライアント（UpdatePostで送信されるbody_mdをキャプチャ）
	mockClient := &mockEsaClientForExecute{
		getPostFunc: func(number int) (*esa.Post, error) {
			return &esa.Post{
				Number:   123,
				Category: "Claude Code/開発日誌/2026/01/28",
				Tags:     []string{},
			}, nil
		},
		updatePostFunc: func(number int, input *esa.PostInput) (*esa.Post, error) {
			capturedBodyMD = input.BodyMD
			return &esa.Post{Number: 123}, nil
		},
	}

	allowedCategories := []string{"Claude Code/開発日誌"}

	// ExecutePost実行
	err := executePostWithClient(tmpFile, allowedCategories, mockClient, "テスト用メッセージ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 送信されたbody_mdを検証
	if capturedBodyMD == "" {
		t.Fatal("body_md was not captured")
	}

	// 1. 先頭が<!-- esa-guard-yamlで始まっていること
	if len(capturedBodyMD) < len(Sentinel) {
		t.Fatalf("body_md too short: %d bytes", len(capturedBodyMD))
	}
	if capturedBodyMD[:len(Sentinel)] != Sentinel {
		t.Errorf("body_md does not start with sentinel, got: %s", capturedBodyMD[:50])
	}

	// 2. 埋め込まれたYAMLを抽出してパース可能なこと
	extracted, err := ExtractEmbeddedYAML(capturedBodyMD)
	if err != nil {
		t.Fatalf("failed to extract embedded YAML: %v", err)
	}

	// 3. 抽出したYAMLが元のPostInputと一致すること
	if extracted.PostNumber == nil || *extracted.PostNumber != postNumber {
		t.Errorf("expected post_number %d, got %v", postNumber, extracted.PostNumber)
	}
	if extracted.Name != "Test Post" {
		t.Errorf("expected name 'Test Post', got '%s'", extracted.Name)
	}
	if extracted.Category != "Claude Code/開発日誌/2026/01/28" {
		t.Errorf("expected category 'Claude Code/開発日誌/2026/01/28', got '%s'", extracted.Category)
	}
	if extracted.Body.Background != "Test background" {
		t.Errorf("expected background 'Test background', got '%s'", extracted.Body.Background)
	}
	if len(extracted.Body.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(extracted.Body.Tasks))
	}
	if extracted.Body.Tasks[0].Title != "Task 1: Test task" {
		t.Errorf("expected task title 'Task 1: Test task', got '%s'", extracted.Body.Tasks[0].Title)
	}

	// 4. Markdownセクションが含まれていること
	// ExtractEmbeddedYAMLが成功している時点でフォーマットは正しいので、
	// 十分な長さがあることだけ確認
	if len(capturedBodyMD) < 100 {
		t.Errorf("body_md seems too short, expected markdown sections")
	}
}

// TestExecutePost_EmptyMessageReturnsError tests that empty message returns an error
func TestExecutePost_EmptyMessageReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.yaml")

	inputYAML := `post_number: 123
name: Test Post
category: Claude Code/開発日誌/2026/01/28
body:
  background: Test background
  tasks:
    - id: task-1
      title: "Task 1: Test task"
      status: not_started
      summary:
        - Task summary
      description: Task description
`

	if err := os.WriteFile(tmpFile, []byte(inputYAML), 0644); err != nil {
		t.Fatal(err)
	}

	mockClient := &mockEsaClientForExecute{}
	allowedCategories := []string{"Claude Code/開発日誌"}

	err := executePostWithClient(tmpFile, allowedCategories, mockClient, "")
	if err == nil {
		t.Fatal("expected error for empty message, got nil")
	}
}

// TestExecutePost_WhitespaceOnlyMessageReturnsError tests that whitespace-only message returns an error
func TestExecutePost_WhitespaceOnlyMessageReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.yaml")

	inputYAML := `post_number: 123
name: Test Post
category: Claude Code/開発日誌/2026/01/28
body:
  background: Test background
  tasks:
    - id: task-1
      title: "Task 1: Test task"
      status: not_started
      summary:
        - Task summary
      description: Task description
`

	if err := os.WriteFile(tmpFile, []byte(inputYAML), 0644); err != nil {
		t.Fatal(err)
	}

	mockClient := &mockEsaClientForExecute{}
	allowedCategories := []string{"Claude Code/開発日誌"}

	err := executePostWithClient(tmpFile, allowedCategories, mockClient, "   ")
	if err == nil {
		t.Fatal("expected error for whitespace-only message, got nil")
	}
}

// TestExecutePost_MessageTooLongReturnsError tests that a message exceeding 10KB returns an error
func TestExecutePost_MessageTooLongReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.yaml")

	inputYAML := `post_number: 123
name: Test Post
category: Claude Code/開発日誌/2026/01/28
body:
  background: Test background
  tasks:
    - id: task-1
      title: "Task 1: Test task"
      status: not_started
      summary:
        - Task summary
      description: Task description
`

	if err := os.WriteFile(tmpFile, []byte(inputYAML), 0644); err != nil {
		t.Fatal(err)
	}

	mockClient := &mockEsaClientForExecute{}
	allowedCategories := []string{"Claude Code/開発日誌"}

	// 10KBを超えるメッセージ（10,241バイト）
	longMessage := string(make([]byte, MaxMessageSize+1))

	err := executePostWithClient(tmpFile, allowedCategories, mockClient, longMessage)
	if err == nil {
		t.Fatal("expected error for too-long message, got nil")
	}
}

// TestExecutePost_CreateWithMessage tests that message is set on both CreatePost and UpdatePost during creation
func TestExecutePost_CreateWithMessage(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.yaml")

	inputYAML := `create_new: true
name: Test Post
category: Claude Code/開発日誌/2026/01/28
body:
  background: Test background
  tasks:
    - id: task-1
      title: "Task 1: Test task"
      status: not_started
      summary:
        - Task summary
      description: Task description
`

	if err := os.WriteFile(tmpFile, []byte(inputYAML), 0644); err != nil {
		t.Fatal(err)
	}

	var capturedCreateMessage string
	var capturedUpdateMessages []string

	mockClient := &mockEsaClientForExecute{
		createPostFunc: func(input *esa.PostInput) (*esa.Post, error) {
			capturedCreateMessage = input.Message
			return &esa.Post{Number: 999, URL: "https://example.esa.io/posts/999"}, nil
		},
		updatePostFunc: func(number int, input *esa.PostInput) (*esa.Post, error) {
			capturedUpdateMessages = append(capturedUpdateMessages, input.Message)
			return &esa.Post{Number: number, URL: "https://example.esa.io/posts/999"}, nil
		},
	}

	allowedCategories := []string{"Claude Code/開発日誌"}
	testMessage := "タスク状態を更新"

	err := executePostWithClient(tmpFile, allowedCategories, mockClient, testMessage)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// CreatePostにmessageが設定されていること
	if capturedCreateMessage != testMessage {
		t.Errorf("expected CreatePost message %q, got %q", testMessage, capturedCreateMessage)
	}

	// UpdatePost（post_number埋め込み用）にもmessageが設定されていること
	if len(capturedUpdateMessages) == 0 {
		t.Fatal("expected UpdatePost to be called at least once")
	}
	for i, msg := range capturedUpdateMessages {
		if msg != testMessage {
			t.Errorf("expected UpdatePost[%d] message %q, got %q", i, testMessage, msg)
		}
	}
}

// TestExecutePost_UpdateWithMessage tests that message is set on UpdatePost during update
func TestExecutePost_UpdateWithMessage(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.yaml")

	inputYAML := `post_number: 123
name: Test Post
category: Claude Code/開発日誌/2026/01/28
body:
  background: Test background
  tasks:
    - id: task-1
      title: "Task 1: Test task"
      status: not_started
      summary:
        - Task summary
      description: Task description
`

	if err := os.WriteFile(tmpFile, []byte(inputYAML), 0644); err != nil {
		t.Fatal(err)
	}

	var capturedMessage string

	mockClient := &mockEsaClientForExecute{
		getPostFunc: func(number int) (*esa.Post, error) {
			return &esa.Post{
				Number:   123,
				Category: "Claude Code/開発日誌/2026/01/28",
				Tags:     []string{},
			}, nil
		},
		updatePostFunc: func(number int, input *esa.PostInput) (*esa.Post, error) {
			capturedMessage = input.Message
			return &esa.Post{Number: 123, URL: "https://example.esa.io/posts/123"}, nil
		},
	}

	allowedCategories := []string{"Claude Code/開発日誌"}
	testMessage := "開発日誌を更新"

	err := executePostWithClient(tmpFile, allowedCategories, mockClient, testMessage)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if capturedMessage != testMessage {
		t.Errorf("expected UpdatePost message %q, got %q", testMessage, capturedMessage)
	}
}

// TestUpdateYAMLAfterCreate_Roundtrip tests that yaml.Marshal output can be re-parsed
func TestUpdateYAMLAfterCreate_Roundtrip(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.yaml")

	// Create initial YAML with create_new: true
	initialYAML := `create_new: true
name: Initial Post
category: LLM/Tasks/2026/01/30
body:
  background: Initial background
  tasks:
    - id: task-1
      title: "Task 1: Initial task"
      status: not_started
      summary:
        - Initial summary
      description: Initial description
`

	if err := os.WriteFile(tmpFile, []byte(initialYAML), 0644); err != nil {
		t.Fatal(err)
	}

	// Simulate post creation by updating the YAML file
	postNumber := 789
	err := updateYAMLAfterCreate(tmpFile, postNumber)
	if err != nil {
		t.Fatalf("updateYAMLAfterCreate() error = %v", err)
	}

	// Re-read the updated YAML file
	updated, err := ReadPostInputFromFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to re-read updated YAML: %v", err)
	}

	// Verify the updates were applied
	if updated.CreateNew {
		t.Error("Expected create_new to be false after update")
	}
	if updated.PostNumber == nil || *updated.PostNumber != postNumber {
		t.Errorf("Expected post_number %d, got: %v", postNumber, updated.PostNumber)
	}
	if updated.Name != "Initial Post" {
		t.Errorf("Expected name preserved, got: %s", updated.Name)
	}
	if updated.Category != "LLM/Tasks/2026/01/30" {
		t.Errorf("Expected category preserved, got: %s", updated.Category)
	}
	if updated.Body.Background != "Initial background" {
		t.Errorf("Expected background preserved, got: %s", updated.Body.Background)
	}
	if len(updated.Body.Tasks) != 1 {
		t.Errorf("Expected 1 task preserved, got: %d", len(updated.Body.Tasks))
	}
}
