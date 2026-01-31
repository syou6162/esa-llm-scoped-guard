package guard

import (
	"bytes"
	"fmt"
	"strings"
)

// ExtractEmbeddedYAML extracts YAML from Markdown (parse only, no schema validation)
func ExtractEmbeddedYAML(markdown string) (*PostInput, error) {
	data := []byte(markdown)

	// 1. Check input size (10MB max for scan limit)
	if len(data) > MaxInputSize {
		return nil, fmt.Errorf("input size exceeds %d bytes (got %d bytes)", MaxInputSize, len(data))
	}

	// 2. Check if document starts with sentinel (exact match, no BOM/whitespace allowed)
	if !bytes.HasPrefix(data, []byte(Sentinel)) {
		return nil, fmt.Errorf("sentinel not found at start of document")
	}

	// 3. Find first closing tag "\n-->"
	closingIdx := bytes.Index(data, []byte(ClosingTag))
	if closingIdx == -1 {
		return nil, fmt.Errorf("closing tag not found")
	}

	// 3. Extract YAML block (skip sentinel, before closing tag)
	yamlStart := len(Sentinel)
	yamlBlock := data[yamlStart:closingIdx]

	// 4. Check YAML block size (10MB max, before parsing)
	if len(yamlBlock) > MaxYAMLSize {
		return nil, fmt.Errorf("YAML block size exceeds %d bytes (got %d bytes)", MaxYAMLSize, len(yamlBlock))
	}

	// 5. Parse YAML using DecodeYAMLSecure (includes validation)
	var input PostInput
	if err := DecodeYAMLSecure(strings.NewReader(string(yamlBlock)), &input, MaxYAMLSize); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// 6. Return parsed input
	return &input, nil
}
