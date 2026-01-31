package guard

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestDecodeYAMLSecure_EmptyInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "whitespace only",
			input: "   \t\n\r",
		},
		{
			name:  "comment only",
			input: "# just a comment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result PostInput
			err := DecodeYAMLSecure(strings.NewReader(tt.input), &result, 0)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			var vErr *ValidationError
			if !errors.As(err, &vErr) {
				t.Fatalf("expected ValidationError, got %T", err)
			}

			if vErr.Code() != ErrCodeYAMLInvalid {
				t.Errorf("expected ErrCodeYAMLInvalid, got %v", vErr.Code())
			}
		})
	}
}

func TestDecodeYAMLSecure_DuplicateKeys(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "duplicate keys in same mapping",
			input: `name: test
name: duplicate`,
			wantErr: true,
		},
		{
			name: "same key name in different mappings (allowed)",
			input: `name: top-level
category: test
body:
  background: test
  tasks:
    - id: task-1
      title: task-title
      status: not_started
      summary:
        - test
      description: test`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result PostInput
			err := DecodeYAMLSecure(strings.NewReader(tt.input), &result, 0)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				var vErr *ValidationError
				if errors.As(err, &vErr) && vErr.Code() != ErrCodeYAMLInvalid {
					t.Errorf("expected ErrCodeYAMLInvalid, got %v", vErr.Code())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestDecodeYAMLSecure_AliasesAndAnchors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "alias reference",
			input: `name: &anchor test
category: *anchor`,
		},
		{
			name: "anchor without alias",
			input: `name: &anchor test
category: other`,
		},
		{
			name: "nested alias",
			input: `body:
  background: &bg test
  tasks:
    - id: task-1
      title: *bg
      status: not_started
      summary:
        - test
      description: test`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result PostInput
			err := DecodeYAMLSecure(strings.NewReader(tt.input), &result, 0)
			if err == nil {
				t.Fatal("expected error for aliases/anchors, got nil")
			}
			if !strings.Contains(err.Error(), "YAML") {
				t.Errorf("expected YAML error, got: %v", err)
			}
		})
	}
}

func TestDecodeYAMLSecure_FlowStyle(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "flow style mapping",
			input:   "name: test\ncategory: {key: value}",
			wantErr: true,
			errMsg:  "flow style",
		},
		{
			name:    "flow style sequence",
			input:   "name: test\ncategory: [item1, item2]",
			wantErr: true,
			errMsg:  "flow style",
		},
		{
			name: "empty flow style sequence (allowed)",
			input: `name: test
category: LLM/Test/2026/01/31
body:
  background: test
  tasks:
    - id: task-1
      title: test
      status: not_started
      summary:
        - test
      description: test
      depends_on: []`,
			wantErr: false,
		},
		{
			name: "empty flow style mapping (allowed)",
			input: `name: test
category: LLM/Test/2026/01/31
body:
  background: test
  tasks: []`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result PostInput
			err := DecodeYAMLSecure(strings.NewReader(tt.input), &result, 0)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error for flow style, got nil")
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got: %v", tt.errMsg, err)
				}
			} else {
				if err != nil && strings.Contains(err.Error(), "flow style") {
					t.Fatalf("unexpected flow style error: %v", err)
				}
			}
		})
	}
}

func TestDecodeYAMLSecure_MergeKeys(t *testing.T) {
	input := `defaults: &defaults
  status: not_started
task:
  <<: *defaults
  id: task-1`

	var result PostInput
	err := DecodeYAMLSecure(strings.NewReader(input), &result, 0)
	if err == nil {
		t.Fatal("expected error for merge keys, got nil")
	}
	// Either merge key error or anchor error (anchor is detected first)
	if !strings.Contains(err.Error(), "merge") && !strings.Contains(err.Error(), "<<") && !strings.Contains(err.Error(), "anchor") {
		t.Errorf("expected merge key or anchor error, got: %v", err)
	}
}

func TestDecodeYAMLSecure_CustomTags(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "timestamp tag",
			input: `date: !!timestamp 2026-01-31`,
		},
		{
			name:  "binary tag",
			input: `data: !!binary SGVsbG8=`,
		},
		{
			name:  "custom tag",
			input: `custom: !!mytag value`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result PostInput
			err := DecodeYAMLSecure(strings.NewReader(tt.input), &result, 0)
			if err == nil {
				t.Fatal("expected error for custom tags, got nil")
			}
			if !strings.Contains(err.Error(), "tag") {
				t.Errorf("expected tag error, got: %v", err)
			}
		})
	}
}

