package bot

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/ndrewnee/lesswrong-bot/config"
	"github.com/ndrewnee/lesswrong-bot/models"
	"github.com/ndrewnee/lesswrong-bot/providers"
	"github.com/ndrewnee/lesswrong-bot/storage/memory"
)

const (
	MessageHelp = `🤖 I'm a bot for reading posts:

Commands:

/top - Top posts

/random - Read random post

/source - Change source:

  1. [Lesswrong.ru](https://lesswrong.ru) (default)
  2. [Slate Star Codex](https://slatestarcodex.com)
  3. [Astral Codex Ten](https://astralcodexten.substack.com)
  4. [Lesswrong.com](https://lesswrong.com)

/help - Help`
)

var mainKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("/top"),
		tgbotapi.NewKeyboardButton("/random"),
		tgbotapi.NewKeyboardButton("/source"),
	),
)

type (
	Bot struct {
		config          config.Config
		botAPI          *tgbotapi.BotAPI
		httpClient      HTTPClient
		storage         Storage
		randomInt       func(n int) int
		providerFactory *providers.ProviderFactory
	}

	Options struct {
		Config     config.Config
		BotAPI     *tgbotapi.BotAPI
		HTTPClient HTTPClient
		Storage    Storage
		RandomInt  func(n int) int
	}

	HTTPClient interface {
		Get(ctx context.Context, uri string) (*http.Response, error)
		Post(ctx context.Context, url, contentType string, body io.Reader) (*http.Response, error)
	}

	Storage interface {
		Get(ctx context.Context, key string) (string, error)
		Set(ctx context.Context, key, value string, expire time.Duration) error
	}
)

func New(options ...Options) (*Bot, error) {
	var opts Options

	if len(options) > 0 {
		opts = options[0]
	}

	if opts.Config == (config.Config{}) {
		opts.Config = config.Parse()
	}

	if opts.BotAPI == nil {
		botAPI, err := tgbotapi.NewBotAPI(opts.Config.Token)
		if err != nil {
			return nil, err
		}

		botAPI.Debug = opts.Config.Debug
		opts.BotAPI = botAPI
	}

	log.Printf("Authorized on account %s", opts.BotAPI.Self.UserName)

	if opts.HTTPClient == nil {
		opts.HTTPClient = NewHTTPClient()
	}

	if opts.Storage == nil {
		opts.Storage = memory.NewStorage()
	}

	if opts.RandomInt == nil {
		opts.RandomInt = rand.Intn
	}

	providerFactory := providers.NewProviderFactory(
		opts.Storage,
		opts.HTTPClient,
		int(opts.Config.CacheExpire.Seconds()),
		opts.RandomInt,
	)

	return &Bot{
		botAPI:          opts.BotAPI,
		config:          opts.Config,
		httpClient:      opts.HTTPClient,
		storage:         opts.Storage,
		randomInt:       opts.RandomInt,
		providerFactory: providerFactory,
	}, nil
}

func (b *Bot) GetUpdatesChan() (tgbotapi.UpdatesChannel, error) {
	if b.config.Webhook {
		webhook := tgbotapi.NewWebhook(b.config.WebhookHost + "/" + b.botAPI.Token)

		if _, err := b.botAPI.SetWebhook(webhook); err != nil {
			return nil, fmt.Errorf("set webhook failed: %s", err)
		}

		info, err := b.botAPI.GetWebhookInfo()
		if err != nil {
			return nil, fmt.Errorf("get webhook info failed: %s", err)
		}

		if info.LastErrorDate != 0 {
			log.Printf("[ERROR] Telegram callback failed: %s", info.LastErrorMessage)
		}

		updates := b.botAPI.ListenForWebhook("/" + b.botAPI.Token)

		go func() {
			if err := http.ListenAndServe(b.config.Address, nil); err != nil {
				log.Printf("[ERROR] Listen and serve failed: %s", err)
			}
		}()

		return updates, nil
	}

	response, err := b.botAPI.RemoveWebhook()
	if err != nil {
		return nil, fmt.Errorf("removed webhook failed: %s", err)
	}

	if !response.Ok {
		return nil, fmt.Errorf("remove webhook response contains error: %s", response.Description)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.botAPI.GetUpdatesChan(u)
	if err != nil {
		return nil, fmt.Errorf("get updates chan failed: %s", err)
	}

	return updates, nil
}

func (b *Bot) MessageHandler(ctx context.Context, update tgbotapi.Update) (tgbotapi.Message, error) {
	if update.CallbackQuery != nil {
		return b.handleCallbackQuery(ctx, update.CallbackQuery)
	}

	if update.Message == nil {
		return tgbotapi.Message{}, nil
	}

	return b.handleMessage(ctx, update.Message)
}

func (b *Bot) handleCallbackQuery(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery) (tgbotapi.Message, error) {
	text, _, err := b.ChangeSource(ctx, callbackQuery.From.ID, models.Source(callbackQuery.Data))
	if err != nil {
		text = b.handleCommandError("source", err, "Change source failed")
	}

	if _, err := b.botAPI.AnswerCallbackQuery(tgbotapi.NewCallback(callbackQuery.ID, "")); err != nil {
		return tgbotapi.Message{}, fmt.Errorf("answer callback failed: %s", err)
	}

	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.DisableWebPagePreview = true

	return b.sendMessage(msg)
}

func (b *Bot) handleMessage(ctx context.Context, message *tgbotapi.Message) (tgbotapi.Message, error) {
	if message.From != nil {
		log.Printf("[%s] %s", message.From.UserName, message.Text)
	}

	if message.Chat == nil {
		return tgbotapi.Message{}, nil
	}

	msg := b.createBaseMessage(message.Chat.ID)

	switch message.Command() {
	case "start", "help":
		return b.handleHelpCommand(msg)
	case "top":
		return b.handleTopCommand(ctx, msg, message.From.ID)
	case "random":
		return b.handleRandomCommand(ctx, msg, message.From.ID)
	case "source":
		return b.handleSourceCommand(ctx, msg, message.From.ID, message.CommandArguments())
	default:
		return b.handleUnknownCommand(msg)
	}
}

func (b *Bot) createBaseMessage(chatID int64) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatID, "")
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.DisableWebPagePreview = true
	return msg
}

