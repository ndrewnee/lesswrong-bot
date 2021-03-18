package config

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
	CacheExpire time.Duration
}

func Parse() Config {
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

	timeout, err := time.ParseDuration(os.Getenv("TIMEOUT"))
	if err != nil {
		timeout = 15 * time.Second
	}

	expire, err := time.ParseDuration(os.Getenv("CACHE_EXPIRE"))
	if err != nil {
		expire = 24 * time.Hour
	}

	return Config{
		RedisURL:    redisURL,
		Address:     ":" + strconv.Itoa(port),
		WebhookHost: webhookHost,
		Token:       os.Getenv("TOKEN"),
		Webhook:     os.Getenv("WEBHOOK") == "true",
		Debug:       os.Getenv("DEBUG") == "true",
		Timeout:     timeout,
		CacheExpire: expire,
	}
}
