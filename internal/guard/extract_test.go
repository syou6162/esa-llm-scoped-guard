package guard

import (
	"strings"
	"testing"
)

func TestExtractEmbeddedYAML_Success(t *testing.T) {
	markdown := `<!-- esa-guard-yaml
create_new: false
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

## Test Content`

	input, err := ExtractEmbeddedYAML(markdown)
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

func TestExtractEmbeddedYAML_WithBOM(t *testing.T) {
	markdown := "\xEF\xBB\xBF<!-- esa-guard-yaml\ncreate_new: true\n-->"

	_, err := ExtractEmbeddedYAML(markdown)
	if err == nil {
		t.Fatal("Expected error for BOM at start, got nil")
	}
	if !strings.Contains(err.Error(), "sentinel not found at start") {
		t.Errorf("Expected 'sentinel not found at start' error, got: %v", err)
	}
}

func TestExtractEmbeddedYAML_WithLeadingWhitespace(t *testing.T) {
	markdown := " <!-- esa-guard-yaml\ncreate_new: true\n-->"

	_, err := ExtractEmbeddedYAML(markdown)
	if err == nil {
		t.Fatal("Expected error for leading whitespace, got nil")
	}
	if !strings.Contains(err.Error(), "sentinel not found at start") {
		t.Errorf("Expected 'sentinel not found at start' error, got: %v", err)
	}
}

func TestExtractEmbeddedYAML_WithLeadingNewline(t *testing.T) {
	markdown := "\n<!-- esa-guard-yaml\ncreate_new: true\n-->"

	_, err := ExtractEmbeddedYAML(markdown)
	if err == nil {
		t.Fatal("Expected error for leading newline, got nil")
	}
	if !strings.Contains(err.Error(), "sentinel not found at start") {
		t.Errorf("Expected 'sentinel not found at start' error, got: %v", err)
	}
}

func TestExtractEmbeddedYAML_SentinelNotAtStart(t *testing.T) {
	markdown := "Some text\n<!-- esa-guard-yaml\ncreate_new: true\n-->"

	_, err := ExtractEmbeddedYAML(markdown)
	if err == nil {
		t.Fatal("Expected error for sentinel not at start, got nil")
	}
	if !strings.Contains(err.Error(), "sentinel not found at start") {
		t.Errorf("Expected 'sentinel not found at start' error, got: %v", err)
	}
}

func TestExtractEmbeddedYAML_NoClosingTag(t *testing.T) {
	markdown := "<!-- esa-guard-yaml\ncreate_new: true"

	_, err := ExtractEmbeddedYAML(markdown)
	if err == nil {
		t.Fatal("Expected error for missing closing tag, got nil")
	}
	if !strings.Contains(err.Error(), "closing tag not found") {
		t.Errorf("Expected 'closing tag not found' error, got: %v", err)
	}
}

func TestExtractEmbeddedYAML_InvalidYAML(t *testing.T) {
	markdown := "<!-- esa-guard-yaml\n[unclosed\n-->"

	_, err := ExtractEmbeddedYAML(markdown)
	if err == nil {
		t.Fatal("Expected error for invalid YAML, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse YAML") {
		t.Errorf("Expected 'failed to parse YAML' error, got: %v", err)
	}
}

func TestExtractEmbeddedYAML_YAMLWithHTMLCommentStart(t *testing.T) {
	markdown := `<!-- esa-guard-yaml
name: "<!--test"
-->`

	// Extraction should succeed (parse only, no validation)
	// Validation of <!-- and --> happens in validator.go
	input, err := ExtractEmbeddedYAML(markdown)
	if err != nil {
		t.Fatalf("Expected no error (extraction is parse-only), got: %v", err)
	}
	if input.Name != "<!--test" {
		t.Errorf("Expected Name '<!--test', got: %s", input.Name)
	}
}

func TestExtractEmbeddedYAML_YAMLWithHTMLCommentEnd(t *testing.T) {
	markdown := `<!-- esa-guard-yaml
name: "-->test"
-->`

	// Extraction should succeed (parse only, no validation)
	// Validation of <!-- and --> happens in validator.go
	input, err := ExtractEmbeddedYAML(markdown)
	if err != nil {
		t.Fatalf("Expected no error (extraction is parse-only), got: %v", err)
	}
	if input.Name != "-->test" {
		t.Errorf("Expected Name '-->test', got: %s", input.Name)
	}
}

func TestExtractEmbeddedYAML_LargeYAMLBlock(t *testing.T) {
	// Create YAML just over 2MB
	largeString := strings.Repeat("a", MaxYAMLSize+1)
	markdown := "<!-- esa-guard-yaml\ndata: \"" + largeString + "\"\n-->"

	_, err := ExtractEmbeddedYAML(markdown)
	if err == nil {
		t.Fatal("Expected error for YAML block exceeding 2MB, got nil")
	}
	if !strings.Contains(err.Error(), "input size exceeds") {
		t.Errorf("Expected 'input size exceeds' error, got: %v", err)
	}
}

func TestExtractEmbeddedYAML_LargeYAMLBlockExactly2MB(t *testing.T) {
	// Create YAML block that results in markdown exactly at MaxInputSize
	// markdown = Sentinel + YAML + ClosingTag
	// MaxInputSize = 10MB, so YAML block size = MaxInputSize - len(Sentinel) - len(ClosingTag)
	yamlContentOverhead := len("data: \"\"")
	maxYAMLBlockSize := MaxInputSize - len(Sentinel) - len(ClosingTag)
	largeString := strings.Repeat("a", maxYAMLBlockSize-yamlContentOverhead)
	markdown := "<!-- esa-guard-yaml\ndata: \"" + largeString + "\"\n-->"

	_, err := ExtractEmbeddedYAML(markdown)
	// Exactly at MaxInputSize should succeed (may fail validation but not size check)
	if err != nil && strings.Contains(err.Error(), "input size exceeds") {
		t.Fatalf("Expected no size error for exactly MaxInputSize, got: %v", err)
	}
}

func TestExtractEmbeddedYAML_LargeYAMLBlockWithinLimit(t *testing.T) {
	// Create YAML just under MaxInputSize (MaxInputSize - 1 byte for content, boundary test)
	yamlContentOverhead := len("data: \"\"")
	maxYAMLBlockSize := MaxInputSize - len(Sentinel) - len(ClosingTag)
	largeString := strings.Repeat("a", maxYAMLBlockSize-yamlContentOverhead-1)
	markdown := "<!-- esa-guard-yaml\ndata: \"" + largeString + "\"\n-->"

	_, err := ExtractEmbeddedYAML(markdown)
	// This should succeed (parse the YAML structure, may fail validation but not size check)
	// Validation errors like "required", "field not found" are acceptable - we only care about size errors
	if err != nil && !strings.Contains(err.Error(), "validation") && !strings.Contains(err.Error(), "required") && !strings.Contains(err.Error(), "field") {
		t.Fatalf("Expected no size error, got: %v", err)
	}
}

func TestExtractEmbeddedYAML_NewPostWithZeroPostNumber(t *testing.T) {
	markdown := `<!-- esa-guard-yaml
create_new: true
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

## Test`

	input, err := ExtractEmbeddedYAML(markdown)
	if err != nil {
		t.Fatalf("Expected no error for new post, got: %v", err)
	}

	if input.PostNumber != nil {
		t.Errorf("Expected nil PostNumber for new post, got: %v", *input.PostNumber)
	}
}

func TestExtractEmbeddedYAML_FakeSentinelInBody(t *testing.T) {
	// Large markdown with fake sentinel in body
	fakeContent := strings.Repeat("Lorem ipsum dolor sit amet. ", 10000)
	markdown := `<!-- esa-guard-yaml
create_new: false
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

## Real Content

` + fakeContent + `

<!-- esa-guard-yaml (fake)
This should be ignored
-->`

	input, err := ExtractEmbeddedYAML(markdown)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if input.PostNumber == nil || *input.PostNumber != 123 {
		t.Errorf("Expected PostNumber 123, got: %v", input.PostNumber)
	}
}

// TestExtractEmbeddedYAML_InputSizeExactly10MB tests extraction with exactly 10MB input
func TestExtractEmbeddedYAML_InputSizeExactly10MB(t *testing.T) {
	// Create a markdown with exactly 10MB total size
	baseContent := "<!-- esa-guard-yaml\ncreate_new: true\n-->\n"
	padding := strings.Repeat("a", MaxInputSize-len(baseContent))
	markdown := baseContent + padding

	if len(markdown) != MaxInputSize {
		t.Fatalf("Expected exactly %d bytes, got %d", MaxInputSize, len(markdown))
	}

	_, err := ExtractEmbeddedYAML(markdown)
	if err != nil && strings.Contains(err.Error(), "input size exceeds") {
		t.Fatalf("Expected no size error for exactly 10MB, got: %v", err)
	}
}

// TestExtractEmbeddedYAML_InputSizeExceeds10MB tests extraction with 10MB+1 input (should fail)
func TestExtractEmbeddedYAML_InputSizeExceeds10MB(t *testing.T) {
	// Create a markdown with 10MB+1 bytes
	baseContent := "<!-- esa-guard-yaml\ncreate_new: true\n-->\n"
	padding := strings.Repeat("a", MaxInputSize-len(baseContent)+1)
	markdown := baseContent + padding

	if len(markdown) != MaxInputSize+1 {
		t.Fatalf("Expected exactly %d bytes, got %d", MaxInputSize+1, len(markdown))
	}

	_, err := ExtractEmbeddedYAML(markdown)
	if err == nil {
		t.Fatal("Expected size error for 10MB+1, got nil")
	}
	if !strings.Contains(err.Error(), "input size exceeds") {
		t.Errorf("Expected 'input size exceeds' error, got: %v", err)
	}
}
