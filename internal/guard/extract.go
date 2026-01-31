package guard

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// ExtractEmbeddedJSON extracts JSON from Markdown (parse only, no schema validation)
func ExtractEmbeddedJSON(markdown string) (*PostInput, error) {
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

	// 3. Extract JSON block (skip sentinel, before closing tag)
	jsonStart := len(Sentinel)
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
