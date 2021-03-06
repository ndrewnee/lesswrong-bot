package bot

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/ndrewnee/lesswrong-bot/config"
	"github.com/ndrewnee/lesswrong-bot/models"
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
		config     config.Config
		botAPI     *tgbotapi.BotAPI
		httpClient HTTPClient
		storage    Storage
		randomInt  func(n int) int
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
		opts.Config = config.ParseConfig()
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

	return &Bot{
		botAPI:     opts.BotAPI,
		config:     opts.Config,
		httpClient: opts.HTTPClient,
		storage:    opts.Storage,
		randomInt:  opts.RandomInt,
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
		text, _, err := b.ChangeSource(ctx, update.CallbackQuery.From.ID, models.Source(update.CallbackQuery.Data))
		if err != nil {
			log.Printf("[ERROR] Command /source failed: %s", err)
			text = "Change source failed"
		}

		if _, err := b.botAPI.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, "")); err != nil {
			return tgbotapi.Message{}, fmt.Errorf("answer callback failed: %s", err)
		}

		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.DisableWebPagePreview = true

		sent, err := b.botAPI.Send(msg)
		if err != nil {
			return tgbotapi.Message{}, fmt.Errorf("send message failed: %s. Text: \n%s", err, msg.Text)
		}

		return sent, nil
	}

	if update.Message == nil {
		return tgbotapi.Message{}, nil
	}

	if update.Message.From != nil {
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
	}

	if update.Message.Chat == nil {
		return tgbotapi.Message{}, nil
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.DisableWebPagePreview = true

	switch update.Message.Command() {
	case "start", "help":
		msg.ReplyMarkup = mainKeyboard
		msg.Text = MessageHelp
	case "top":
		text, err := b.TopPosts(ctx, update.Message.From.ID)
		if err != nil {
			log.Printf("[ERROR] Command /top failed: %s", err)
			text = "Top posts not found"
		}

		msg.Text = text
	case "random":
		text, err := b.RandomPost(ctx, update.Message.From.ID)
		if err != nil {
			log.Printf("[ERROR] Command /random failed: %s", err)
			text = "Random post not found"
		}

		msg.Text = text
	case "source":
		text, keyboard, err := b.ChangeSource(ctx, update.Message.From.ID, models.Source(update.Message.CommandArguments()))
		if err != nil {
			log.Printf("[ERROR] Command /source failed: %s", err)
			text = "Change source failed"
		}

		msg.Text = text
		msg.ReplyMarkup = keyboard
	default:
		msg.Text = "I don't know that command"
	}

	sent, err := b.botAPI.Send(msg)
	if err != nil {
		errMsg := msg
		errMsg.Text = "Oops, something went wrong!"
		_, _ = b.botAPI.Send(errMsg)

		return tgbotapi.Message{}, fmt.Errorf("send message failed: %s. Text: \n%s", err, msg.Text)
	}

	return sent, nil
}
