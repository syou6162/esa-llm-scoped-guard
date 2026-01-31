package guard

import (
	"bytes"
	"encoding/json"
	"fmt"
)

const sentinel = "<!-- esa-guard-json\n"
const closingTag = "\n-->"

// ExtractEmbeddedJSON extracts JSON from Markdown (parse only, no schema validation)
func ExtractEmbeddedJSON(markdown string) (*PostInput, error) {
	data := []byte(markdown)

	// 1. Check if document starts with sentinel (exact match, no BOM/whitespace allowed)
	if !bytes.HasPrefix(data, []byte(sentinel)) {
		return nil, fmt.Errorf("sentinel not found at start of document")
	}

	// 2. Find first closing tag "\n-->"
	closingIdx := bytes.Index(data, []byte(closingTag))
	if closingIdx == -1 {
		return nil, fmt.Errorf("closing tag not found")
	}

	// 3. Extract JSON block (skip sentinel, before closing tag)
	jsonStart := len(sentinel)
	jsonBlock := data[jsonStart:closingIdx]

	// 4. Check JSON block size (2MB max, before parsing)
	if len(jsonBlock) > MaxJSONSize {
		return nil, fmt.Errorf("JSON block size exceeds %d bytes (got %d bytes)", MaxJSONSize, len(jsonBlock))
	}

	// 5. Parse JSON (no schema validation)
	var input PostInput
	if err := json.Unmarshal(jsonBlock, &input); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// 6. Return parsed input
	return &input, nil
}
