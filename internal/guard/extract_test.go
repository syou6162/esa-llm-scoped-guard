package guard

import (
	"strings"
	"testing"
)

func TestExtractEmbeddedJSON_Success(t *testing.T) {
	markdown := `<!-- esa-guard-json
{"create_new":false,"post_number":123,"name":"Test","category":"LLM/Test/2026/01/31","body":{"background":"test","tasks":[{"id":"task-1","title":"Task 1: Test","status":"not_started","summary":["test"],"description":"test"}]}}
-->

## Test Content`

	input, err := ExtractEmbeddedJSON(markdown)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if input.PostNumber == nil || *input.PostNumber != 123 {
		t.Errorf("Expected PostNumber 123, got: %v", input.PostNumber)
	}
	if input.Name != "Test" {
		t.Errorf("Expected Name 'Test', got: %s", input.Name)
	}
}

func TestExtractEmbeddedJSON_WithBOM(t *testing.T) {
	markdown := "\xEF\xBB\xBF<!-- esa-guard-json\n{}\n-->"

	_, err := ExtractEmbeddedJSON(markdown)
	if err == nil {
		t.Fatal("Expected error for BOM at start, got nil")
	}
	if !strings.Contains(err.Error(), "sentinel not found at start") {
		t.Errorf("Expected 'sentinel not found at start' error, got: %v", err)
	}
}

func TestExtractEmbeddedJSON_WithLeadingWhitespace(t *testing.T) {
	markdown := " <!-- esa-guard-json\n{}\n-->"

	_, err := ExtractEmbeddedJSON(markdown)
	if err == nil {
		t.Fatal("Expected error for leading whitespace, got nil")
	}
	if !strings.Contains(err.Error(), "sentinel not found at start") {
		t.Errorf("Expected 'sentinel not found at start' error, got: %v", err)
	}
}

func TestExtractEmbeddedJSON_WithLeadingNewline(t *testing.T) {
	markdown := "\n<!-- esa-guard-json\n{}\n-->"

	_, err := ExtractEmbeddedJSON(markdown)
	if err == nil {
		t.Fatal("Expected error for leading newline, got nil")
	}
	if !strings.Contains(err.Error(), "sentinel not found at start") {
		t.Errorf("Expected 'sentinel not found at start' error, got: %v", err)
	}
}

func TestExtractEmbeddedJSON_SentinelNotAtStart(t *testing.T) {
	markdown := "Some text\n<!-- esa-guard-json\n{}\n-->"

	_, err := ExtractEmbeddedJSON(markdown)
	if err == nil {
		t.Fatal("Expected error for sentinel not at start, got nil")
	}
	if !strings.Contains(err.Error(), "sentinel not found at start") {
		t.Errorf("Expected 'sentinel not found at start' error, got: %v", err)
	}
}

func TestExtractEmbeddedJSON_NoClosingTag(t *testing.T) {
	markdown := "<!-- esa-guard-json\n{}"

	_, err := ExtractEmbeddedJSON(markdown)
	if err == nil {
		t.Fatal("Expected error for missing closing tag, got nil")
	}
	if !strings.Contains(err.Error(), "closing tag not found") {
		t.Errorf("Expected 'closing tag not found' error, got: %v", err)
	}
}

func TestExtractEmbeddedJSON_InvalidJSON(t *testing.T) {
	markdown := "<!-- esa-guard-json\n{invalid json}\n-->"

	_, err := ExtractEmbeddedJSON(markdown)
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse JSON") {
		t.Errorf("Expected 'failed to parse JSON' error, got: %v", err)
	}
}

func TestExtractEmbeddedJSON_JSONWithHTMLCommentStart(t *testing.T) {
	markdown := `<!-- esa-guard-json
{"name":"<!--test"}
-->`

	// Extraction should succeed (parse only, no validation)
	// Validation of <!-- and --> happens in validator.go
	input, err := ExtractEmbeddedJSON(markdown)
	if err != nil {
		t.Fatalf("Expected no error (extraction is parse-only), got: %v", err)
	}
	if input.Name != "<!--test" {
		t.Errorf("Expected Name '<!--test', got: %s", input.Name)
	}
}

func TestExtractEmbeddedJSON_JSONWithHTMLCommentEnd(t *testing.T) {
	markdown := `<!-- esa-guard-json
{"name":"-->test"}
-->`

	// Extraction should succeed (parse only, no validation)
	// Validation of <!-- and --> happens in validator.go
	input, err := ExtractEmbeddedJSON(markdown)
	if err != nil {
		t.Fatalf("Expected no error (extraction is parse-only), got: %v", err)
	}
	if input.Name != "-->test" {
		t.Errorf("Expected Name '-->test', got: %s", input.Name)
	}
}

func TestExtractEmbeddedJSON_LargeJSONBlock(t *testing.T) {
	// Create JSON just over 2MB
	largeString := strings.Repeat("a", MaxJSONSize+1)
	markdown := "<!-- esa-guard-json\n{\"data\":\"" + largeString + "\"}\n-->"

	_, err := ExtractEmbeddedJSON(markdown)
	if err == nil {
		t.Fatal("Expected error for JSON block exceeding 2MB, got nil")
	}
	if !strings.Contains(err.Error(), "JSON block size exceeds") {
		t.Errorf("Expected 'JSON block size exceeds' error, got: %v", err)
	}
}

func TestExtractEmbeddedJSON_LargeJSONBlockWithinLimit(t *testing.T) {
	// Create JSON just under 2MB (accounting for JSON structure overhead)
	largeString := strings.Repeat("a", MaxJSONSize-200)
	markdown := "<!-- esa-guard-json\n{\"data\":\"" + largeString + "\"}\n-->"

	_, err := ExtractEmbeddedJSON(markdown)
	// This should succeed (parse the JSON structure)
	if err != nil && !strings.Contains(err.Error(), "validation") && !strings.Contains(err.Error(), "required") {
		t.Fatalf("Expected no size error, got: %v", err)
	}
}

func TestExtractEmbeddedJSON_NewPostWithZeroPostNumber(t *testing.T) {
	markdown := `<!-- esa-guard-json
{"create_new":true,"name":"Test","category":"LLM/Test/2026/01/31","body":{"background":"test","tasks":[{"id":"task-1","title":"Task 1: Test","status":"not_started","summary":["test"],"description":"test"}]}}
-->

## Test`

	input, err := ExtractEmbeddedJSON(markdown)
	if err != nil {
		t.Fatalf("Expected no error for new post, got: %v", err)
	}

	if input.PostNumber != nil {
		t.Errorf("Expected nil PostNumber for new post, got: %v", *input.PostNumber)
	}
}

func TestExtractEmbeddedJSON_FakeSentinelInBody(t *testing.T) {
	// Large markdown with fake sentinel in body
	fakeContent := strings.Repeat("Lorem ipsum dolor sit amet. ", 10000)
	markdown := `<!-- esa-guard-json
{"create_new":false,"post_number":123,"name":"Test","category":"LLM/Test/2026/01/31","body":{"background":"test","tasks":[{"id":"task-1","title":"Task 1: Test","status":"not_started","summary":["test"],"description":"test"}]}}
-->

## Real Content

` + fakeContent + `

<!-- esa-guard-json (fake)
This should be ignored
-->`

	input, err := ExtractEmbeddedJSON(markdown)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if input.PostNumber == nil || *input.PostNumber != 123 {
		t.Errorf("Expected PostNumber 123, got: %v", input.PostNumber)
	}
}
