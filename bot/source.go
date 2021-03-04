package bot

import (
	"context"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/ndrewnee/lesswrong-bot/models"
)

func (b *Bot) ChangeSource(ctx context.Context, update tgbotapi.Update) (string, error) {
	key := fmt.Sprintf("source:%d", update.Message.From.ID)

	cachedSource, err := b.storage.Get(ctx, key)
	if err != nil {
		log.Printf("[ERROR] Get source failed: %s, key: %s", err, key)
	}

	source := models.Source(cachedSource)
	if !source.IsValid() {
		source = models.SourceLesswrongRu
	}

	arg := update.Message.CommandArguments()
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
