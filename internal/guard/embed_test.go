package guard

import (
	"strings"
	"testing"
)

func TestGenerateMarkdownWithYAML(t *testing.T) {
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

	markdown, err := GenerateMarkdownWithYAML(input)
	if err != nil {
		t.Fatalf("GenerateMarkdownWithYAML() error = %v", err)
	}

	// Check sentinel at start
	if !strings.HasPrefix(markdown, "<!-- esa-guard-yaml\n") {
		t.Errorf("Markdown should start with sentinel, got: %s", markdown[:50])
	}

	// Check closing tag exists
	if !strings.Contains(markdown, "\n-->") {
		t.Error("Markdown should contain closing tag")
	}

	// Check YAML block exists (multiple lines)
	lines := strings.Split(markdown, "\n")
	if len(lines) < 3 {
		t.Fatal("Expected at least 3 lines (sentinel, YAML, closing)")
	}

	// Check Markdown content follows after closing tag
	if !strings.Contains(markdown, "## サマリー") {
		t.Error("Expected Markdown content after YAML")
	}
}

func TestGenerateMarkdownWithYAML_NoLeadingNewline(t *testing.T) {
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

	markdown, err := GenerateMarkdownWithYAML(input)
	if err != nil {
		t.Fatalf("GenerateMarkdownWithYAML() error = %v", err)
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

func TestGenerateMarkdownWithYAML_ExceedsMaxSize(t *testing.T) {
	// Create input that will exceed 10MB when embedded
	largeBackground := strings.Repeat("a", MaxInputSize)

	postNum := 123
	input := &PostInput{
		PostNumber: &postNum,
		Name:       "Test",
		Category:   "LLM/Test/2026/01/31",
		Body: Body{
			Background: largeBackground,
			Tasks: []Task{
				{ID: "task-1", Title: "Task 1: Test", Status: TaskStatusNotStarted, Summary: []string{"test"}, Description: "test"},
			},
		},
	}

	_, err := GenerateMarkdownWithYAML(input)
	if err == nil {
		t.Fatal("Expected error for embedded markdown exceeding 10MB, got nil")
	}

	if !strings.Contains(err.Error(), "embedded markdown exceeds") {
		t.Errorf("Expected error message about embedded markdown size, got: %v", err)
	}
}

// TestGenerateMarkdownWithYAML_Roundtrip tests that yaml.Marshal output can be re-parsed
func TestGenerateMarkdownWithYAML_Roundtrip(t *testing.T) {
	postNum := 456
	input := &PostInput{
		PostNumber: &postNum,
		Name:       "Roundtrip Test",
		Category:   "LLM/Tasks/2026/01/30",
		Body: Body{
			Background:   "Background with multiple lines\nand special chars: コロン：括弧（）",
			RelatedLinks: []string{"https://example.com/link1", "https://example.com/link2"},
			Instructions: []string{"Instruction 1", "Instruction 2 with コロン："},
			Tasks: []Task{
				{
					ID:          "task-1",
					Title:       "Task 1: First task",
					Status:      TaskStatusInProgress,
					Summary:     []string{"Summary line 1", "Summary line 2"},
					Description: "Task description\nwith newlines",
					GitHubURLs:  []string{"https://github.com/owner/repo/pull/1"},
					DependsOn:   []string{},
				},
				{
					ID:          "task-2",
					Title:       "Task 2: Second task",
					Status:      TaskStatusNotStarted,
					Summary:     []string{"Summary for task 2"},
					Description: "Another description",
					DependsOn:   []string{"task-1"},
				},
			},
		},
	}

	// Generate markdown with embedded YAML
	markdown, err := GenerateMarkdownWithYAML(input)
	if err != nil {
		t.Fatalf("GenerateMarkdownWithYAML() error = %v", err)
	}

	// Re-parse the embedded YAML
	reparsed, err := ExtractEmbeddedYAML(markdown)
	if err != nil {
		t.Fatalf("Failed to re-parse embedded YAML: %v", err)
	}

	// Verify structure is preserved
	if reparsed.PostNumber == nil || *reparsed.PostNumber != 456 {
		t.Errorf("Expected post_number 456, got: %v", reparsed.PostNumber)
	}
	if reparsed.Name != "Roundtrip Test" {
		t.Errorf("Expected name 'Roundtrip Test', got: %s", reparsed.Name)
	}
	if reparsed.Category != "LLM/Tasks/2026/01/30" {
		t.Errorf("Expected category 'LLM/Tasks/2026/01/30', got: %s", reparsed.Category)
	}
	if reparsed.Body.Background != input.Body.Background {
		t.Errorf("Background mismatch: expected %q, got %q", input.Body.Background, reparsed.Body.Background)
	}
	if len(reparsed.Body.RelatedLinks) != 2 {
		t.Errorf("Expected 2 related links, got: %d", len(reparsed.Body.RelatedLinks))
	}
	if len(reparsed.Body.Instructions) != 2 {
		t.Errorf("Expected 2 instructions, got: %d", len(reparsed.Body.Instructions))
	}
	if len(reparsed.Body.Tasks) != 2 {
		t.Errorf("Expected 2 tasks, got: %d", len(reparsed.Body.Tasks))
	}
	if len(reparsed.Body.Tasks[0].DependsOn) != 0 {
		t.Errorf("Expected task-1 depends_on to be empty, got: %v", reparsed.Body.Tasks[0].DependsOn)
	}
	if len(reparsed.Body.Tasks[1].DependsOn) != 1 || reparsed.Body.Tasks[1].DependsOn[0] != "task-1" {
		t.Errorf("Expected task-2 depends_on ['task-1'], got: %v", reparsed.Body.Tasks[1].DependsOn)
	}
}
