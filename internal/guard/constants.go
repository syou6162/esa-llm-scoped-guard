package guard

const (
	// MaxInputSize is the maximum size for input files and API responses (10MB)
	MaxInputSize = 10 * 1024 * 1024

	// MaxYAMLSize is the maximum size for embedded YAML blocks (10MB)
	MaxYAMLSize = 10 * 1024 * 1024

	// Sentinel is the opening tag for embedded YAML in Markdown
	Sentinel = "<!-- esa-guard-yaml\n"

	// ClosingTag is the closing tag for embedded YAML in Markdown
	ClosingTag = "\n-->"
)
