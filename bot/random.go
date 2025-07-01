package bot

import (
	"context"

	"github.com/ndrewnee/lesswrong-bot/formatter"
)

func (b *Bot) RandomPost(ctx context.Context, userID int) (string, error) {
	source := b.getUserSource(ctx, userID)

	provider := b.providerFactory.CreateProvider(source)
	post, err := provider.GetRandomPost(ctx)
	if err != nil {
		return "", err
	}

	converter := b.providerFactory.GetMarkdownConverter(source)
	urlWithText := b.providerFactory.ShouldUseURLWithText(source)

	formatter := formatter.NewMarkdownFormatter()
	return formatter.FormatPost(post, converter, urlWithText)
}


