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
	bodyMD := `<!-- esa-guard-json
{"post_number":123,"name":"Test","category":"LLM/Test/2026/01/31","body":{"background":"test","tasks":[{"id":"task-1","title":"Task 1: Test","status":"not_started","summary":["test"],"description":"test"}]}}
-->

## サマリー
- [ ] Task 1: Test`

	client := &mockFetchClient{bodyMD: bodyMD}

	output, err := executeFetchWithClient(123, client)
	if err != nil {
		t.Fatalf("executeFetchWithClient() error = %v", err)
	}

	// Check output is pretty-printed JSON
	if !strings.Contains(output, "{\n") {
		t.Error("Expected pretty-printed JSON (with newlines)")
	}

	if !strings.Contains(output, `"post_number": 123`) {
		t.Error("Expected post_number in output")
	}
}

func TestExecuteFetch_NoEmbeddedJSON(t *testing.T) {
	bodyMD := `## Regular Markdown

No JSON here.`

	client := &mockFetchClient{bodyMD: bodyMD}

	_, err := executeFetchWithClient(123, client)
	if err == nil {
		t.Fatal("Expected error for missing embedded JSON")
	}

	if !strings.Contains(err.Error(), "sentinel not found") {
		t.Errorf("Expected 'sentinel not found' error, got: %v", err)
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
	// Create body larger than 10MB
	largeBody := "<!-- esa-guard-json\n{}\n-->\n" + strings.Repeat("a", MaxInputSize+1000)

	client := &mockFetchClient{bodyMD: largeBody}

	_, err := executeFetchWithClient(123, client)
	if err == nil {
		t.Fatal("Expected error for body exceeding 10MB")
	}

	if !strings.Contains(err.Error(), "exceeds") {
		t.Errorf("Expected size exceed error, got: %v", err)
	}
}

func TestExecuteFetch_InvalidJSON(t *testing.T) {
	bodyMD := `<!-- esa-guard-json
{invalid json}
-->

Content`

	client := &mockFetchClient{bodyMD: bodyMD}

	_, err := executeFetchWithClient(123, client)
	if err == nil {
		t.Fatal("Expected error for invalid JSON")
	}

	if !strings.Contains(err.Error(), "failed to parse JSON") {
		t.Errorf("Expected JSON parse error, got: %v", err)
	}
}

func TestExecuteFetch_PostNumberMismatch(t *testing.T) {
	// Embedded JSON has post_number 999, but we request 123
	bodyMD := `<!-- esa-guard-json
{"post_number":999,"name":"Test","category":"LLM/Test/2026/01/31","body":{"background":"test","tasks":[{"id":"task-1","title":"Task 1: Test","status":"not_started","summary":["test"],"description":"test"}]}}
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

	if !strings.Contains(err.Error(), "embedded JSON has 999") {
		t.Errorf("Expected error to mention embedded post_number 999, got: %v", err)
	}

	if !strings.Contains(err.Error(), "requested 123") {
		t.Errorf("Expected error to mention requested post_number 123, got: %v", err)
	}
}

func TestExecuteFetch_PostNumberNil(t *testing.T) {
	// Embedded JSON has no post_number (nil) - should be allowed
	bodyMD := `<!-- esa-guard-json
{"name":"Test","category":"LLM/Test/2026/01/31","body":{"background":"test","tasks":[{"id":"task-1","title":"Task 1: Test","status":"not_started","summary":["test"],"description":"test"}]}}
-->

## サマリー
- [ ] Task 1: Test`

	client := &mockFetchClient{bodyMD: bodyMD}

	output, err := executeFetchWithClient(123, client)
	if err != nil {
		t.Fatalf("executeFetchWithClient() with nil post_number should succeed, got error: %v", err)
	}

	// Check output is pretty-printed JSON
	if !strings.Contains(output, "{\n") {
		t.Error("Expected pretty-printed JSON (with newlines)")
	}
}
