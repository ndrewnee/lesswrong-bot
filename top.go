package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gocolly/colly"
)

// As https://slatestarcodex.com top posts won't change anymore it's much more effecient to return hardcoded list.
const MessageTopSlate = `🏆 Top posts from https://slatestarcodex.com

1. [Beware The Man Of One Study](https://slatestarcodex.com/2014/12/12/beware-the-man-of-one-study/)

2. [Meditations on Moloch](https://slatestarcodex.com/2014/07/30/meditations-on-moloch/)

3. [I Can Tolerate Anything Except The Outgroup](https://slatestarcodex.com/2014/09/30/i-can-tolerate-anything-except-the-outgroup/)

4. [Book Review: Albion’s Seed](https://slatestarcodex.com/2016/04/27/book-review-albions-seed/)

5. [Nobody Is Perfect, Everything Is Commensurable](https://slatestarcodex.com/2014/12/19/nobody-is-perfect-everything-is-commensurable/)

6. [The Control Group Is Out Of Control](https://slatestarcodex.com/2014/04/28/the-control-group-is-out-of-control/)

7. [Considerations On Cost Disease](https://slatestarcodex.com/2017/02/09/considerations-on-cost-disease/)

8. [Archipelago And Atomic Communitarianism](https://slatestarcodex.com/2014/06/07/archipelago-and-atomic-communitarianism/)

9. [The Categories Were Made For Man, Not Man For The Categories](https://slatestarcodex.com/2014/11/21/the-categories-were-made-for-man-not-man-for-the-categories/)

10. [Who By Very Slow Decay](https://slatestarcodex.com/2013/07/17/who-by-very-slow-decay/)`

func (b *Bot) CommandTop(source Source) (string, error) {
	switch source {
	case SourceSlate:
		return MessageTopSlate, nil
	case SourceAstral:
		return b.CommandTopAstral()
	case SourceLesswrongRu:
		return b.CommandTopLesswrongRu()
	case SourceLesswrong:
		return b.CommandTopLesswrong()
	default:
		return MessageTopSlate, nil
	}
}

func (b *Bot) CommandTopAstral() (string, error) {
	httpResponse, err := b.httpClient.Get("https://astralcodexten.substack.com/api/v1/archive?sort=top&limit=10")
	if err != nil {
		return "", fmt.Errorf("get astralcodexten posts failed: %s", err)
	}

	defer httpResponse.Body.Close()

	var topPosts []AstralPost

	if err := json.NewDecoder(httpResponse.Body).Decode(&topPosts); err != nil {
		return "", fmt.Errorf("unmarshal astralcodexten top posts failed: %s", err)
	}

	text := bytes.NewBufferString("🏆 Top posts from https://astralcodexten.substack.com\n\n")

	for i, post := range topPosts {
		if post.Audience == "only_paid" {
			continue
		}

		text.WriteString(fmt.Sprintf("%d. [%s](%s)\n\n", i+1, post.Title, post.CanonicalURL))

		if post.Subtitle != "" && post.Subtitle != "..." {
			text.WriteString(fmt.Sprintf("    %s\n\n", post.Subtitle))
		}
	}

	return text.String(), nil
}

func (b *Bot) CommandTopLesswrongRu() (string, error) {
	// Load posts for the first time.
	if len(b.cache.lesswrongRuPosts) == 0 {
		postsCollector := colly.NewCollector()

		postsCollector.OnHTML("ul > li.leaf.menu-depth-4", func(e *colly.HTMLElement) {
			b.cache.lesswrongRuPosts = append(b.cache.lesswrongRuPosts, Post{
				Title: e.Text,
				URL:   e.Request.AbsoluteURL(e.ChildAttr("a", "href")),
			})
		})

		if err := postsCollector.Visit("https://lesswrong.ru/w"); err != nil {
			return "", fmt.Errorf("get lesswrong.ru posts failed: %s", err)
		}
	}

	if len(b.cache.lesswrongRuPosts) == 0 {
		return "", fmt.Errorf("lesswrong.ru posts not found")
	}

	text := bytes.NewBufferString("🏆 Random posts from https://lesswrong.ru\n\n")

	// As lesswrong.ru doesn't have page with top posts return random posts instead.
	for i := 0; i < DefaultLimit; i++ {
		n := b.randomInt(len(b.cache.lesswrongRuPosts))
		post := b.cache.lesswrongRuPosts[n]

		text.WriteString(fmt.Sprintf("%d. [%s](%s)\n\n", i+1, post.Title, post.URL))
	}

	return text.String(), nil
}

func (b *Bot) CommandTopLesswrong() (string, error) {
	weekAgo := time.Now().AddDate(0, 0, -7)

	query := fmt.Sprintf(`{
		posts(input: {terms: {view: "top", limit: 12, after: "%s"}}) {
			results {
				title
				pageUrl
				user {
					displayName
				}
			}
		}
	}`, weekAgo.Format("2006-01-02"))

	request, err := json.Marshal(map[string]string{"query": query})
	if err != nil {
		return "", fmt.Errorf("marshal request for lesswrong.com top posts failed: %s", err)
	}

	httpResponse, err := b.httpClient.Post("https://www.lesswrong.com/graphql", "application/json", bytes.NewBuffer(request))
	if err != nil {
		return "", fmt.Errorf("get lesswrong.com top posts failed: %s", err)
	}

	defer httpResponse.Body.Close()

	var response LesswrongResponse

	if err := json.NewDecoder(httpResponse.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("unmarshal lesswrong.com top posts failed: %s", err)
	}

	text := bytes.NewBufferString("🏆 Top posts this week from https://lesswrong.com:\n\n")

	for i, post := range response.Data.Posts.Results {
		text.WriteString(fmt.Sprintf("%d. [%s](%s) (%s)\n\n", i+1, post.Title, post.PageURL, post.User.DisplayName))
	}

	return text.String(), nil
}
