package formatter

import (
	"fmt"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"

	"github.com/ndrewnee/lesswrong-bot/config"
	"github.com/ndrewnee/lesswrong-bot/models"
)

type MarkdownFormatter struct{}

func NewMarkdownFormatter() *MarkdownFormatter {
	return &MarkdownFormatter{}
}

func (f *MarkdownFormatter) FormatPost(post models.Post, converter *md.Converter, urlWithText bool) (string, error) {
	markdownOrig, err := converter.ConvertString(post.HTML)
	if err != nil {
		return "", fmt.Errorf("convert html to markdown failed: %s", err)
	}

	markdown := f.truncateContent(markdownOrig)
	markdown = f.applyMarkdownFixes(markdown)

	link := fmt.Sprintf("[%s](%s)", post.Title, post.URL)
	postURL := post.URL
	if urlWithText {
		postURL = link
	}

	return fmt.Sprintf("📝 %s\n\n%s\n\n%s", link, markdown, postURL), nil
}

func (f *MarkdownFormatter) truncateContent(markdown string) string {
	if len(markdown) <= config.PostMaxLength {
		return markdown
	}

	// Convert to runes to properly split between unicode symbols.
	runes := []rune(markdown)
	truncated := string(runes[:config.PostMaxLength])

	// Truncate after next line end to not break markdown text.
	rest := string(runes[config.PostMaxLength:])
	if n := strings.IndexByte(rest, '\n'); n != -1 {
		truncated += rest[:n]
	} else {
		return markdown // Return original if we can't find a good truncation point
	}

	return f.cleanupTruncatedMarkdown(truncated)
}

func (f *MarkdownFormatter) cleanupTruncatedMarkdown(markdown string) string {
	// Clean up artifacts from truncation
	markdown = strings.ReplaceAll(markdown, "* * *", "")
	markdown = strings.ReplaceAll(markdown, "```", "")
	return markdown
}

func (f *MarkdownFormatter) applyMarkdownFixes(markdown string) string {
	// Apply various markdown fixes for better Telegram compatibility
	fixes := []struct {
		old, new string
	}{
		{"[[", "["},
		{"]]", "]"},
		{"![]", "[Image]"},
		{"_[", ""},
		{"]_", ""},
	}

	for _, fix := range fixes {
		markdown = strings.ReplaceAll(markdown, fix.old, fix.new)
	}

	// Additional Telegram-specific fixes
	markdown = f.fixTelegramMarkdown(markdown)

	return markdown
}

func (f *MarkdownFormatter) fixTelegramMarkdown(markdown string) string {
	// Fix specific problematic patterns first
	markdown = f.fixUnmatchedBrackets(markdown)
	
	// Then fix unmatched emphasis markers
	markdown = f.fixUnmatchedUnderscores(markdown)
	markdown = f.fixUnmatchedAsterisks(markdown)
	
	// Clean up line endings
	lines := strings.Split(markdown, "\n")
	for i, line := range lines {
		lines[i] = f.cleanLineEnding(line)
	}
	
	return strings.Join(lines, "\n")
}

func (f *MarkdownFormatter) fixUnmatchedUnderscores(text string) string {
	// Count underscores and remove trailing ones if unmatched
	underscoreCount := strings.Count(text, "_")
	if underscoreCount%2 != 0 {
		// Remove the last underscore if count is odd
		lastIndex := strings.LastIndex(text, "_")
		if lastIndex != -1 {
			text = text[:lastIndex] + text[lastIndex+1:]
		}
	}
	return text
}

func (f *MarkdownFormatter) fixUnmatchedBrackets(text string) string {
	// Remove incomplete bracket sequences like "\[" at the end
	text = strings.TrimSuffix(text, "\\[")
	text = strings.TrimSuffix(text, "\\")
	
	// Fix common bracket patterns
	text = strings.ReplaceAll(text, "\\[", "[")
	text = strings.ReplaceAll(text, "\\]", "]")
	
	return text
}

func (f *MarkdownFormatter) fixUnmatchedAsterisks(text string) string {
	// Handle both single (*italic*) and double (**bold**) asterisks
	// Count remaining single asterisks after removing double asterisks
	remainingText := strings.ReplaceAll(text, "**", "")
	singleAsteriskCount := strings.Count(remainingText, "*")
	
	// If we have unmatched single asterisks, remove the last one
	if singleAsteriskCount%2 != 0 {
		lastIndex := strings.LastIndex(text, "*")
		// Make sure we're not breaking a double asterisk
		if lastIndex > 0 && text[lastIndex-1] != '*' && lastIndex < len(text)-1 && text[lastIndex+1] != '*' {
			text = text[:lastIndex] + text[lastIndex+1:]
		} else if lastIndex == len(text)-1 && (lastIndex == 0 || text[lastIndex-1] != '*') {
			// It's a trailing single asterisk
			text = text[:lastIndex]
		}
	}
	
	return text
}

func (f *MarkdownFormatter) cleanLineEnding(line string) string {
	// Remove problematic characters at the end of lines
	line = strings.TrimSuffix(line, "\\")
	// Only trim trailing underscores/asterisks if they would be unmatched
	return strings.TrimSpace(line)
}