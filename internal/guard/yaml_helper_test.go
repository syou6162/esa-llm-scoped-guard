package guard

import (
	"errors"
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
