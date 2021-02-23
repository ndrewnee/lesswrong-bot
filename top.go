package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/gocolly/colly"
)

func commandTop() (string, error) {
	archiveResponse, err := http.Get("https://astralcodexten.substack.com/api/v1/archive?sort=top&limit=10")
	if err != nil {
		return "", fmt.Errorf("get posts archive failed: %w", err)
	}

	var topPosts []Post

	if err := json.NewDecoder(archiveResponse.Body).Decode(&topPosts); err != nil {
		return "", fmt.Errorf("unmarshal top posts archive failed: %w", err)
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

	return text.String(), nil
}

var slatePostReg = regexp.MustCompile(`https://slatestarcodex.com/\d{4,}/\d{2,}/\d{2,}`)

func commandTopSlate(mdConverter *md.Converter) (string, error) {
	collector := colly.NewCollector()
	text := bytes.NewBufferString("🏆 Top posts\n\n")

	i := 1
	collector.OnHTML("div .entry-content a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		if link == "https://slatestarcodex.com/2015/04/21/universal-love-said-the-cactus-person/" {
			return
		}

		if slatePostReg.MatchString(link) {
			text.WriteString(fmt.Sprintf("%v. [%s](%s)\n\n", i, e.Text, link))
			i++
		}
	})

	if err := collector.Visit("https://slatestarcodex.com/about/"); err != nil {
		return "", fmt.Errorf("get slatestarcodex top failed: %w", err)
	}

	return text.String(), nil
}
