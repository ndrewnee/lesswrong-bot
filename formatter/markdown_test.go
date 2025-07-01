package formatter

import (
	"strings"
	"testing"

	md "github.com/JohannesKaufmann/html-to-markdown"

	"github.com/ndrewnee/lesswrong-bot/models"
)

func TestMarkdownFormatter_FixesSpecificTelegramParsingIssue(t *testing.T) {
	formatter := NewMarkdownFormatter()

	// Test the specific problematic pattern from the error
	tests := []struct {
		name     string
		input    string
		shouldFix bool
	}{
		{
			name:      "Fix specific problematic pattern _\\[",
			input:     "_\\[Epistemic status: very low. Total conjecture\\",
			shouldFix: true,
		},
		{
			name:      "Fix standalone \\[",
			input:     "Some text \\[with brackets",
			shouldFix: true,
		},
		{
			name:      "Leave valid markdown alone",
			input:     "_italic_ and **bold** and [link](url)",
			shouldFix: false,
		},
		{
			name:      "Fix trailing backslash",
			input:     "Some text ending with \\",
			shouldFix: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.applyMarkdownFixes(tt.input)
			
			if tt.shouldFix {
				// Should not contain problematic patterns
				if strings.Contains(result, "_\\[") {
					t.Errorf("Still contains _\\[ pattern: %s", result)
				}
				if strings.Contains(result, "\\[") {
					t.Errorf("Still contains \\[ pattern: %s", result)
				}
				if strings.HasSuffix(result, "\\") {
					t.Errorf("Still ends with backslash: %s", result)
				}
			} else {
				// Should preserve valid markdown
				if !strings.Contains(result, "_italic_") || !strings.Contains(result, "**bold**") {
					t.Errorf("Valid markdown was broken: %s", result)
				}
			}
		})
	}
}

func TestMarkdownFormatter_FormatPost_HandlesOriginalError(t *testing.T) {
	formatter := NewMarkdownFormatter()
	converter := md.NewConverter("slatestarcodex.com", true, nil)

	// Simulate the exact problematic HTML that caused the original error
	problematicHTML := `<em>\[Epistemic status: very low. Total conjecture based on insufficient evidence.\</em>
<p>"Voodoo death" refers to supposed cases where people died after being cursed by witch doctors.</p>`

	post := models.Post{
		Title: "Devoodooifying Psychology",
		URL:   "https://slatestarcodex.com/2016/08/25/devoodooifying-psychology/",
		HTML:  problematicHTML,
	}

	result, err := formatter.FormatPost(post, converter, false)
	if err != nil {
		t.Fatalf("FormatPost() failed: %v", err)
	}

	// The result should not contain the problematic patterns that break Telegram
	if strings.Contains(result, "_\\[") {
		t.Error("Result still contains _\\[ pattern that breaks Telegram parsing")
	}
	
	if strings.Contains(result, "\\[") {
		t.Error("Result still contains \\[ pattern that breaks Telegram parsing")
	}

	// Should not end with backslash
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if strings.HasSuffix(line, "\\") {
			t.Errorf("Line ends with backslash: %s", line)
		}
	}

	// Should still contain the main content
	if !strings.Contains(result, "Devoodooifying Psychology") {
		t.Error("Result missing title")
	}
	if !strings.Contains(result, "Voodoo death") {
		t.Error("Result missing content")
	}
}