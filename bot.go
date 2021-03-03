package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"

	md "github.com/JohannesKaufmann/html-to-markdown"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/ndrewnee/lesswrong-bot/internal/storage/memory"
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

const (
	SourceLesswrongRu Source = "1"
	SourceSlate       Source = "2"
	SourceAstral      Source = "3"
	SourceLesswrong   Source = "4"
)

type Source string

func (s Source) String() string {
	switch s {
	case SourceLesswrongRu:
		return "https://lesswrong.ru"
	case SourceSlate:
		return "https://slatestarcodex.com"
	case SourceAstral:
		return "https://astralcodexten.substack.com"
	case SourceLesswrong:
		return "https://lesswrong.com"
	default:
		return ""
	}
}

func (s Source) Value() string {
	return string(s)
}

func (s Source) IsValid() bool {
	return s.String() != ""
}

type (
	Bot struct {
		settings    Settings
		botAPI      *tgbotapi.BotAPI
		httpClient  HTTPClient
		storage     Storage
		mdConverter *md.Converter
		randomInt   func(n int) int
		cache       Cache
	}

	Options struct {
		Settings    Settings
		BotAPI      *tgbotapi.BotAPI
		HTTPClient  HTTPClient
		Storage     Storage
		MDConverter *md.Converter
		RandomInt   func(n int) int
	}

	HTTPClient interface {
		Get(uri string) (*http.Response, error)
		Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
	}

	Cache struct {
		astralPosts      []Post
		slatePosts       []Post
		lesswrongRuPosts []Post
	}

	Storage interface {
		Get(ctx context.Context, key string) (string, error)
		Set(ctx context.Context, key, value string) error
	}
)

func NewBot(options ...Options) (*Bot, error) {
	var opts Options

	if len(options) > 0 {
		opts = options[0]
	}

	if opts.Settings == (Settings{}) {
		opts.Settings = ParseSettings()
	}

	if opts.BotAPI == nil {
		botAPI, err := tgbotapi.NewBotAPI(opts.Settings.Token)
		if err != nil {
			return nil, err
		}

		botAPI.Debug = opts.Settings.Debug
		opts.BotAPI = botAPI
	}

	log.Printf("Authorized on account %s", opts.BotAPI.Self.UserName)

	if opts.HTTPClient == nil {
		opts.HTTPClient = http.DefaultClient
	}

	if opts.Storage == nil {
		opts.Storage = memory.NewStorage()
	}

	if opts.MDConverter == nil {
		opts.MDConverter = md.NewConverter("", true, nil)
	}

	if opts.RandomInt == nil {
		opts.RandomInt = rand.Intn
	}

	return &Bot{
		botAPI:      opts.BotAPI,
		settings:    opts.Settings,
		httpClient:  opts.HTTPClient,
		storage:     opts.Storage,
		mdConverter: opts.MDConverter,
		randomInt:   opts.RandomInt,
	}, nil
}

func (b *Bot) GetUpdatesChan() (tgbotapi.UpdatesChannel, error) {
	if b.settings.Webhook {
		webhook := tgbotapi.NewWebhook(b.settings.WebhookHost + "/" + b.botAPI.Token)

		if _, err := b.botAPI.SetWebhook(webhook); err != nil {
			return nil, fmt.Errorf("set webhook failed: %s", err)
		}

		info, err := b.botAPI.GetWebhookInfo()
		if err != nil {
			return nil, fmt.Errorf("get webhook info failed: %s", err)
		}

		if info.LastErrorDate != 0 {
			log.Println("[ERROR] Telegram callback failed", info.LastErrorMessage)
		}

		updates := b.botAPI.ListenForWebhook("/" + b.botAPI.Token)

		go func() {
			if err := http.ListenAndServe(b.settings.Address, nil); err != nil {
				log.Println("[ERROR] Listen and serve failed: ", err)
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
	var err error

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
	case "help":
		msg.Text = MessageHelp
	case "top":
		key := fmt.Sprintf("Source:%d", update.Message.From.ID)

		source, err := b.storage.Get(ctx, key)
		if err != nil {
			log.Printf("[ERROR] Get source for user failed: %s", err)
		}

		msg.Text, err = b.CommandTop(Source(source))
		if err != nil {
			log.Println("[ERROR] Command /top failed: ", err)
			msg.Text = "Top posts not found"
		}
	case "random":
		key := fmt.Sprintf("Source:%d", update.Message.From.ID)

		source, err := b.storage.Get(ctx, key)
		if err != nil {
			log.Printf("[ERROR] Get source for user failed: %s", err)
		}

		msg.Text, err = b.CommandRandom(Source(source))
		if err != nil {
			log.Println("[ERROR] Command /random failed: ", err)
			msg.Text = "Random post not found"
		}
	case "source":
		source := Source(update.Message.CommandArguments())
		if !source.IsValid() {
			source = SourceLesswrongRu
		}

		msg.Text = "Changed source to " + source.String()
		key := fmt.Sprintf("Source:%d", update.Message.From.ID)

		if err := b.storage.Set(ctx, key, source.Value()); err != nil {
			log.Printf("[ERROR] Set source for user failed: %s", err)
			msg.Text = fmt.Sprintf("Change source to %s failed", source.String())
		}
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
