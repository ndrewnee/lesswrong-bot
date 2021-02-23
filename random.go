package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
)

const (
	DefaultLimit  = 12
	PostMaxLength = 800
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

func commandRandom(mdConverter *md.Converter) (string, error) {
	if len(posts) == 0 {
		for offset := 0; true; offset += DefaultLimit {
			uri := fmt.Sprintf("https://astralcodexten.substack.com/api/v1/archive?sort=new&limit=%v&offset=%v",
				DefaultLimit,
				offset,
			)

			archiveResponse, err := http.Get(uri)
			if err != nil {
				log.Println("[ERROR] Get posts archive failed: ", err)
				break
			}

			var newPosts []Post

			if err := json.NewDecoder(archiveResponse.Body).Decode(&newPosts); err != nil {
				log.Println("[ERROR] Unmarshal new posts archive failed: ", err)
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
		return "", fmt.Errorf("posts not found")
	}

	i := rand.Intn(len(posts))
	post := posts[i]

	postResponse, err := http.Get("https://astralcodexten.substack.com/api/v1/posts/" + post.Slug)
	if err != nil {
		return "", fmt.Errorf("get post from server failed: %w", err)
	}

	if err := json.NewDecoder(postResponse.Body).Decode(&post); err != nil {
		return "", fmt.Errorf("unmarshal post failed: %w", err)
	}

	markdown, err := mdConverter.ConvertString(post.BodyHTML)
	if err != nil {
		return "", fmt.Errorf("convert html to markdown failed: %w", err)
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

	return fmt.Sprintf("📝 [%s](%s)\n\n%s", post.Title, post.CanonicalURL, markdown), nil
}
