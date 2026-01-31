package guard

import (
	"fmt"
	"strings"

	"github.com/syou6162/esa-llm-scoped-guard/internal/esa"
	"gopkg.in/yaml.v3"
)

// ExecuteFetch fetches a post from esa.io and outputs embedded YAML in pretty-print format
func ExecuteFetch(postNumber int, teamName string, accessToken string) error {
	client := esa.NewEsaClient(teamName, accessToken)
	output, err := executeFetchWithClient(postNumber, client)
	if err != nil {
		return err
	}

	fmt.Print(output)
	return nil
}

// executeFetchWithClient fetches a post and extracts embedded YAML (testable version)
func executeFetchWithClient(postNumber int, client esa.EsaClientInterface) (string, error) {
	// 1. Get post from esa.io API
	post, err := client.GetPost(postNumber)
	if err != nil {
		return "", fmt.Errorf("failed to get post: %w", err)
	}

	// 2. Check body size (10MB max)
	if len(post.BodyMD) > MaxInputSize {
		return "", fmt.Errorf("post body exceeds %d bytes limit", MaxInputSize)
	}

	// 3. Check if body is empty
	if post.BodyMD == "" {
		return "", fmt.Errorf("post body is empty")
	}

	// 4. Extract embedded YAML (parse-only, no schema validation)
	input, err := ExtractEmbeddedYAML(post.BodyMD)
	if err != nil {
		// Convert extraction errors to plan-specified error messages
		errMsg := err.Error()
		if strings.Contains(errMsg, "sentinel not found") {
			return "", fmt.Errorf("no embedded YAML found in post %d", postNumber)
		}
		// For other errors (closing tag not found, parse errors, size errors, etc.)
		return "", fmt.Errorf("invalid YAML in post %d: %s", postNumber, errMsg)
	}

	// 5. Check post_number consistency (fail closed security check)
	// fetch command only targets existing posts (post_number required).
	// nil post_number is rejected because fetch is for retrieving existing posts from esa.io.
	if input.PostNumber == nil {
		return "", fmt.Errorf("post_number is required in embedded JSON (fetch targets existing posts only)")
	}
	if *input.PostNumber != postNumber {
		return "", fmt.Errorf("post_number mismatch: embedded JSON has %d, but requested %d", *input.PostNumber, postNumber)
	}

	// 6. Pretty-print YAML for output
	prettyYAML, err := yaml.Marshal(input)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML: %w", err)
	}

	return string(prettyYAML), nil
}
