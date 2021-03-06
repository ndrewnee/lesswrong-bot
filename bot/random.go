package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/gocolly/colly"

	"github.com/ndrewnee/lesswrong-bot/models"
)

func (b *Bot) RandomPost(ctx context.Context, userID int) (string, error) {
	key := fmt.Sprintf("source:%d", userID)

	source, err := b.storage.Get(ctx, key)
	if err != nil {
		log.Printf("[ERROR] Get source failed: %s, key: %s", err, key)
	}

	switch models.Source(source) {
	case models.SourceLesswrongRu:
		return b.randomLesswrongRu(ctx)
	case models.SourceSlate:
		return b.randomSlate(ctx)
	case models.SourceAstral:
		return b.randomAstral(ctx)
	case models.SourceLesswrong:
		return b.randomLesswrong(ctx)
	default:
		return b.randomLesswrongRu(ctx)
	}
}

func (b *Bot) randomSlate(ctx context.Context) (string, error) {
	postsCached, err := b.storage.Get(ctx, "posts:slatestarcodex")
	if err != nil {
		return "", fmt.Errorf("get slatestarcodex cached posts failed: %s", err)
	}

	var posts []models.Post

	if postsCached != "" {
		if err := json.Unmarshal([]byte(postsCached), &posts); err != nil {
			return "", fmt.Errorf("unmarshal slatestarcodex cached posts failed: %s", err)
		}
	}

	// Load posts for the first time.
	if len(posts) == 0 {
		archivesCollector := colly.NewCollector()

		archivesCollector.OnHTML("a[href][rel=bookmark]", func(e *colly.HTMLElement) {
			posts = append(posts, models.Post{
				Title: e.Text,
				URL:   e.Attr("href"),
			})
		})

		if err := archivesCollector.Visit("https://slatestarcodex.com/archives/"); err != nil {
			return "", fmt.Errorf("get slatestarcodex posts failed: %s", err)
		}

		postsCache, err := json.Marshal(posts)
		if err != nil {
			return "", fmt.Errorf("marshal slatestarcodex posts failed: %s", err)
		}

		if err := b.storage.Set(ctx, "posts:slatestarcodex", string(postsCache), b.config.CacheExpire); err != nil {
			return "", fmt.Errorf("cache slatestarcodex posts failed: %s", err)
		}
	}

	if len(posts) == 0 {
		return "", fmt.Errorf("slatestarcodex posts not found")
	}

	i := b.randomInt(len(posts))
	post := posts[i]

	postCollector := colly.NewCollector()

	postCollector.OnHTML("div .entry-content", func(e *colly.HTMLElement) {
		post.HTML, _ = e.DOM.Html()
	})

	if err := postCollector.Visit(post.URL); err != nil {
		return "", fmt.Errorf("get slatestarcodex random post failed: %s", err)
	}

	return b.postToMarkdown(post, md.NewConverter(models.DomainSlate, true, nil), false)
}

func (b *Bot) randomAstral(ctx context.Context) (string, error) {
	postsCached, err := b.storage.Get(ctx, "posts:astralcodexten")
	if err != nil {
		return "", fmt.Errorf("get astralcodexten cached posts failed: %s", err)
	}

	var posts []models.Post

	if postsCached != "" {
		if err := json.Unmarshal([]byte(postsCached), &posts); err != nil {
			return "", fmt.Errorf("unmarshal astralcodexten cached posts failed: %s", err)
		}
	}

	// Load posts for the first time.
	if len(posts) == 0 {
		// As substack limits list to 12 posts in one request we fetch all posts using offset.
		for offset := 0; true; offset += models.DefaultLimit {
			uri := fmt.Sprintf("https://astralcodexten.substack.com/api/v1/archive?sort=new&limit=%d&offset=%d",
				models.DefaultLimit,
				offset,
			)

			httpResponse, err := b.httpClient.Get(ctx, uri)
			if err != nil {
				log.Printf("[ERROR] Get astralcodexten posts failed: %s", err)
				break
			}

			var newPosts []models.AstralPost

			if err := json.NewDecoder(httpResponse.Body).Decode(&newPosts); err != nil {
				log.Printf("[ERROR] Unmarshal astralcodexten new posts failed: %s", err)
				httpResponse.Body.Close()
				break
			}

			httpResponse.Body.Close()

			if len(newPosts) == 0 {
				break
			}

			for _, astralPost := range newPosts {
				if astralPost.Audience != "only_paid" {
					posts = append(posts, astralPost.AsPost())
				}
			}
		}

		postsCache, err := json.Marshal(posts)
		if err != nil {
			return "", fmt.Errorf("marshal astralcodexten posts failed: %s", err)
		}

		if err := b.storage.Set(ctx, "posts:astralcodexten", string(postsCache), b.config.CacheExpire); err != nil {
			return "", fmt.Errorf("cache astralcodexten posts failed: %s", err)
		}
	}

	if len(posts) == 0 {
		return "", fmt.Errorf("astralcodexten posts not found")
	}

	i := b.randomInt(len(posts))
	post := posts[i]

	httpResponse, err := b.httpClient.Get(ctx, "https://astralcodexten.substack.com/api/v1/posts/"+post.Slug)
	if err != nil {
		return "", fmt.Errorf("get astralcodexten random post failed: %s", err)
	}

	defer httpResponse.Body.Close()

	var astralPost models.AstralPost

	if err := json.NewDecoder(httpResponse.Body).Decode(&astralPost); err != nil {
		return "", fmt.Errorf("unmarshal astralcodexten post failed: %s", err)
	}

	return b.postToMarkdown(astralPost.AsPost(), md.NewConverter(models.DomainAstral, true, nil), false)
}