func TestDecodeYAMLSecure_MultipleDocuments(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "two documents",
			input: `name: test1
---
name: test2`,
		},
		{
			name: "document with empty second",
			input: `name: test
---
`,
		},
		{
			name: "document with comment after separator",
			input: `name: test
---
# comment only`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result PostInput
			err := DecodeYAMLSecure(strings.NewReader(tt.input), &result, 0)
			if err == nil {
				t.Fatal("expected error for multiple documents, got nil")
			}
			if !strings.Contains(err.Error(), "multiple documents") && !strings.Contains(err.Error(), "trailing") {
				t.Errorf("expected multiple documents error, got: %v", err)
			}
		})
	}
}

func TestDecodeYAMLSecure_DepthLimit(t *testing.T) {
	tests := []struct {
		name    string
		depth   int
		wantErr bool
	}{
		{
			name:    "depth 45 (within limit)",
			depth:   45,
			wantErr: false,
		},
		{
			name:    "depth 52 (exceeds limit)",
			depth:   52,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build deeply nested YAML
			var builder strings.Builder
			builder.WriteString("data:\n")

			// Create nested structure
			indent := "  "
			for i := 0; i < tt.depth; i++ {
				builder.WriteString(indent)
				builder.WriteString("level")
				builder.WriteString(string(rune('0' + i%10)))
				builder.WriteString(":\n")
				indent += "  "
			}
			builder.WriteString(indent)
			builder.WriteString("value: deep\n")

			var result map[string]interface{}
			err := DecodeYAMLSecure(strings.NewReader(builder.String()), &result, 0)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected depth limit error, got nil")
				}
				if !strings.Contains(err.Error(), "deep") && !strings.Contains(err.Error(), "nesting") {
					t.Errorf("expected depth error, got: %v", err)
				}
			} else {
				if err != nil && (strings.Contains(err.Error(), "deep") || strings.Contains(err.Error(), "nesting")) {
					t.Fatalf("unexpected depth error: %v", err)
				}
			}
		})
	}
}

func TestDecodeYAMLSecure_NodeCountLimit(t *testing.T) {
	// Build YAML with ~10500 nodes (should exceed 10000 limit)
	var builder strings.Builder
	builder.WriteString("items:\n")

	// Each item has ~2 nodes (key + value), so we need 5250 items to exceed 10000
	for i := 0; i < 5250; i++ {
		builder.WriteString("  item")
		builder.WriteString(strings.Repeat("x", 10))
		// Append counter to make keys unique (avoid duplicate key errors)
		builder.WriteString(fmt.Sprintf("%d", i))
		builder.WriteString(": value\n")
	}

	var result map[string]interface{}
	err := DecodeYAMLSecure(strings.NewReader(builder.String()), &result, 0)
	if err == nil {
		t.Fatal("expected node count limit error, got nil")
	}
	if !strings.Contains(err.Error(), "nodes") && !strings.Contains(err.Error(), "too many") {
		t.Errorf("expected node count error, got: %v", err)
	}
}

func TestDecodeYAMLSecure_UnknownFields(t *testing.T) {
	input := `name: test
category: test
unknown_field: should fail
body:
  background: test
  tasks:
    - id: task-1
      title: test
      status: not_started
      summary:
        - test
      description: test`

	var result PostInput
	err := DecodeYAMLSecure(strings.NewReader(input), &result, 0)
	if err == nil {
		t.Fatal("expected error for unknown field, got nil")
	}
	if !strings.Contains(err.Error(), "field") && !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected unknown field error, got: %v", err)
	}
}

func TestDecodeYAMLSecure_BOM(t *testing.T) {
	input := "\xEF\xBB\xBF" + `name: test
category: LLM/Test/2026/01/31
body:
  background: test
  tasks:
    - id: task-1
      title: test
      status: not_started
      summary:
        - test
      description: test`

	var result PostInput
	err := DecodeYAMLSecure(strings.NewReader(input), &result, 0)
	// BOM should be handled by yaml.v3, but validation errors are OK
	if err != nil && !strings.Contains(err.Error(), "validation") && !strings.Contains(err.Error(), "required") && !strings.Contains(err.Error(), "field") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestDecodeYAMLSecure_NonScalarMapKeys は非スカラーマップキーを拒否することをテスト
func TestDecodeYAMLSecure_NonScalarMapKeys(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "sequence as key",
			input: `? [a, b]
: value`,
		},
		{
			name: "mapping as key",
			input: `? {nested: key}
: value`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result PostInput
			err := DecodeYAMLSecure(strings.NewReader(tt.input), &result, 0)
			if err == nil {
				t.Fatal("expected error for non-scalar map key, got nil")
			}
			if !strings.Contains(err.Error(), "non-scalar") {
				t.Errorf("expected 'non-scalar' in error message, got: %v", err)
			}
		})
	}
}
