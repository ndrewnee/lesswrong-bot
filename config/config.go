package config

import (
	"os"
	"strconv"
	"time"
)

const (
	// Default values for configuration
	DefaultPort         = 9999
	DefaultWebhookHost  = "https://lesswrong-bot.herokuapp.com"
	DefaultRedisURL     = "redis://localhost:6379/1"
	DefaultTimeout      = 15 * time.Second
	DefaultCacheExpire  = 24 * time.Hour
	
	// Application constants
	DefaultPostLimit    = 12
	TopPostsLimit       = 10
	TopPostsWeeklyDays  = 7
	PostMaxLength       = 500
	LesswrongPostsMax   = 2000
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
		port = DefaultPort
	}

	webhookHost := os.Getenv("WEBHOOK_HOST")
	if webhookHost == "" {
		webhookHost = DefaultWebhookHost
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = DefaultRedisURL
	}

	timeout, err := time.ParseDuration(os.Getenv("TIMEOUT"))
	if err != nil {
		timeout = DefaultTimeout
	}

	expire, err := time.ParseDuration(os.Getenv("CACHE_EXPIRE"))
	if err != nil {
		expire = DefaultCacheExpire
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