func (b *Bot) randomLesswrongRu(ctx context.Context) (string, error) {
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

	i := b.randomInt(len(posts))
	post := posts[i]

	postCollector := colly.NewCollector()

	postCollector.OnHTML("div.tex2jax", func(e *colly.HTMLElement) {
		post.HTML, _ = e.DOM.Html()
	})

	if err := postCollector.Visit(post.URL); err != nil {
		return "", fmt.Errorf("get lesswrong.ru random post failed: %s", err)
	}

	return b.postToMarkdown(post, md.NewConverter(models.DomainLesswrongRu, true, nil), true)
}

func (b *Bot) randomLesswrong(ctx context.Context) (string, error) {
	query := fmt.Sprintf(`{
		posts(input: {terms: {view: "new", limit: 1, meta: null, offset: %d}}) {
			results {
				title
				pageUrl
				htmlBody
			}
		}
	}`, b.randomInt(models.LesswrongPostsMaxCount))

	request, err := json.Marshal(map[string]string{"query": query})
	if err != nil {
		return "", fmt.Errorf("marshal request for lesswrong.com random post failed: %s", err)
	}

	httpResponse, err := b.httpClient.Post(ctx, "https://www.lesswrong.com/graphql", "application/json", bytes.NewBuffer(request))
	if err != nil {
		return "", fmt.Errorf("get lesswrong.com random post failed: %s", err)
	}

	defer httpResponse.Body.Close()

	var response models.LesswrongResponse

	if err := json.NewDecoder(httpResponse.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("unmarshal lesswrong.com random post failed: %s", err)
	}

	if len(response.Data.Posts.Results) == 0 {
		return "", fmt.Errorf("lesswrong.com random post not found")
	}

	result := response.Data.Posts.Results[0]

	return b.postToMarkdown(result.AsPost(), md.NewConverter(models.DomainLesswrong, true, nil), false)
}

func (b *Bot) postToMarkdown(post models.Post, mdConverter *md.Converter, urlWithText bool) (string, error) {
	markdownOrig, err := mdConverter.ConvertString(post.HTML)
	if err != nil {
		return "", fmt.Errorf("convert lesswrong.ru html to markdown failed: %s", err)
	}

	markdown := markdownOrig

	// Cut post for preview mode.
	if len(markdown) > models.PostMaxLength {
		// Convert to runes to properly split between unicode symbols.
		runes := []rune(markdown)
		markdown = string(runes[:models.PostMaxLength])

		// Truncate after next line end to not break markdown text.
		rest := string(runes[models.PostMaxLength:])
		if n := strings.IndexByte(rest, '\n'); n != -1 {
			markdown += rest[:n]
		} else {
			markdown = markdownOrig
		}

		// Stupid hotfixes when markdown was cut in the middle.
		markdown = strings.ReplaceAll(markdown, "* * *", "")
		markdown = strings.ReplaceAll(markdown, "```", "")
	}

	// Stupid hotfixes for some invalid markdowns.
	markdown = strings.ReplaceAll(markdown, "[[", "[")
	markdown = strings.ReplaceAll(markdown, "]]", "]")
	markdown = strings.ReplaceAll(markdown, "![]", "[Image]")
	markdown = strings.ReplaceAll(markdown, "_[", "")
	markdown = strings.ReplaceAll(markdown, "]_", "")

	link := fmt.Sprintf("[%s](%s)", post.Title, post.URL)

	postURL := post.URL
	if urlWithText {
		postURL = link
	}

	return fmt.Sprintf("📝 %s\n\n%s\n\n%s", link, markdown, postURL), nil
}
