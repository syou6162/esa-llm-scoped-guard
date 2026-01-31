package guard

const (
	// MaxInputSize is the maximum size for input files and API responses (10MB)
	MaxInputSize = 10 * 1024 * 1024

	// MaxYAMLSize is the maximum size for embedded YAML blocks (10MB)
	// Note: Unified with MaxInputSize from previous 2MB limit to simplify size validation
	// and align with overall input size constraints. YAML is more verbose than JSON,
	// so the larger limit accommodates realistic use cases while maintaining DoS protection.
	MaxYAMLSize = 10 * 1024 * 1024

	// Sentinel is the opening tag for embedded YAML in Markdown
	Sentinel = "<!-- esa-guard-yaml\n"

	// ClosingTag is the closing tag for embedded YAML in Markdown
	ClosingTag = "\n-->"
)
