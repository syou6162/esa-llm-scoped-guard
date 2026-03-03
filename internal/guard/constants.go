package guard

const (
	// MaxInputSize is the maximum size for input files and API responses (10MB)
	MaxInputSize = 10 * 1024 * 1024

	// MaxYAMLSize is the maximum size for embedded YAML blocks within Markdown (10MB)
	// Note: This is unified with MaxInputSize for simplicity. In practice, the actual
	// embeddable YAML size is smaller than 10MB because the total embedded Markdown
	// (sentinel + YAML + closing tag + content) must not exceed MaxInputSize.
	// Unified from previous 2MB JSON limit to accommodate YAML's more verbose format
	// while maintaining DoS protection.
	MaxYAMLSize = 10 * 1024 * 1024

	// MaxMessageSize is the maximum size for post message (10KB)
	MaxMessageSize = 10 * 1024

	// Sentinel is the opening tag for embedded YAML in Markdown
	Sentinel = "<!-- esa-guard-yaml\n"

	// ClosingTag is the closing tag for embedded YAML in Markdown
	ClosingTag = "\n-->"
)
