package guard

import (
	"strings"
	"testing"
)

func TestGenerateMarkdownWithJSON(t *testing.T) {
	postNum := 123
	input := &PostInput{
		PostNumber: &postNum,
		Name:       "Test Post",
		Category:   "LLM/Test/2026/01/31",
		Body: Body{
			Background: "Test background",
			Tasks: []Task{
				{
					ID:          "task-1",
					Title:       "Task 1: Test task",
					Status:      TaskStatusNotStarted,
					Summary:     []string{"Test summary"},
					Description: "Test description",
				},
			},
		},
	}

	markdown, err := GenerateMarkdownWithJSON(input)
	if err != nil {
		t.Fatalf("GenerateMarkdownWithJSON() error = %v", err)
	}

	// Check sentinel at start
	if !strings.HasPrefix(markdown, "<!-- esa-guard-json\n") {
		t.Errorf("Markdown should start with sentinel, got: %s", markdown[:50])
	}

	// Check closing tag exists
	if !strings.Contains(markdown, "\n-->") {
		t.Error("Markdown should contain closing tag")
	}

	// Check JSON is compact (no pretty print)
	lines := strings.Split(markdown, "\n")
	if len(lines) < 3 {
		t.Fatal("Expected at least 3 lines (sentinel, JSON, closing)")
	}

	// Line 2 should be compact JSON (single line)
	jsonLine := lines[1]
	if !strings.HasPrefix(jsonLine, "{") {
		t.Errorf("Expected JSON line to start with {, got: %s", jsonLine[:10])
	}

	// Check Markdown content follows after closing tag
	if !strings.Contains(markdown, "## サマリー") {
		t.Error("Expected Markdown content after JSON")
	}
}

func TestGenerateMarkdownWithJSON_NoLeadingNewline(t *testing.T) {
	postNum := 123
	input := &PostInput{
		PostNumber: &postNum,
		Name:       "Test",
		Category:   "LLM/Test/2026/01/31",
		Body: Body{
			Background: "test",
			Tasks: []Task{
				{ID: "task-1", Title: "Task 1: Test", Status: TaskStatusNotStarted, Summary: []string{"test"}, Description: "test"},
			},
		},
	}

	markdown, err := GenerateMarkdownWithJSON(input)
	if err != nil {
		t.Fatalf("GenerateMarkdownWithJSON() error = %v", err)
	}

	// Ensure no leading newline before sentinel
	if markdown[0] == '\n' || markdown[0] == '\r' {
		t.Error("Markdown should not start with newline")
	}

	// Ensure exactly 2 newlines between closing tag and content
	expected := "\n-->\n\n## サマリー"
	if !strings.Contains(markdown, expected) {
		t.Errorf("Expected exactly 2 newlines between closing tag and content")
	}
}

func TestGenerateMarkdown_NoLeadingWhitespace(t *testing.T) {
	body := &Body{
		Background: "test",
		Tasks: []Task{
			{ID: "task-1", Title: "Task 1: Test", Status: TaskStatusNotStarted, Summary: []string{"test"}, Description: "test"},
		},
	}

	markdown := GenerateMarkdown(body)

	// Check no leading whitespace or newline
	if len(markdown) > 0 && (markdown[0] == ' ' || markdown[0] == '\t' || markdown[0] == '\n' || markdown[0] == '\r') {
		t.Errorf("GenerateMarkdown should not start with whitespace, got: %q", markdown[:10])
	}

	// Should start with "##"
	if !strings.HasPrefix(markdown, "##") {
		t.Errorf("Expected GenerateMarkdown to start with '##', got: %s", markdown[:10])
	}
}
