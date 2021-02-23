package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/gocolly/colly"
)

const (
	DefaultLimit  = 12
	PostMaxLength = 1500
)

// TODO Cache posts in some storage.
var (
	astralPosts []AstralPost
	slatePosts  []SlatePost
)

type (
	AstralPost struct {
		Slug         string `json:"slug"`
		Title        string `json:"title"`
		Subtitle     string `json:"subtitle"`
		CanonicalURL string `json:"canonical_url"`
		BodyHTML     string `json:"body_html"`
		Audience     string `json:"audience"`
	}

	SlatePost struct {
		Title    string
		URL      string
		BodyHTML string
	}
)

func CommandRandom(source Source, mdConverter *md.Converter) (string, error) {
	switch source {
	case SourceSlate:
		return CommandRandomSlate(mdConverter)
	case SourceAstral:
		return CommandRandomAstral(mdConverter)
	default:
		return CommandRandomSlate(mdConverter)
	}
}

func CommandRandomSlate(mdConverter *md.Converter) (string, error) {
	// Load posts for the first time.
	if len(slatePosts) == 0 {
		archiveCollector := colly.NewCollector()

		archiveCollector.OnHTML("a[href][rel=bookmark]", func(e *colly.HTMLElement) {
			slatePosts = append(slatePosts, SlatePost{
				Title: e.Text,
				URL:   e.Attr("href"),
			})
		})

		if err := archiveCollector.Visit("https://slatestarcodex.com/archives/"); err != nil {
			return "", fmt.Errorf("get slatestarcodex archives failed: %w", err)
		}
	}

	if len(slatePosts) == 0 {
		return "", fmt.Errorf("posts not found")
	}

	i := rand.Intn(len(slatePosts))
	post := slatePosts[i]

	postCollector := colly.NewCollector()

	postCollector.OnHTML("div .entry-content", func(e *colly.HTMLElement) {
		post.BodyHTML, _ = e.DOM.Html()
	})

	if err := postCollector.Visit(post.URL); err != nil {
		return "", fmt.Errorf("get slatestarcodex post failed: %w", err)
	}

	markdown, err := mdConverter.ConvertString(post.BodyHTML)
	if err != nil {
		return "", fmt.Errorf("convert html to markdown failed: %w", err)
	}

	// Cut post for preview mode.
	if len(markdown) > PostMaxLength {
		// Convert to runes to properly split between unicode symbols.
		r := []rune(markdown)

		// Truncate after next line end to not break markdown text.
		n := strings.IndexByte(string(r[PostMaxLength:]), '\n')
		if n != -1 {
			markdown = string(r[:PostMaxLength+n])
		} else {
			markdown = string(r[:PostMaxLength])
		}
	}

	return fmt.Sprintf("📝 [%s](%s)\n\n%s", post.Title, post.URL, markdown), nil
}

func CommandRandomAstral(mdConverter *md.Converter) (string, error) {
	// Load posts for the first time.
	if len(astralPosts) == 0 {
		// As substack limits list to 12 posts in one request we fetch all posts using offset.
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

			var newPosts []AstralPost

			if err := json.NewDecoder(archiveResponse.Body).Decode(&newPosts); err != nil {
				log.Println("[ERROR] Unmarshal new posts archive failed: ", err)
				break
			}

			if len(newPosts) == 0 {
				break
			}

			for _, post := range newPosts {
				if post.Audience != "only_paid" {
					astralPosts = append(astralPosts, post)
				}
			}
		}
	}

	if len(astralPosts) == 0 {
		return "", fmt.Errorf("posts not found")
	}

	i := rand.Intn(len(astralPosts))
	post := astralPosts[i]

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

	// Cut post for preview mode.
	if len(markdown) > PostMaxLength {
		// Convert to runes to properly split between unicode symbols.
		r := []rune(markdown)

		// Truncate after next line end to not break markdown text.
		n := strings.IndexByte(string(r[PostMaxLength:]), '\n')
		if n != -1 {
			markdown = string(r[:PostMaxLength+n])
		} else {
			markdown = string(r[:PostMaxLength])
		}
	}

	return fmt.Sprintf("📝 [%s](%s)\n\n%s", post.Title, post.CanonicalURL, markdown), nil
}
