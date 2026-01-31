package guard

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// GenerateMarkdownWithYAML generates Markdown with embedded YAML comment at the start.
// The generated Markdown structure is:
//
//	<!-- esa-guard-yaml
//	name: ...
//	category: ...
//	...
//	-->
//
//	## サマリー
//	...
func GenerateMarkdownWithYAML(input *PostInput) (string, error) {
	// Marshal to YAML (block style)
	yamlBytes, err := yaml.Marshal(input)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML: %w", err) // fail closed
	}

	// Generate Markdown content
	markdown := GenerateMarkdown(&input.Body)

	// Ensure GenerateMarkdown doesn't start with whitespace/newline (invariant)
	// This is a safety check to prevent breaking sentinel extraction
	if len(markdown) > 0 && (markdown[0] == ' ' || markdown[0] == '\t' || markdown[0] == '\n' || markdown[0] == '\r') {
		// Normalize by trimming leading whitespace (defensive programming)
		markdown = trimLeadingWhitespace(markdown)
	}

	// Embed YAML comment at start: sentinel + YAML + closing + 2 newlines + content
	embedded := fmt.Sprintf("%s%s%s\n\n%s", Sentinel, string(yamlBytes), ClosingTag, markdown)

	// Check total embedded markdown size (10MB limit)
	if len(embedded) > MaxInputSize {
		return "", fmt.Errorf("embedded markdown exceeds %d bytes (got %d bytes)", MaxInputSize, len(embedded))
	}

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
