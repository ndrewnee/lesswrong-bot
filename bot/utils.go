package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// handleResponse handles the common logic for API responses and unmarshals JSON
func (b *Bot) handleResponse(httpResponse *http.Response, target interface{}) error {
	defer httpResponse.Body.Close()

	// Check if response is successful
	if httpResponse.StatusCode != 0 && httpResponse.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(httpResponse.Body)
		return fmt.Errorf("API returned status %d: %s", httpResponse.StatusCode, string(bodyBytes))
	}

	// Read the response body to check if it's valid JSON
	bodyBytes, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return fmt.Errorf("read response body failed: %s", err)
	}

	// Check if response starts with HTML (error page)
	if len(bodyBytes) > 0 && bodyBytes[0] == '<' {
		return fmt.Errorf("API returned HTML instead of JSON: %s", string(bodyBytes[:min(200, len(bodyBytes))]))
	}

	// Unmarshal JSON into target
	if err := json.Unmarshal(bodyBytes, target); err != nil {
		return fmt.Errorf("unmarshal failed: %s", err)
	}

	return nil
}
