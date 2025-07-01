package bot

import (
	"context"
	"fmt"
	"log"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"

	"github.com/ndrewnee/lesswrong-bot/models"
)

func (b *Bot) RandomPost(ctx context.Context, userID int) (string, error) {
	key := fmt.Sprintf("source:%d", userID)

	source, err := b.storage.Get(ctx, key)
	if err != nil {
		log.Printf("[ERROR] Get source failed: %s, key: %s", err, key)
	}

	sourceModel := models.Source(source)
	if !sourceModel.IsValid() {
		sourceModel = models.SourceLesswrongRu
	}

	provider := b.providerFactory.CreateProvider(sourceModel)
	post, err := provider.GetRandomPost(ctx)
	if err != nil {
		return "", err
	}

	converter := b.providerFactory.GetMarkdownConverter(sourceModel)
	urlWithText := b.providerFactory.ShouldUseURLWithText(sourceModel)

	return b.postToMarkdown(post, converter, urlWithText)
}


func (b *Bot) postToMarkdown(post models.Post, mdConverter *md.Converter, urlWithText bool) (string, error) {
	markdownOrig, err := mdConverter.ConvertString(post.HTML)
	if err != nil {
		return "", fmt.Errorf("convert lesswrong.ru html to markdown failed: %s", err)
	}

	markdown := markdownOrig

	// Cut post for preview mode.
	if len(markdown) > models.PostMaxLength {
		// Convert to runes to properly split between unicode symbols.
		runes := []rune(markdown)
		markdown = string(runes[:models.PostMaxLength])

		// Truncate after next line end to not break markdown text.
		rest := string(runes[models.PostMaxLength:])
		if n := strings.IndexByte(rest, '\n'); n != -1 {
			markdown += rest[:n]
		} else {
			markdown = markdownOrig
		}

		// Stupid hotfixes when markdown was cut in the middle.
		markdown = strings.ReplaceAll(markdown, "* * *", "")
		markdown = strings.ReplaceAll(markdown, "```", "")
	}

	// Stupid hotfixes for some invalid markdowns.
	markdown = strings.ReplaceAll(markdown, "[[", "[")
	markdown = strings.ReplaceAll(markdown, "]]", "]")
	markdown = strings.ReplaceAll(markdown, "![]", "[Image]")
	markdown = strings.ReplaceAll(markdown, "_[", "")
	markdown = strings.ReplaceAll(markdown, "]_", "")

	link := fmt.Sprintf("[%s](%s)", post.Title, post.URL)

	postURL := post.URL
	if urlWithText {
		postURL = link
	}

	return fmt.Sprintf("📝 %s\n\n%s\n\n%s", link, markdown, postURL), nil
}
