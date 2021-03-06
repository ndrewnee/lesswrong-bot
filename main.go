package main

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/ndrewnee/lesswrong-bot/bot"
	"github.com/ndrewnee/lesswrong-bot/config"
	"github.com/ndrewnee/lesswrong-bot/storage/memory"
	"github.com/ndrewnee/lesswrong-bot/storage/redis"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	config := config.ParseConfig()

	var (
		storage bot.Storage
		err     error
	)

	storage, err = redis.NewStorage(config.RedisURL)
	if err != nil {
		log.Printf("Connect to redis failed, using memory storage instead: %s", err)
		storage = memory.NewStorage()
	}

	tgbot, err := bot.New(bot.Options{Config: config, Storage: storage})
	if err != nil {
		log.Fatal("Init telegram bot failed: ", err)
	}

	updates, err := tgbot.GetUpdatesChan()
	if err != nil {
		log.Fatal("Get updates chan failed: ", err)
	}

	for update := range updates {
		ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)

		if _, err := tgbot.MessageHandler(ctx, update); err != nil {
			log.Printf("[ERROR] Message not sent: %s", err)
		}

		cancel()
	}
}
