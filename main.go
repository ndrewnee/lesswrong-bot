package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	DefaultLimit  = 12
	BodyMaxLength = 800
)

const (
	MessageHelp = `🤖 I'm a bot for reading articles from https://astralcodexten.substack.com

Commands:
	
/random - Read random article
/help - Help`

	MessageRandom = `📝 %s

➜ %s

%s`
)

var articles []Article

type Article struct {
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	CanonicalURL string `json:"canonical_url"`
	BodyHTML     string `json:"body_html"`
	Audience     string `json:"audience"`
}

func main() {
	rand.Seed(time.Now().UnixNano())
	mdConverter := md.NewConverter("", true, nil)

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	if os.Getenv("DEBUG") == "true" {
		bot.Debug = true
	}

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

		if !update.Message.IsCommand() {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.DisableWebPagePreview = true

		switch update.Message.Command() {
		case "help":
			msg.Text = MessageHelp
		case "random":
			msg.Text = "Article not found"

			if len(articles) == 0 {
				for offset := 0; true; offset += DefaultLimit {
					uri := fmt.Sprintf("https://astralcodexten.substack.com/api/v1/archive?sort=new&limit=%v&offset=%v",
						DefaultLimit,
						offset,
					)

					archiveResponse, err := http.Get(uri)
					if err != nil {
						log.Println("[ERROR] Get articles archive failed", err)
						break
					}

					var newArticles []Article

					if err := json.NewDecoder(archiveResponse.Body).Decode(&newArticles); err != nil {
						log.Println("[ERROR] Unmarshal articles archive failed", err)
						break
					}

					if len(newArticles) == 0 {
						break
					}

					for _, article := range newArticles {
						if article.Audience == "everyone" {
							articles = append(articles, article)
						}
					}
				}
			}

			if len(articles) == 0 {
				break
			}

			i := rand.Intn(len(articles))
			article := articles[i]

			articleResponse, err := http.Get("https://astralcodexten.substack.com/api/v1/posts/" + article.Slug)
			if err != nil {
				log.Println("[ERROR] Get article from server failed: ", err)
				break
			}

			if err := json.NewDecoder(articleResponse.Body).Decode(&article); err != nil {
				log.Println("[ERROR] Unmarshal article failed: ", err)
				break
			}

			markdown, err := mdConverter.ConvertString(article.BodyHTML)
			if err != nil {
				log.Println("[ERROR] Convert html to markdown failed: ", err)
				break
			}

			if len(markdown) > BodyMaxLength {
				r := []rune(markdown)

				n := strings.IndexByte(string(r[BodyMaxLength:]), '\n')
				if n != -1 {
					markdown = string(r[:BodyMaxLength+n+1])
				} else {
					markdown = string(r[:BodyMaxLength])
				}
			}

			msg.Text = fmt.Sprintf(MessageRandom, article.Title, article.CanonicalURL, markdown)
		default:
			msg.Text = "I don't know that command"
		}

		if _, err := bot.Send(msg); err != nil {
			log.Println("[ERROR] Send message failed: ", err)
		}
	}
}
