package providers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gocolly/colly"

	"github.com/ndrewnee/lesswrong-bot/models"
)

type LessWrongRuProvider struct {
	storage     Storage
	cacheExpire int
	randomInt   func(int) int
}

func NewLessWrongRuProvider(storage Storage, cacheExpire int, randomInt func(int) int) *LessWrongRuProvider {
	return &LessWrongRuProvider{
		storage:     storage,
		cacheExpire: cacheExpire,
		randomInt:   randomInt,
	}
}

func (p *LessWrongRuProvider) GetName() string {
	return "LessWrong.ru"
}

func (p *LessWrongRuProvider) GetCacheKey() string {
	return "posts:lesswrong.ru"
}

func (p *LessWrongRuProvider) GetRandomPost(ctx context.Context) (models.Post, error) {
	postsCached, err := p.storage.Get(ctx, p.GetCacheKey())
	if err != nil {
		return models.Post{}, fmt.Errorf("get lesswrong.ru cached posts failed: %s", err)
	}

	var posts []models.Post

	if postsCached != "" {
		if err := json.Unmarshal([]byte(postsCached), &posts); err != nil {
			return models.Post{}, fmt.Errorf("unmarshal lesswrong.ru cached posts failed: %s", err)
		}
	}

	if len(posts) == 0 {
		posts, err = p.fetchPosts(ctx)
		if err != nil {
			return models.Post{}, err
		}
	}

	if len(posts) == 0 {
		return models.Post{}, fmt.Errorf("lesswrong.ru posts not found")
	}

	i := p.randomInt(len(posts))
	post := posts[i]

	postCollector := colly.NewCollector()

	postCollector.OnHTML("div.tex2jax", func(e *colly.HTMLElement) {
		post.HTML, _ = e.DOM.Html()
	})

	if err := postCollector.Visit(post.URL); err != nil {
		return models.Post{}, fmt.Errorf("get lesswrong.ru random post failed: %s", err)
	}

	return post, nil
}

func (p *LessWrongRuProvider) fetchPosts(ctx context.Context) ([]models.Post, error) {
	var posts []models.Post

	postsCollector := colly.NewCollector()

	postsCollector.OnHTML("li.leaf.menu-depth-3,li.leaf.menu-depth-4", func(e *colly.HTMLElement) {
		posts = append(posts, models.Post{
			Title: e.Text,
			URL:   e.Request.AbsoluteURL(e.ChildAttr("a", "href")),
		})
	})

	if err := postsCollector.Visit("https://lesswrong.ru/w"); err != nil {
		return nil, fmt.Errorf("get lesswrong.ru posts failed: %s", err)
	}

	postsCache, err := json.Marshal(posts)
	if err != nil {
		return nil, fmt.Errorf("marshal lesswrong.ru posts failed: %s", err)
	}

	if err := p.storage.Set(ctx, p.GetCacheKey(), string(postsCache), p.cacheExpire); err != nil {
		return nil, fmt.Errorf("cache lesswrong.ru posts failed: %s", err)
	}

	return posts, nil
}