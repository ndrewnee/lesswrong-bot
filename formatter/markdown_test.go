package formatter

import (
	"strings"
	"testing"

	md "github.com/JohannesKaufmann/html-to-markdown"

	"github.com/ndrewnee/lesswrong-bot/models"
)

func TestMarkdownFormatter_FixTelegramMarkdown(t *testing.T) {
	formatter := NewMarkdownFormatter()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Fix unmatched underscore",
			input:    "_[Epistemic status: very low. Total conjecture based on insufficient evidence.\\",
			expected: "[Epistemic status: very low. Total conjecture based on insufficient evidence.",
		},
		{
			name:     "Fix escaped brackets",
			input:    "Some text \\[with brackets\\]",
			expected: "Some text [with brackets]",
		},
		{
			name:     "Fix incomplete escape at end",
			input:    "Some text ending with \\",
			expected: "Some text ending with",
		},
		{
			name:     "Fix unmatched asterisks",
			input:    "**Bold text* with unmatched",
			expected: "**Bold text with unmatched",
		},
		{
			name:     "Leave matched markdown alone",
			input:    "_italic_ and **bold** text",
			expected: "_italic_ and **bold** text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.fixTelegramMarkdown(tt.input)
			if result != tt.expected {
				t.Errorf("fixTelegramMarkdown() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestMarkdownFormatter_FormatPost_HandlesProblematicMarkdown(t *testing.T) {
	formatter := NewMarkdownFormatter()
	converter := md.NewConverter("", true, nil)

	// Simulate the problematic HTML from the error
	problematicHTML := `<em>[Epistemic status: very low. Total conjecture based on insufficient evidence.\</em>
<p>Some content here...</p>`

	post := models.Post{
		Title: "Test Post",
		URL:   "https://example.com/test",
		HTML:  problematicHTML,
	}

	result, err := formatter.FormatPost(post, converter, false)
	if err != nil {
		t.Fatalf("FormatPost() failed: %v", err)
	}

	// The result should not contain unmatched underscores or incomplete escapes
	if result == "" {
		t.Error("FormatPost() returned empty result")
	}

	// Should not contain incomplete escape sequences
	if strings.Contains(result, "\\[") && !strings.Contains(result, "\\]") {
		t.Error("Result contains incomplete escape sequence")
	}

	// Should not contain unmatched underscores
	underscoreCount := strings.Count(result, "_")
	if underscoreCount%2 != 0 {
		t.Errorf("Result contains unmatched underscores: %d", underscoreCount)
	}
}