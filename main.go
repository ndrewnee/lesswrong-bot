package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Article struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle"`
	Description  string `json:"description"`
	CanonicalURL string `json:"canonical_url"`
}

func main() {
	response, err := http.Get("https://astralcodexten.substack.com/api/v1/archive?sort=top&search=&offset=0&limit=12")
	if err != nil {
		log.Fatal(err)
	}

	var articles []Article
	if err := json.NewDecoder(response.Body).Decode(&articles); err != nil {
		log.Fatal(err)
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		if update.Message.IsCommand() {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

			switch update.Message.Command() {
			case "help":
				msg.Text = "type /random"
			case "random":
				text := "Article not found"

				if i := rand.Intn(len(articles)); i < len(articles) {
					text = articles[i].CanonicalURL
				}

				msg.Text = text
			default:
				msg.Text = "I don't know that command"
			}

			if _, err := bot.Send(msg); err != nil {
				log.Printf("[ERROR] Send message failed: %s", err)
			}
		}
	}
}
