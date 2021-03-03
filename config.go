package main

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	RedisURL    string
	Address     string
	Token       string
	WebhookHost string
	Webhook     bool
	Debug       bool
	Timeout     time.Duration
}

func ParseConfig() Config {
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		port = 9999
	}

	timeout, err := strconv.Atoi(os.Getenv("TIMEOUT"))
	if err != nil {
		timeout = 10
	}

	webhookHost := os.Getenv("WEBHOOK_HOST")
	if webhookHost == "" {
		webhookHost = "https://lesswrong-bot.herokuapp.com"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/1"
	}

	return Config{
		RedisURL:    redisURL,
		Address:     ":" + strconv.Itoa(port),
		WebhookHost: webhookHost,
		Token:       os.Getenv("TOKEN"),
		Webhook:     os.Getenv("WEBHOOK") == "true",
		Debug:       os.Getenv("DEBUG") == "true",
		Timeout:     time.Duration(timeout) * time.Second,
	}
}
