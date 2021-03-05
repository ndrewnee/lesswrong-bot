package bot

import (
	"context"
	"fmt"
	"log"

	"github.com/ndrewnee/lesswrong-bot/models"
)

func (b *Bot) ChangeSource(ctx context.Context, userID int, arg string) (string, error) {
	key := fmt.Sprintf("source:%d", userID)

	cachedSource, err := b.storage.Get(ctx, key)
	if err != nil {
		log.Printf("[ERROR] Get source failed: %s, key: %s", err, key)
	}

	source := models.Source(cachedSource)
	if !source.IsValid() {
		source = models.SourceLesswrongRu
	}

	if arg == "" {
		return "Current source is " + source.String(), nil
	}

	newSource := models.Source(arg)
	if !newSource.IsValid() {
		return "New source is invalid. Current source is " + source.String(), nil
	}

	if err := b.storage.Set(ctx, key, newSource.Value(), 0); err != nil {
		return "", fmt.Errorf("set source failed: %s, key: %s, source: %s", err, key, newSource)
	}

	return "Changed source to " + newSource.String(), nil
}
