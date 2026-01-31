package guard

const (
	// MaxInputSize is the maximum size for input files and API responses (10MB)
	MaxInputSize = 10 * 1024 * 1024

	// MaxJSONSize is the maximum size for embedded JSON blocks (2MB)
	MaxJSONSize = 2 * 1024 * 1024

	// Sentinel is the opening tag for embedded JSON in Markdown
	Sentinel = "<!-- esa-guard-json\n"

	// ClosingTag is the closing tag for embedded JSON in Markdown
	ClosingTag = "\n-->"
)
