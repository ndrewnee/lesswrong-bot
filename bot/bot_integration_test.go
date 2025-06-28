//go:build integration

package bot

import (
	"context"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/stretchr/testify/require"

	"github.com/ndrewnee/lesswrong-bot/config"
	"github.com/ndrewnee/lesswrong-bot/storage/memory"
	"github.com/ndrewnee/lesswrong-bot/storage/redis"
)

// Test helpers
func setupTestBot(t *testing.T) (*Bot, int64, int) {
	chatID, err := strconv.ParseInt(os.Getenv("TEST_CHAT_ID"), 10, 64)
	require.NoError(t, err, "Env var TEST_CHAT_ID should be set")

	userID, err := strconv.Atoi(os.Getenv("TEST_USER_ID"))
	require.NoError(t, err, "Env var TEST_USER_ID should be set")

	config := config.Parse()
	var storage Storage = memory.NewStorage()

	if os.Getenv("TEST_USE_REDIS") == "true" {
		storage, err = redis.NewStorage(config.RedisURL)
		require.NoError(t, err, "Connect to redis failed")
	}

	tgbot, err := New(Options{Config: config, Storage: storage})
	require.NoError(t, err)

	return tgbot, chatID, userID
}

func createUpdate(userID int, chatID int64, text string, cmdLength int) tgbotapi.Update {
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{ID: userID},
			Chat: &tgbotapi.Chat{ID: chatID},
			Text: text,
		},
	}

	if cmdLength > 0 {
		update.Message.Entities = &[]tgbotapi.MessageEntity{
			{
				Offset: 0,
				Type:   "bot_command",
				Length: cmdLength,
			},
		}
	}

	return update
}

// Individual GetUpdatesChan tests
func TestBot_GetUpdatesChan_ShouldFailWithEmptyWebhookHost(t *testing.T) {
	tgbot, err := New()
	require.NoError(t, err)

	tgbot.config = config.Config{
		Webhook:     true,
		WebhookHost: "",
	}

	got, err := tgbot.GetUpdatesChan()
	require.Error(t, err)
	require.Nil(t, got)
	time.Sleep(time.Second)
}

func TestBot_GetUpdatesChan_ShouldGetWebhookChan(t *testing.T) {
	tgbot, err := New()
	require.NoError(t, err)

	tgbot.config = config.Config{
		Webhook:     true,
		WebhookHost: "https://lesswrong-bot.herokuapp.com",
	}

	got, err := tgbot.GetUpdatesChan()
	require.NoError(t, err)
	require.NotNil(t, got)
	time.Sleep(time.Second)
}

func TestBot_GetUpdatesChan_ShouldGetPollingChan(t *testing.T) {
	tgbot, err := New()
	require.NoError(t, err)

	got, err := tgbot.GetUpdatesChan()
	require.NoError(t, err)
	require.NotNil(t, got)
	time.Sleep(time.Second)
}


// Individual MessageHandler tests
func TestBot_MessageHandler_ShouldFailToAnswerCallbackQuery(t *testing.T) {
	tgbot, _, userID := setupTestBot(t)
	
	update := tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			From: &tgbotapi.User{ID: userID},
			ID:   "invalid",
			Data: "1",
		},
	}
	
	msg, err := tgbot.MessageHandler(context.TODO(), update)
	require.Error(t, err)
	require.Empty(t, msg)
}

func TestBot_MessageHandler_ShouldHandleNilMessage(t *testing.T) {
	tgbot, _, _ := setupTestBot(t)
	
	update := tgbotapi.Update{}
	
	msg, err := tgbot.MessageHandler(context.TODO(), update)
	require.NoError(t, err)
	require.Empty(t, msg)
}

func TestBot_MessageHandler_ShouldHandleCommandWithNilChat(t *testing.T) {
	tgbot, _, _ := setupTestBot(t)
	
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Entities: &[]tgbotapi.MessageEntity{
				{
					Offset: 0,
					Type:   "bot_command",
				},
			},
		},
	}
	
	msg, err := tgbot.MessageHandler(context.TODO(), update)
	require.NoError(t, err)
	require.Empty(t, msg)
}

func TestBot_MessageHandler_ShouldHandleNonCommandMessage(t *testing.T) {
	tgbot, chatID, _ := setupTestBot(t)
	
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: chatID},
		},
	}
	
	msg, err := tgbot.MessageHandler(context.TODO(), update)
	require.NoError(t, err)
	require.Equal(t, "I don't know that command", msg.Text)
}

