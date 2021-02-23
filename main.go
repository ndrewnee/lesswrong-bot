package main

import (
	"log"
	"math/rand"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	settings := parseSettings()
	mdConverter := md.NewConverter("", true, nil)

	bot, err := tgbotapi.NewBotAPI(settings.Token)
	if err != nil {
		log.Fatal("Init telegram bot api failed: ", err)
	}

	bot.Debug = settings.Debug

	log.Printf("Authorized on account %s", bot.Self.UserName)

	updates, err := getUpdatesChan(bot, settings)
	if err != nil {
		log.Fatal("Get updates chan failed: ", err)
	}

	for update := range updates {
		if err := messageHandler(mdConverter, bot, update); err != nil {
			log.Println("[ERROR] Message not sent: ", err)
		}
	}
}
