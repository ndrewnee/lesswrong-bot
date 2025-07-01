package bot

import (
	"context"
)

func (b *Bot) TopPosts(ctx context.Context, userID int) (string, error) {
	source := b.getUserSource(ctx, userID)
	provider := b.providerFactory.CreateTopPostsProvider(source)
	return provider.GetTopPosts(ctx)
}