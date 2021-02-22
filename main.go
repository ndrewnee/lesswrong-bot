package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"

	md "github.com/JohannesKaufmann/html-to-markdown"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Article struct {
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	CanonicalURL string `json:"canonical_url"`
	BodyHTML     string `json:"body_html"`
}

func main() {
	converter := md.NewConverter("", true, nil)

	response, err := http.Get("https://astralcodexten.substack.com/api/v1/archive?sort=new&search=&offset=0&limit=12")
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
				msg.Text = "Article not found"

				i := rand.Intn(len(articles))
				if len(articles) <= i {
					break
				}

				response, err := http.Get("https://astralcodexten.substack.com/api/v1/posts/" + articles[i].Slug)
				if err != nil {
					log.Printf("Get article failed: %s", err)
					continue
				}

				var article Article
				if err := json.NewDecoder(response.Body).Decode(&article); err != nil {
					log.Printf("Unmarshal article failed: %s", err)
					continue
				}

				markdown, err := converter.ConvertString(article.BodyHTML)
				if err != nil {
					log.Printf("Convert html to markdown failed: %s", err)
					continue
				}

				msg.Text = fmt.Sprintf("📝 %s\n\n%.1500s...\n\n➜ %s", article.Title, markdown, article.CanonicalURL)
			default:
				msg.Text = "I don't know that command"
			}

			if _, err := bot.Send(msg); err != nil {
				log.Printf("[ERROR] Send message failed: %s", err)
			}
		}
	}
}