func (b *Bot) handleHelpCommand(msg tgbotapi.MessageConfig) (tgbotapi.Message, error) {
	msg.ReplyMarkup = mainKeyboard
	msg.Text = MessageHelp
	return b.sendMessage(msg)
}

func (b *Bot) handleTopCommand(ctx context.Context, msg tgbotapi.MessageConfig, userID int) (tgbotapi.Message, error) {
	text, err := b.TopPosts(ctx, userID)
	if err != nil {
		text = b.handleCommandError("top", err, "Top posts not found")
	}
	msg.Text = text
	return b.sendMessage(msg)
}

func (b *Bot) handleRandomCommand(ctx context.Context, msg tgbotapi.MessageConfig, userID int) (tgbotapi.Message, error) {
	text, err := b.RandomPost(ctx, userID)
	if err != nil {
		text = b.handleCommandError("random", err, "Random post not found")
	}
	msg.Text = text
	return b.sendMessage(msg)
}

func (b *Bot) handleSourceCommand(ctx context.Context, msg tgbotapi.MessageConfig, userID int, args string) (tgbotapi.Message, error) {
	text, keyboard, err := b.ChangeSource(ctx, userID, models.Source(args))
	if err != nil {
		text = b.handleCommandError("source", err, "Change source failed")
	}
	msg.Text = text
	msg.ReplyMarkup = keyboard
	return b.sendMessage(msg)
}

func (b *Bot) handleUnknownCommand(msg tgbotapi.MessageConfig) (tgbotapi.Message, error) {
	msg.Text = "I don't know that command"
	return b.sendMessage(msg)
}

func (b *Bot) sendMessage(msg tgbotapi.MessageConfig) (tgbotapi.Message, error) {
	// Validate markdown before sending if ParseMode is set
	if msg.ParseMode == tgbotapi.ModeMarkdown {
		if err := b.validateMarkdown(msg.Text); err != nil {
			log.Printf("[WARN] Invalid markdown detected, sending as plain text: %s", err)
			msg.ParseMode = ""
		}
	}

	sent, err := b.botAPI.Send(msg)
	if err != nil {
		// If markdown parsing failed, try sending as plain text
		if strings.Contains(err.Error(), "can't parse entities") && msg.ParseMode == tgbotapi.ModeMarkdown {
			log.Printf("[WARN] Markdown parsing failed, retrying as plain text")
			msg.ParseMode = ""
			sent, err = b.botAPI.Send(msg)
			if err == nil {
				return sent, nil
			}
		}
		
		errMsg := msg
		errMsg.Text = "Oops, something went wrong!"
		errMsg.ParseMode = ""
		_, _ = b.botAPI.Send(errMsg)
		return tgbotapi.Message{}, fmt.Errorf("send message failed: %s. Text: \n%s", err, msg.Text)
	}
	return sent, nil
}

func (b *Bot) validateMarkdown(text string) error {
	// Basic validation for common markdown issues
	underscoreCount := strings.Count(text, "_")
	asteriskCount := strings.Count(text, "*")
	
	// Check for unmatched underscores (should be even for proper emphasis)
	if underscoreCount%2 != 0 {
		return fmt.Errorf("unmatched underscores detected: %d", underscoreCount)
	}
	
	// Check for unmatched asterisks (should be even for proper bold)
	if asteriskCount%2 != 0 {
		return fmt.Errorf("unmatched asterisks detected: %d", asteriskCount)
	}
	
	// Check for problematic patterns
	if strings.Contains(text, "\\[") && !strings.Contains(text, "\\]") {
		return fmt.Errorf("incomplete escaped bracket sequence")
	}
	
	return nil
}

func (b *Bot) getUserSource(ctx context.Context, userID int) models.Source {
	key := fmt.Sprintf("source:%d", userID)
	source, err := b.storage.Get(ctx, key)
	if err != nil {
		log.Printf("[ERROR] Get source failed: %s, key: %s", err, key)
	}
	
	sourceModel := models.Source(source)
	if !sourceModel.IsValid() {
		return models.SourceLesswrongRu
	}
	return sourceModel
}

func (b *Bot) handleCommandError(command string, err error, fallbackMessage string) string {
	log.Printf("[ERROR] Command /%s failed: %s", command, err)
	return fallbackMessage
}
