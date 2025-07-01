package providers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gocolly/colly"

	"github.com/ndrewnee/lesswrong-bot/models"
)

type SlateProvider struct {
	storage     Storage
	cacheExpire int
	randomInt   func(int) int
}

func NewSlateProvider(storage Storage, cacheExpire int, randomInt func(int) int) *SlateProvider {
	return &SlateProvider{
		storage:     storage,
		cacheExpire: cacheExpire,
		randomInt:   randomInt,
	}
}

func (p *SlateProvider) GetName() string {
	return "Slate Star Codex"
}

func (p *SlateProvider) GetCacheKey() string {
	return "posts:slatestarcodex"
}

func (p *SlateProvider) GetRandomPost(ctx context.Context) (models.Post, error) {
	postsCached, err := p.storage.Get(ctx, p.GetCacheKey())
	if err != nil {
		return models.Post{}, fmt.Errorf("get slatestarcodex cached posts failed: %s", err)
	}

	var posts []models.Post

	if postsCached != "" {
		if err := json.Unmarshal([]byte(postsCached), &posts); err != nil {
			return models.Post{}, fmt.Errorf("unmarshal slatestarcodex cached posts failed: %s", err)
		}
	}

	if len(posts) == 0 {
		posts, err = p.fetchPosts(ctx)
		if err != nil {
			return models.Post{}, err
		}
	}

	if len(posts) == 0 {
		return models.Post{}, fmt.Errorf("slatestarcodex posts not found")
	}

	i := p.randomInt(len(posts))
	post := posts[i]

	postCollector := colly.NewCollector()

	postCollector.OnHTML("div.pjgm-postcontent", func(e *colly.HTMLElement) {
		post.HTML, _ = e.DOM.Html()
	})

	if err := postCollector.Visit(post.URL); err != nil {
		return models.Post{}, fmt.Errorf("get slatestarcodex random post failed: %s", err)
	}

	return post, nil
}

func (p *SlateProvider) fetchPosts(ctx context.Context) ([]models.Post, error) {
	var posts []models.Post

	archivesCollector := colly.NewCollector()

	archivesCollector.OnHTML("a[href][rel=bookmark]", func(e *colly.HTMLElement) {
		posts = append(posts, models.Post{
			Title: e.Text,
			URL:   e.Attr("href"),
		})
	})

	if err := archivesCollector.Visit("https://slatestarcodex.com/archives/"); err != nil {
		return nil, fmt.Errorf("get slatestarcodex posts failed: %s", err)
	}

	postsCache, err := json.Marshal(posts)
	if err != nil {
		return nil, fmt.Errorf("marshal slatestarcodex posts failed: %s", err)
	}

	if err := p.storage.Set(ctx, p.GetCacheKey(), string(postsCache), p.cacheExpire); err != nil {
		return nil, fmt.Errorf("cache slatestarcodex posts failed: %s", err)
	}

	return posts, nil
}