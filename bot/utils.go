package bot

import (
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

// handleLesswrongResponse handles the common logic for Lesswrong API responses
func (b *Bot) handleLesswrongResponse(httpResponse *http.Response, operation string) ([]byte, error) {
	defer httpResponse.Body.Close()

	// Check if response is successful
	if httpResponse.StatusCode != 0 && httpResponse.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(httpResponse.Body)
		return nil, fmt.Errorf("lesswrong.com API returned status %d: %s", httpResponse.StatusCode, string(bodyBytes))
	}

	// Read the response body to check if it's valid JSON
	bodyBytes, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("read lesswrong.com response body failed: %s", err)
	}

	// Check if response starts with HTML (error page)
	if len(bodyBytes) > 0 && bodyBytes[0] == '<' {
		return nil, fmt.Errorf("lesswrong.com API returned HTML instead of JSON: %s", string(bodyBytes[:min(200, len(bodyBytes))]))
	}

	return bodyBytes, nil
} 
