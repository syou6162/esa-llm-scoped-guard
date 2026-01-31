package guard

import (
	"fmt"
	"strings"
	"testing"

	"github.com/syou6162/esa-llm-scoped-guard/internal/esa"
)

type mockFetchClient struct {
	bodyMD string
	err    error
}

func (m *mockFetchClient) CreatePost(post *esa.PostInput) (*esa.Post, error) {
	return nil, fmt.Errorf("CreatePost should not be called in fetch")
}

func (m *mockFetchClient) UpdatePost(postNumber int, post *esa.PostInput) (*esa.Post, error) {
	return nil, fmt.Errorf("UpdatePost should not be called in fetch")
}

func (m *mockFetchClient) GetPost(postNumber int) (*esa.Post, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &esa.Post{
		Number:   postNumber,
		Name:     "Test Post",
		Category: "LLM/Test/2026/01/31",
		BodyMD:   m.bodyMD,
	}, nil
}

func TestExecuteFetch_Success(t *testing.T) {
	bodyMD := `<!-- esa-guard-yaml
post_number: 123
name: Test
category: LLM/Test/2026/01/31
body:
  background: test
  tasks:
    - id: task-1
      title: "Task 1: Test"
      status: not_started
      summary:
        - test
      description: test
-->

## サマリー
- [ ] Task 1: Test`

	client := &mockFetchClient{bodyMD: bodyMD}

	output, err := executeFetchWithClient(123, client)
	if err != nil {
		t.Fatalf("executeFetchWithClient() error = %v", err)
	}

	// Check output is pretty-printed YAML
	if !strings.Contains(output, "post_number: 123") {
		t.Error("Expected post_number in output")
	}

	if !strings.Contains(output, "name: Test") {
		t.Error("Expected name in output")
	}
}

func TestExecuteFetch_NoEmbeddedYAML(t *testing.T) {
	bodyMD := `## Regular Markdown

No YAML here.`

	client := &mockFetchClient{bodyMD: bodyMD}

	_, err := executeFetchWithClient(123, client)
	if err == nil {
		t.Fatal("Expected error for missing embedded YAML")
	}

	if !strings.Contains(err.Error(), "no embedded YAML found in post") {
		t.Errorf("Expected 'no embedded YAML found in post' error, got: %v", err)
	}
}

func TestExecuteFetch_EmptyBody(t *testing.T) {
	client := &mockFetchClient{bodyMD: ""}

	_, err := executeFetchWithClient(123, client)
	if err == nil {
		t.Fatal("Expected error for empty body")
	}

	if !strings.Contains(err.Error(), "post body is empty") {
		t.Errorf("Expected 'post body is empty' error, got: %v", err)
	}
}

func TestExecuteFetch_BodyTooLarge(t *testing.T) {
	// Create body larger than 10MB (MaxInputSize + 1, boundary test)
	largeBody := "<!-- esa-guard-yaml\n{}\n-->\n" + strings.Repeat("a", MaxInputSize+1)

	client := &mockFetchClient{bodyMD: largeBody}

	_, err := executeFetchWithClient(123, client)
	if err == nil {
		t.Fatal("Expected error for body exceeding 10MB")
	}

	if !strings.Contains(err.Error(), "exceeds") {
		t.Errorf("Expected size exceed error, got: %v", err)
	}
}

func TestExecuteFetch_BodyExactly10MB(t *testing.T) {
	// Create body exactly 10MB (boundary test)
	baseContent := "<!-- esa-guard-yaml\n{}\n-->\n"
	largeBody := baseContent + strings.Repeat("a", MaxInputSize-len(baseContent))

	client := &mockFetchClient{bodyMD: largeBody}

	// Exactly 10MB should succeed (no size error)
	_, err := executeFetchWithClient(123, client)
	// May fail on JSON extraction but not on size check
	if err != nil && strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("Expected no size error for exactly 10MB, got: %v", err)
	}
}

func TestExecuteFetch_BodyJustUnder10MB(t *testing.T) {
	// Create body just under 10MB (MaxInputSize - 1, boundary test)
	baseContent := "<!-- esa-guard-yaml\ncreate_new: true\n-->\n"
	largeBody := baseContent + strings.Repeat("a", MaxInputSize-len(baseContent)-1)

	client := &mockFetchClient{bodyMD: largeBody}

	// Just under 10MB should succeed (no size error)
	_, err := executeFetchWithClient(123, client)
	// May fail on YAML extraction but not on size check
	if err != nil && strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("Expected no size error for body just under 10MB, got: %v", err)
	}
}

func TestExecuteFetch_InvalidYAML(t *testing.T) {
	bodyMD := `<!-- esa-guard-yaml
[unclosed
-->

Content`

	client := &mockFetchClient{bodyMD: bodyMD}

	_, err := executeFetchWithClient(123, client)
	if err == nil {
		t.Fatal("Expected error for invalid YAML")
	}

	if !strings.Contains(err.Error(), "invalid YAML in post") {
		t.Errorf("Expected 'invalid YAML in post' error, got: %v", err)
	}
}

func TestExecuteFetch_PostNumberMismatch(t *testing.T) {
	// Embedded YAML has post_number 999, but we request 123
	bodyMD := `<!-- esa-guard-yaml
post_number: 999
name: Test
category: LLM/Test/2026/01/31
body:
  background: test
  tasks:
    - id: task-1
      title: "Task 1: Test"
      status: not_started
      summary:
        - test
      description: test
-->

## サマリー
- [ ] Task 1: Test`

	client := &mockFetchClient{bodyMD: bodyMD}

	_, err := executeFetchWithClient(123, client)
	if err == nil {
		t.Fatal("Expected error for post_number mismatch")
	}

	if !strings.Contains(err.Error(), "post_number mismatch") {
		t.Errorf("Expected 'post_number mismatch' error, got: %v", err)
	}

	if !strings.Contains(err.Error(), "embedded YAML has 999") {
		t.Errorf("Expected error to mention embedded post_number 999, got: %v", err)
	}

	if !strings.Contains(err.Error(), "requested 123") {
		t.Errorf("Expected error to mention requested post_number 123, got: %v", err)
	}
}

func TestExecuteFetch_PostNumberNil(t *testing.T) {
	// Embedded JSON has no post_number (nil) - should be rejected (fail closed)
	bodyMD := `<!-- esa-guard-yaml
{"name":"Test","category":"LLM/Test/2026/01/31","body":{"background":"test","tasks":[{"id":"task-1","title":"Task 1: Test","status":"not_started","summary":["test"],"description":"test"}]}}
-->

## サマリー
- [ ] Task 1: Test`

	client := &mockFetchClient{bodyMD: bodyMD}

	_, err := executeFetchWithClient(123, client)
	if err == nil {
		t.Fatal("Expected error for nil post_number (fetch targets existing posts only)")
	}

	if !strings.Contains(err.Error(), "post_number is required") {
		t.Errorf("Expected 'post_number is required' error, got: %v", err)
	}
}
