package guard

import (
	"encoding/json"
	"fmt"
)

// GenerateMarkdownWithJSON generates Markdown with embedded JSON comment at the start.
// JSON is compact (single line) to save space and reduce risk of --> injection.
// The generated Markdown structure is:
//
//	<!-- esa-guard-json
//	{...compact JSON...}
//	-->
//
//	## サマリー
//	...
func GenerateMarkdownWithJSON(input *PostInput) (string, error) {
	// Marshal to compact JSON (no pretty print)
	jsonBytes, err := json.Marshal(input)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err) // fail closed
	}

	// Generate Markdown content
	markdown := GenerateMarkdown(&input.Body)

	// Ensure GenerateMarkdown doesn't start with whitespace/newline (invariant)
	// This is a safety check to prevent breaking sentinel extraction
	if len(markdown) > 0 && (markdown[0] == ' ' || markdown[0] == '\t' || markdown[0] == '\n' || markdown[0] == '\r') {
		// Normalize by trimming leading whitespace (defensive programming)
		markdown = trimLeadingWhitespace(markdown)
	}

	// Embed JSON comment at start: sentinel + JSON + closing + 2 newlines + content
	embedded := fmt.Sprintf("<!-- esa-guard-json\n%s\n-->\n\n%s", string(jsonBytes), markdown)

	return embedded, nil
}

// trimLeadingWhitespace removes leading whitespace and newlines from the start of a string.
// This is a defensive function to ensure GenerateMarkdown invariant is maintained.
func trimLeadingWhitespace(s string) string {
	start := 0
	for start < len(s) {
		c := s[start]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			start++
		} else {
			break
		}
	}
	return s[start:]
}