func TestBot_MessageHandler_ShouldHandleUnknownCommand(t *testing.T) {
	tgbot, chatID, _ := setupTestBot(t)
	
	update := createUpdate(0, chatID, "/unknown", 8)
	
	msg, err := tgbot.MessageHandler(context.TODO(), update)
	require.NoError(t, err)
	require.Equal(t, "I don't know that command", msg.Text)
}

func TestBot_MessageHandler_ShouldHandleHelpCommand(t *testing.T) {
	tgbot, chatID, _ := setupTestBot(t)
	
	update := createUpdate(0, chatID, "/help", 5)
	
	msg, err := tgbot.MessageHandler(context.TODO(), update)
	require.NoError(t, err)
	
	want := `🤖 I'm a bot for reading posts:

Commands:

/top - Top posts

/random - Read random post

/source - Change source:

  1. Lesswrong.ru (default)
  2. Slate Star Codex
  3. Astral Codex Ten
  4. Lesswrong.com

/help - Help`
	require.Equal(t, want, msg.Text)
}

func TestBot_MessageHandler_ShouldShowCurrentSource(t *testing.T) {
	tgbot, chatID, userID := setupTestBot(t)
	
	update := createUpdate(userID, chatID, "/source", 7)
	
	msg, err := tgbot.MessageHandler(context.TODO(), update)
	require.NoError(t, err)
	require.Equal(t, "Current source is https://lesswrong.ru", msg.Text)
}

func TestBot_MessageHandler_ShouldNotChangeInvalidSource(t *testing.T) {
	tgbot, chatID, userID := setupTestBot(t)
	
	update := createUpdate(userID, chatID, "/source invalid", 7)
	
	msg, err := tgbot.MessageHandler(context.TODO(), update)
	require.NoError(t, err)
	require.Equal(t, "New source is invalid. Current source is https://lesswrong.ru", msg.Text)
}

func TestBot_MessageHandler_ShouldChangeSourceToSlateStarCodex(t *testing.T) {
	tgbot, chatID, userID := setupTestBot(t)
	
	update := createUpdate(userID, chatID, "/source 2", 7)
	
	msg, err := tgbot.MessageHandler(context.TODO(), update)
	require.NoError(t, err)
	require.Equal(t, "Changed source to https://slatestarcodex.com", msg.Text)
}

func TestBot_MessageHandler_ShouldGetTopPostsFromSlateStarCodex(t *testing.T) {
	tgbot, chatID, userID := setupTestBot(t)
	
	// First set source to SlateStarCodex
	sourceUpdate := createUpdate(userID, chatID, "/source 2", 7)
	_, err := tgbot.MessageHandler(context.TODO(), sourceUpdate)
	require.NoError(t, err)
	
	// Then get top posts
	update := createUpdate(userID, chatID, "/top", 4)
	
	msg, err := tgbot.MessageHandler(context.TODO(), update)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(msg.Text, "🏆 Top posts from https://slatestarcodex.com"))
}

func TestBot_MessageHandler_ShouldGetRandomPostFromSlateStarCodex(t *testing.T) {
	tgbot, chatID, userID := setupTestBot(t)
	
	// First set source to SlateStarCodex
	sourceUpdate := createUpdate(userID, chatID, "/source 2", 7)
	_, err := tgbot.MessageHandler(context.TODO(), sourceUpdate)
	require.NoError(t, err)
	
	// Then get random post
	update := createUpdate(userID, chatID, "/random", 7)
	
	msg, err := tgbot.MessageHandler(context.TODO(), update)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(msg.Text, "📝"))
}

func TestBot_MessageHandler_ShouldChangeSourceToAstralCodexTen(t *testing.T) {
	tgbot, chatID, userID := setupTestBot(t)
	
	update := createUpdate(userID, chatID, "/source 3", 7)
	
	msg, err := tgbot.MessageHandler(context.TODO(), update)
	require.NoError(t, err)
	require.Equal(t, "Changed source to https://astralcodexten.substack.com", msg.Text)
}

