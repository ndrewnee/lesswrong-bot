package main

import (
	"os"
	"strconv"
)

type Settings struct {
	RedisURL    string
	Address     string
	Token       string
	WebhookHost string
	Webhook     bool
	Debug       bool
}

func ParseSettings() Settings {
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		port = 9999
	}

	webhookHost := os.Getenv("WEBHOOK_HOST")
	if webhookHost == "" {
		webhookHost = "https://lesswrong-bot.herokuapp.com"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/1"
	}

	return Settings{
		RedisURL:    redisURL,
		Address:     ":" + strconv.Itoa(port),
		WebhookHost: webhookHost,
		Token:       os.Getenv("TOKEN"),
		Webhook:     os.Getenv("WEBHOOK") == "true",
		Debug:       os.Getenv("DEBUG") == "true",
	}
}
