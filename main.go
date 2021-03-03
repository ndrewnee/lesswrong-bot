package main

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/ndrewnee/lesswrong-bot/internal/storage/memory"
	"github.com/ndrewnee/lesswrong-bot/internal/storage/redis"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	settings := ParseSettings()

	var (
		storage Storage
		err     error
	)

	storage, err = redis.NewStorage(settings.RedisURL)
	if err != nil {
		log.Printf("Connect to redis failed, using memory storage instead: %s", err)
		storage = memory.NewStorage()
	}

	bot, err := NewBot(Options{Settings: settings, Storage: storage})
	if err != nil {
		log.Fatal("Init telegram bot failed: ", err)
	}

	updates, err := bot.GetUpdatesChan()
	if err != nil {
		log.Fatal("Get updates chan failed: ", err)
	}

	for update := range updates {
		ctx, cancel := context.WithTimeout(context.Background(), settings.Timeout)

		if _, err := bot.MessageHandler(ctx, update); err != nil {
			log.Println("[ERROR] Message not sent: ", err)
		}

		cancel()
	}
}
