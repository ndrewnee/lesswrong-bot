package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	settings := ParseSettings()

	botAPI, err := tgbotapi.NewBotAPI(settings.Token)
	if err != nil {
		log.Fatal("Init telegram bot api failed: ", err)
	}

	botAPI.Debug = settings.Debug

	log.Printf("Authorized on account %s", botAPI.Self.UserName)

	bot := NewBot(botAPI, http.DefaultClient, md.NewConverter("", true, nil), rand.Intn)

	updates, err := bot.GetUpdatesChan(settings)
	if err != nil {
		log.Fatal("Get updates chan failed: ", err)
	}

	for update := range updates {
		if _, err := bot.MessageHandler(update); err != nil {
			log.Println("[ERROR] Message not sent: ", err)
		}
	}
}
