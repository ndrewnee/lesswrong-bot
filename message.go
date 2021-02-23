package main

import (
	"fmt"
	"log"
	"net/http"

	md "github.com/JohannesKaufmann/html-to-markdown"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	MessageHelp = `🤖 I'm a bot for reading posts from https://astralcodexten.substack.com

Commands:
	
/top - Top posts

/random - Read random post

/help - Help`
)

func getUpdatesChan(bot *tgbotapi.BotAPI, settings Settings) (tgbotapi.UpdatesChannel, error) {
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
			log.Println("[ERROR] Telegram callback faileds", info.LastErrorMessage)
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

func messageHandler(mdConverter *md.Converter, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
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
		msg.Text, err = commandTop()
		if err != nil {
			log.Println("[ERROR] Command /top failed: ", err)
			msg.Text = "Top posts not found"
		}
	case "random":
		msg.Text, err = commandRandom(mdConverter)
		if err != nil {
			log.Println("[ERROR] Command /random failed: ", err)
			msg.Text = "Radnom post not found"
		}
	default:
		msg.Text = "I don't know that command"
	}

	if _, err := bot.Send(msg); err != nil {
		return fmt.Errorf("send message failed: %w", err)
	}

	return nil
}
