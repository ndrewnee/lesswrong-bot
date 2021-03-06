package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gocolly/colly"

	"github.com/ndrewnee/lesswrong-bot/models"
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

func (b *Bot) TopPosts(ctx context.Context, userID int) (string, error) {
	key := fmt.Sprintf("source:%d", userID)

	source, err := b.storage.Get(ctx, key)
	if err != nil {
		log.Printf("[ERROR] Get source failed: %s, key: %s", err, key)
	}

	switch models.Source(source) {
	case models.SourceLesswrongRu:
		return b.topLesswrongRu(ctx)
	case models.SourceSlate:
		return MessageTopSlate, nil
	case models.SourceAstral:
		return b.topAstral(ctx)
	case models.SourceLesswrong:
		return b.topLesswrong(ctx)
	default:
		return b.topLesswrongRu(ctx)
	}
}

func (b *Bot) topAstral(ctx context.Context) (string, error) {
	httpResponse, err := b.httpClient.Get(ctx, "https://astralcodexten.substack.com/api/v1/archive?sort=top&limit=10")
	if err != nil {
		return "", fmt.Errorf("get astralcodexten posts failed: %s", err)
	}

	defer httpResponse.Body.Close()

	var topPosts []models.AstralPost

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

func (b *Bot) topLesswrongRu(ctx context.Context) (string, error) {
	postsCached, err := b.storage.Get(ctx, "posts:lesswrong.ru")
	if err != nil {
		return "", fmt.Errorf("get lesswrong.ru cached posts failed: %s", err)
	}

	var posts []models.Post

	if postsCached != "" {
		if err := json.Unmarshal([]byte(postsCached), &posts); err != nil {
			return "", fmt.Errorf("unmarshal lesswrong.ru cached posts failed: %s", err)
		}
	}

	// Load posts for the first time.
	if len(posts) == 0 {
		postsCollector := colly.NewCollector()

		postsCollector.OnHTML("li.leaf.menu-depth-3,li.leaf.menu-depth-4", func(e *colly.HTMLElement) {
			posts = append(posts, models.Post{
				Title: e.Text,
				URL:   e.Request.AbsoluteURL(e.ChildAttr("a", "href")),
			})
		})

		if err := postsCollector.Visit("https://lesswrong.ru/w"); err != nil {
			return "", fmt.Errorf("get lesswrong.ru posts failed: %s", err)
		}

		postsCache, err := json.Marshal(posts)
		if err != nil {
			return "", fmt.Errorf("marshal lesswrong.ru posts failed: %s", err)
		}

		if err := b.storage.Set(ctx, "posts:lesswrong.ru", string(postsCache), b.config.CacheExpire); err != nil {
			return "", fmt.Errorf("cache lesswrong.ru posts failed: %s", err)
		}
	}

	if len(posts) == 0 {
		return "", fmt.Errorf("lesswrong.ru posts not found")
	}

	text := bytes.NewBufferString("🏆 Random posts from https://lesswrong.ru\n\n")

	// As lesswrong.ru doesn't have page with top posts return random posts instead.
	for i := 0; i < models.DefaultLimit; i++ {
		n := b.randomInt(len(posts))
		post := posts[n]

		text.WriteString(fmt.Sprintf("%d. [%s](%s)\n\n", i+1, post.Title, post.URL))
	}

	return text.String(), nil
}

func (b *Bot) topLesswrong(ctx context.Context) (string, error) {
	query := fmt.Sprintf(`{
		posts(input: {terms: {view: "top", limit: 12, meta: null, after: "%s"}}) {
			results {
				title
				pageUrl
				user {
					displayName
				}
			}
		}
	}`, time.Now().AddDate(0, 0, -7).Format("2006-01-02"))

	body, err := json.Marshal(map[string]string{"query": query})
	if err != nil {
		return "", fmt.Errorf("marshal request for lesswrong.com top posts failed: %s", err)
	}

	httpResponse, err := b.httpClient.Post(ctx, "https://www.lesswrong.com/graphql", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("get lesswrong.com top posts failed: %s", err)
	}

	defer httpResponse.Body.Close()

	var response models.LesswrongResponse

	if err := json.NewDecoder(httpResponse.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("unmarshal lesswrong.com top posts failed: %s", err)
	}

	text := bytes.NewBufferString("🏆 Top posts this week from https://lesswrong.com:\n\n")

	for i, post := range response.Data.Posts.Results {
		text.WriteString(fmt.Sprintf("%d. [%s](%s) (%s)\n\n", i+1, post.Title, post.PageURL, post.User.DisplayName))
	}

	return text.String(), nil
}
