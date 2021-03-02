package main

import (
	"log"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	bot, err := NewBot()
	if err != nil {
		log.Fatal("Init telegram bot failed: ", err)
	}

	updates, err := bot.GetUpdatesChan()
	if err != nil {
		log.Fatal("Get updates chan failed: ", err)
	}

	for update := range updates {
		if _, err := bot.MessageHandler(update); err != nil {
			log.Println("[ERROR] Message not sent: ", err)
		}
	}
}
