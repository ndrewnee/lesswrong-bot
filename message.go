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

var userSource = map[int]Source{}

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

func GetUpdatesChan(bot *tgbotapi.BotAPI, settings Settings) (tgbotapi.UpdatesChannel, error) {
	if settings.Webhook {
		webhook := tgbotapi.NewWebhook(settings.WebhookHost + "/" + bot.Token)

		if _, err := bot.SetWebhook(webhook); err != nil {
			return nil, fmt.Errorf("set webhook failed: %w", err)
		}

		info, err := bot.GetWebhookInfo()
		if err != nil {
			return nil, fmt.Errorf("get webhook info failed: %w", err)
		}

		if info.LastErrorDate != 0 {
			log.Println("[ERROR] Telegram callback failed", info.LastErrorMessage)
		}

		updates := bot.ListenForWebhook("/" + bot.Token)

		go func() {
			if err := http.ListenAndServe(settings.Address, nil); err != nil {
				log.Println("[ERROR] Listen and serve failed: ", err)
			}
		}()

		return updates, nil
	}

	response, err := bot.RemoveWebhook()
	if err != nil {
		return nil, fmt.Errorf("removed webhook failed: %w", err)
	}

	if !response.Ok {
		return nil, fmt.Errorf("remove webhook response contains error: %s", response.Description)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		return nil, fmt.Errorf("get updates chan failed: %w", err)
	}

	return updates, nil
}

func MessageHandler(mdConverter *md.Converter, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	var err error

	if update.Message == nil {
		return nil
	}

	log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

	if !update.Message.IsCommand() {
		return nil
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.DisableWebPagePreview = true

	switch update.Message.Command() {
	case "help":
		msg.Text = MessageHelp
	case "top":
		msg.Text, err = CommandTop(userSource[update.Message.From.ID])
		if err != nil {
			log.Println("[ERROR] Command /top failed: ", err)
			msg.Text = "Top posts not found"
		}
	case "random":
		msg.Text, err = CommandRandom(userSource[update.Message.From.ID], mdConverter)
		if err != nil {
			log.Println("[ERROR] Command /random failed: ", err)
			msg.Text = "Random post not found"
		}
	case "source":
		source := Source(update.Message.CommandArguments())

		if source != SourceSlate && source != SourceAstral {
			source = SourceSlate
		}

		userSource[update.Message.From.ID] = source

		msg.Text = "Changed source to " + source.String()
	default:
		msg.Text = "I don't know that command"
	}

	if _, err := bot.Send(msg); err != nil {
		msg.Text = "Oops, something went wrong!"
		_, _ = bot.Send(msg)

		return fmt.Errorf("send message failed: %w", err)
	}

	return nil
}
