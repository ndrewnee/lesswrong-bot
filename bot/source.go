package bot

import (
	"context"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/ndrewnee/lesswrong-bot/models"
)

var sourceKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Lesswrong.ru", "1"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Slate Start Codex", "2"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Astral Codex Ten", "3"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Lesswrong.com", "4"),
	),
)

func (b *Bot) ChangeSource(ctx context.Context, userID int, arg string) (string, interface{}, error) {
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
		return "Current source is " + source.String(), sourceKeyboard, nil
	}

	newSource := models.Source(arg)
	if !newSource.IsValid() {
		return "New source is invalid. Current source is " + source.String(), sourceKeyboard, nil
	}

	if err := b.storage.Set(ctx, key, newSource.Value(), 0); err != nil {
		return "", nil, fmt.Errorf("set source failed: %s, key: %s, source: %s", err, key, newSource)
	}

	return "Changed source to " + newSource.String(), nil, nil
}
