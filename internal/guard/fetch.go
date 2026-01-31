package guard

import (
	"encoding/json"
	"fmt"

	"github.com/syou6162/esa-llm-scoped-guard/internal/esa"
)

// ExecuteFetch fetches a post from esa.io and outputs embedded JSON in pretty-print format
func ExecuteFetch(postNumber int, teamName string, accessToken string) error {
	client := esa.NewEsaClient(teamName, accessToken)
	output, err := executeFetchWithClient(postNumber, client)
	if err != nil {
		return err
	}

	fmt.Print(output)
	return nil
}

// executeFetchWithClient fetches a post and extracts embedded JSON (testable version)
func executeFetchWithClient(postNumber int, client esa.EsaClientInterface) (string, error) {
	// 1. Get post from esa.io API
	post, err := client.GetPost(postNumber)
	if err != nil {
		return "", fmt.Errorf("failed to get post: %w", err)
	}

	// 2. Check body size (10MB max)
	if len(post.BodyMD) > MaxInputSize {
		return "", fmt.Errorf("post body exceeds %d bytes (got %d bytes)", MaxInputSize, len(post.BodyMD))
	}

	// 3. Check if body is empty
	if post.BodyMD == "" {
		return "", fmt.Errorf("post body is empty")
	}

	// 4. Extract embedded JSON (parse-only, no schema validation)
	input, err := ExtractEmbeddedJSON(post.BodyMD)
	if err != nil {
		// Add post number to error message for better context
		return "", fmt.Errorf("failed to extract JSON from post %d: %w", postNumber, err)
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

	// 6. Pretty-print JSON for output
	prettyJSON, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(prettyJSON), nil
}
