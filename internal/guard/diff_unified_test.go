package guard

import (
	"strings"
	"testing"
)

// TestGenerateUnifiedDiff_EOFNewlineRemoved tests that removing a trailing newline is visible in the diff
func TestGenerateUnifiedDiff_EOFNewlineRemoved(t *testing.T) {
	old := "a\n"
	new := "a"

	diff := generateUnifiedDiff(old, new)

	if diff == "" {
		t.Error("expected non-empty diff when EOF newline is removed")
	}

	// Should contain both old and new versions of the line
	if !strings.Contains(diff, "-a") {
		t.Error("diff should contain '-a' (old line with newline)")
	}
	if !strings.Contains(diff, "+a") {
		t.Error("diff should contain '+a' (new line without newline)")
	}
}

// TestGenerateUnifiedDiff_EOFNewlineAdded tests that adding a trailing newline is visible in the diff
func TestGenerateUnifiedDiff_EOFNewlineAdded(t *testing.T) {
	old := "a"
	new := "a\n"

	diff := generateUnifiedDiff(old, new)

	if diff == "" {
		t.Error("expected non-empty diff when EOF newline is added")
	}

	// Should contain both old and new versions of the line
	if !strings.Contains(diff, "-a") {
		t.Error("diff should contain '-a' (old line without newline)")
	}
	if !strings.Contains(diff, "+a") {
		t.Error("diff should contain '+a' (new line with newline)")
	}
}

// TestGenerateUnifiedDiff_TrailingEmptyLineRemoved tests that removing a trailing empty line is visible
func TestGenerateUnifiedDiff_TrailingEmptyLineRemoved(t *testing.T) {
	old := "a\n\n"
	new := "a\n"

	diff := generateUnifiedDiff(old, new)

	if diff == "" {
		t.Error("expected non-empty diff when trailing empty line is removed")
	}

	// Should show deletion of empty line
	lines := strings.Split(diff, "\n")
	hasMinusBlank := false
	for _, line := range lines {
		if line == "-" {
			hasMinusBlank = true
			break
		}
	}

	if !hasMinusBlank {
		t.Error("diff should contain deletion of empty line (line starting with '-' and nothing else)")
	}
}

// TestGenerateUnifiedDiff_TrailingEmptyLineAdded tests that adding a trailing empty line is visible
func TestGenerateUnifiedDiff_TrailingEmptyLineAdded(t *testing.T) {
	old := "a\n"
	new := "a\n\n"

	diff := generateUnifiedDiff(old, new)

	if diff == "" {
		t.Error("expected non-empty diff when trailing empty line is added")
	}

	// Should show addition of empty line
	lines := strings.Split(diff, "\n")
	hasPlusBlank := false
	for _, line := range lines {
		if line == "+" {
			hasPlusBlank = true
			break
		}
	}

	if !hasPlusBlank {
		t.Error("diff should contain addition of empty line (line starting with '+' and nothing else)")
	}
}

// TestGenerateUnifiedDiff_ChangeAtFileStart tests correct line numbers for changes at file start
func TestGenerateUnifiedDiff_ChangeAtFileStart(t *testing.T) {
	old := "old\nx\n"
	new := "new\nx\n"

	diff := generateUnifiedDiff(old, new)

	if diff == "" {
		t.Error("expected non-empty diff")
	}

	// Should have hunk starting at line 1
	if !strings.Contains(diff, "@@ -1,") {
		t.Error("hunk should start at line 1 for old file")
	}
	if !strings.Contains(diff, "+1,") {
		t.Error("hunk should start at line 1 for new file")
	}
}

// TestGenerateUnifiedDiff_ChangeAtFileEnd tests correct line numbers for changes at file end
func TestGenerateUnifiedDiff_ChangeAtFileEnd(t *testing.T) {
	old := "x\na\n"
	new := "x\nb\n"

	diff := generateUnifiedDiff(old, new)

	if diff == "" {
		t.Error("expected non-empty diff")
	}

	// Should contain the change
	if !strings.Contains(diff, "-a") {
		t.Error("diff should contain '-a'")
	}
	if !strings.Contains(diff, "+b") {
		t.Error("diff should contain '+b'")
	}
}

// TestGenerateUnifiedDiff_MultipleHunksSplit tests that hunks are split when separated by enough context
func TestGenerateUnifiedDiff_MultipleHunksSplit(t *testing.T) {
	// Two changes separated by 4 unchanged lines (> contextLines=3)
	old := "a\nx\nx\nx\nx\nb\n"
	new := "aa\nx\nx\nx\nx\nbb\n"

	diff := generateUnifiedDiff(old, new)

	if diff == "" {
		t.Error("expected non-empty diff")
	}

	// Count hunk markers
	hunkCount := strings.Count(diff, "@@")
	if hunkCount < 4 {
		t.Errorf("expected at least 2 hunks (4+ @@ markers), got %d markers", hunkCount)
	}
}

// TestGenerateUnifiedDiff_SingleHunkNotSplit tests that hunks are not split when changes are close
func TestGenerateUnifiedDiff_SingleHunkNotSplit(t *testing.T) {
	// Two changes separated by 2 unchanged lines (<= contextLines=3)
	old := "a\nx\nx\nb\n"
	new := "aa\nx\nx\nbb\n"

	diff := generateUnifiedDiff(old, new)

	if diff == "" {
		t.Error("expected non-empty diff")
	}

	// Count hunk markers - should be exactly 1 hunk (2 @@ markers)
	hunkCount := strings.Count(diff, "@@")
	if hunkCount != 2 {
		t.Errorf("expected exactly 1 hunk (2 @@ markers), got %d markers", hunkCount)
	}
}
