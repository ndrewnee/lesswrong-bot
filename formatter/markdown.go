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

	return markdown
}