func TestBot_MessageHandler_ShouldGetTopPostsFromAstralCodexTen(t *testing.T) {
	tgbot, chatID, userID := setupTestBot(t)
	
	// First set source to AstralCodexTen
	sourceUpdate := createUpdate(userID, chatID, "/source 3", 7)
	_, err := tgbot.MessageHandler(context.TODO(), sourceUpdate)
	require.NoError(t, err)
	
	// Then get top posts
	update := createUpdate(userID, chatID, "/top", 4)
	
	msg, err := tgbot.MessageHandler(context.TODO(), update)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(msg.Text, "🏆 Top posts from https://astralcodexten.substack.com"))
}

func TestBot_MessageHandler_ShouldGetRandomPostFromAstralCodexTen(t *testing.T) {
	tgbot, chatID, userID := setupTestBot(t)
	
	// First set source to AstralCodexTen
	sourceUpdate := createUpdate(userID, chatID, "/source 3", 7)
	_, err := tgbot.MessageHandler(context.TODO(), sourceUpdate)
	require.NoError(t, err)
	
	// Then get random post
	update := createUpdate(userID, chatID, "/random", 7)
	
	msg, err := tgbot.MessageHandler(context.TODO(), update)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(msg.Text, "📝"))
}

func TestBot_MessageHandler_ShouldChangeSourceToLessWrongRu(t *testing.T) {
	tgbot, chatID, userID := setupTestBot(t)
	
	update := createUpdate(userID, chatID, "/source 1", 7)
	
	msg, err := tgbot.MessageHandler(context.TODO(), update)
	require.NoError(t, err)
	require.Equal(t, "Changed source to https://lesswrong.ru", msg.Text)
}

func TestBot_MessageHandler_ShouldGetTopPostsFromLessWrongRu(t *testing.T) {
	tgbot, chatID, userID := setupTestBot(t)
	
	// First set source to LessWrong.ru
	sourceUpdate := createUpdate(userID, chatID, "/source 1", 7)
	_, err := tgbot.MessageHandler(context.TODO(), sourceUpdate)
	require.NoError(t, err)
	
	// Then get top posts
	update := createUpdate(userID, chatID, "/top", 4)
	
	msg, err := tgbot.MessageHandler(context.TODO(), update)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(msg.Text, "🏆 Random posts from https://lesswrong.ru"))
}

func TestBot_MessageHandler_ShouldGetRandomPostFromLessWrongRu(t *testing.T) {
	tgbot, chatID, userID := setupTestBot(t)
	
	// First set source to LessWrong.ru
	sourceUpdate := createUpdate(userID, chatID, "/source 1", 7)
	_, err := tgbot.MessageHandler(context.TODO(), sourceUpdate)
	require.NoError(t, err)
	
	// Then get random post
	update := createUpdate(userID, chatID, "/random", 7)
	
	msg, err := tgbot.MessageHandler(context.TODO(), update)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(msg.Text, "📝"))
}

func TestBot_MessageHandler_ShouldChangeSourceToLessWrongCom(t *testing.T) {
	tgbot, chatID, userID := setupTestBot(t)
	
	update := createUpdate(userID, chatID, "/source 4", 7)
	
	msg, err := tgbot.MessageHandler(context.TODO(), update)
	require.NoError(t, err)
	require.Equal(t, "Changed source to https://lesswrong.com", msg.Text)
}

func TestBot_MessageHandler_ShouldGetTopPostsFromLessWrongCom(t *testing.T) {
	tgbot, chatID, userID := setupTestBot(t)
	
	// First set source to LessWrong.com
	sourceUpdate := createUpdate(userID, chatID, "/source 4", 7)
	_, err := tgbot.MessageHandler(context.TODO(), sourceUpdate)
	require.NoError(t, err)
	
	// Then get top posts
	update := createUpdate(userID, chatID, "/top", 4)
	
	msg, err := tgbot.MessageHandler(context.TODO(), update)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(msg.Text, "🏆 Top posts this week from https://lesswrong.com"))
}

func TestBot_MessageHandler_ShouldGetRandomPostFromLessWrongCom(t *testing.T) {
	tgbot, chatID, userID := setupTestBot(t)
	
	// First set source to LessWrong.com
	sourceUpdate := createUpdate(userID, chatID, "/source 4", 7)
	_, err := tgbot.MessageHandler(context.TODO(), sourceUpdate)
	require.NoError(t, err)
	
	// Then get random post
	update := createUpdate(userID, chatID, "/random", 7)
	
	msg, err := tgbot.MessageHandler(context.TODO(), update)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(msg.Text, "📝"))
}
