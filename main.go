package main

import (
	"bytes"
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
	PostMaxLength = 800
)

const (
	MessageHelp = `🤖 I'm a bot for reading posts from https://astralcodexten.substack.com

Commands:
	
/top - Top posts

/random - Read random post

/help - Help`
)

var posts []Post

type Post struct {
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle"`
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
		case "top":
			archiveResponse, err := http.Get("https://astralcodexten.substack.com/api/v1/archive?sort=top&limit=10")
			if err != nil {
				log.Println("[ERROR] Get posts archive failed", err)
				break
			}

			var topPosts []Post

			if err := json.NewDecoder(archiveResponse.Body).Decode(&topPosts); err != nil {
				log.Println("[ERROR] Unmarshal top posts archive failed", err)
				break
			}

			text := bytes.NewBufferString("🏆 Top posts\n\n")

			for i, post := range topPosts {
				if post.Audience == "only_paid" {
					continue
				}

				text.WriteString(fmt.Sprintf("%v. [%s](%s)\n\n", i+1, post.Title, post.CanonicalURL))

				if post.Subtitle != "" && post.Subtitle != "..." {
					text.WriteString(fmt.Sprintf("    %s\n\n", post.Subtitle))
				}
			}

			msg.Text = text.String()
		case "random":
			msg.Text = "Post not found"

			if len(posts) == 0 {
				for offset := 0; true; offset += DefaultLimit {
					uri := fmt.Sprintf("https://astralcodexten.substack.com/api/v1/archive?sort=new&limit=%v&offset=%v",
						DefaultLimit,
						offset,
					)

					archiveResponse, err := http.Get(uri)
					if err != nil {
						log.Println("[ERROR] Get posts archive failed", err)
						break
					}

					var newPosts []Post

					if err := json.NewDecoder(archiveResponse.Body).Decode(&newPosts); err != nil {
						log.Println("[ERROR] Unmarshal new posts archive failed", err)
						break
					}

					if len(newPosts) == 0 {
						break
					}

					for _, post := range newPosts {
						if post.Audience != "only_paid" {
							posts = append(posts, post)
						}
					}
				}
			}

			if len(posts) == 0 {
				break
			}

			i := rand.Intn(len(posts))
			post := posts[i]

			postResponse, err := http.Get("https://astralcodexten.substack.com/api/v1/posts/" + post.Slug)
			if err != nil {
				log.Println("[ERROR] Get post from server failed: ", err)
				break
			}

			if err := json.NewDecoder(postResponse.Body).Decode(&post); err != nil {
				log.Println("[ERROR] Unmarshal post failed: ", err)
				break
			}

			markdown, err := mdConverter.ConvertString(post.BodyHTML)
			if err != nil {
				log.Println("[ERROR] Convert html to markdown failed: ", err)
				break
			}

			if len(markdown) > PostMaxLength {
				r := []rune(markdown)

				n := strings.IndexByte(string(r[PostMaxLength:]), '\n')
				if n != -1 {
					markdown = string(r[:PostMaxLength+n+1])
				} else {
					markdown = string(r[:PostMaxLength])
				}
			}

			msg.Text = fmt.Sprintf("📝 [%s](%s)\n\n%s", post.Title, post.CanonicalURL, markdown)
		default:
			msg.Text = "I don't know that command"
		}

		if _, err := bot.Send(msg); err != nil {
			log.Println("[ERROR] Send message failed: ", err)
		}
	}
}
