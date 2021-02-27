package main

import (
	"fmt"
	"log"
	"net/http"

	md "github.com/JohannesKaufmann/html-to-markdown"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	MessageHelp = `🤖 I'm a bot for reading posts from https://slatestarcodex.com (default) and https://astralcodexten.substack.com.

Commands:

/top - Top posts

/random - Read random post

/source - Change source (1 - slatestarcodex, 2 - astralcodexten)

/help - Help`
)

const (
	SourceSlate  Source = "1"
	SourceAstral Source = "2"
)

type Source string

func (s Source) String() string {
	switch s {
	case SourceSlate:
		return "https://slatestarcodex.com"
	case SourceAstral:
		return "https://astralcodexten.substack.com"
	default:
		return "https://slatestarcodex.com"
	}
}

type (
	Bot struct {
		botAPI      *tgbotapi.BotAPI
		httpClient  HTTPClient
		mdConverter *md.Converter
		// Used as cache. Possible race. TODO Lock with mutex.
		userSource  map[int]Source
		astralPosts []AstralPost
		slatePosts  []SlatePost
	}

	HTTPClient interface {
		Get(uri string) (*http.Response, error)
	}
)

func NewBot(
	botAPI *tgbotapi.BotAPI,
	httpClient HTTPClient,
	mdConverter *md.Converter,
) *Bot {
	return &Bot{
		botAPI:      botAPI,
		httpClient:  httpClient,
		mdConverter: mdConverter,
		userSource:  make(map[int]Source),
	}
}

func (b *Bot) GetUpdatesChan(settings Settings) (tgbotapi.UpdatesChannel, error) {
	if settings.Webhook {
		webhook := tgbotapi.NewWebhook(settings.WebhookHost + "/" + b.botAPI.Token)

		if _, err := b.botAPI.SetWebhook(webhook); err != nil {
			return nil, fmt.Errorf("set webhook failed: %w", err)
		}

		info, err := b.botAPI.GetWebhookInfo()
		if err != nil {
			return nil, fmt.Errorf("get webhook info failed: %w", err)
		}

		if info.LastErrorDate != 0 {
			log.Println("[ERROR] Telegram callback failed", info.LastErrorMessage)
		}

		updates := b.botAPI.ListenForWebhook("/" + b.botAPI.Token)

		go func() {
			if err := http.ListenAndServe(settings.Address, nil); err != nil {
				log.Println("[ERROR] Listen and serve failed: ", err)
			}
		}()

		return updates, nil
	}

	response, err := b.botAPI.RemoveWebhook()
	if err != nil {
		return nil, fmt.Errorf("removed webhook failed: %w", err)
	}

	if !response.Ok {
		return nil, fmt.Errorf("remove webhook response contains error: %s", response.Description)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.botAPI.GetUpdatesChan(u)
	if err != nil {
		return nil, fmt.Errorf("get updates chan failed: %w", err)
	}

	return updates, nil
}

func (b *Bot) MessageHandler(update tgbotapi.Update) (tgbotapi.Message, error) {
	var err error

	if update.Message == nil {
		return tgbotapi.Message{}, nil
	}

	if update.Message.From != nil {
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
	}

	if !update.Message.IsCommand() {
		return tgbotapi.Message{}, nil
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
		msg.Text, err = b.CommandTop(b.userSource[update.Message.From.ID])
		if err != nil {
			log.Println("[ERROR] Command /top failed: ", err)
			msg.Text = "Top posts not found"
		}
	case "random":
		msg.Text, err = b.CommandRandom(b.userSource[update.Message.From.ID])
		if err != nil {
			log.Println("[ERROR] Command /random failed: ", err)
			msg.Text = "Random post not found"
		}
	case "source":
		source := Source(update.Message.CommandArguments())

		if source != SourceSlate && source != SourceAstral {
			source = SourceSlate
		}

		b.userSource[update.Message.From.ID] = source

		msg.Text = "Changed source to " + source.String()
	default:
		msg.Text = "I don't know that command"
	}

	sent, err := b.botAPI.Send(msg)
	if err != nil {
		msg.Text = "Oops, something went wrong!"
		_, _ = b.botAPI.Send(msg)

		return tgbotapi.Message{}, fmt.Errorf("send message failed: %w", err)
	}

	return sent, nil
}
