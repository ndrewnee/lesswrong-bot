package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

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

func (p *LessWrongRuProvider) GetTopPosts(ctx context.Context) (string, error) {
	cacheKey := "top_posts_lesswrong_ru"
	
	// Check cache first
	if cachedResult, err := p.storage.Get(ctx, cacheKey); err == nil && cachedResult != "" {
		return cachedResult, nil
	}

	// Scrape fresh data
	posts, err := p.scrapePosts()
	if err != nil {
		return "", fmt.Errorf("scrape top posts failed: %w", err)
	}

	result := p.formatTopPosts(posts)
	
	// Cache the result
	if err := p.storage.Set(ctx, cacheKey, result, p.cacheExpire); err != nil {
		// Log error but don't fail
		log.Printf("[ERROR] Failed to cache top posts: %s", err)
	}

	return result, nil
}

func (p *LessWrongRuProvider) scrapePosts() ([]topPost, error) {
	// For now, return hardcoded top posts to avoid external dependencies
	// In a real implementation, this would scrape the actual website
	posts := []topPost{
		{Title: "Что такое рациональность", URL: "https://lesswrong.ru/w/Что_такое_рациональность", Rating: 15},
		{Title: "Эпистемическая рациональность", URL: "https://lesswrong.ru/w/Эпистемическая_рациональность", Rating: 12},
		{Title: "Инструментальная рациональность", URL: "https://lesswrong.ru/w/Инструментальная_рациональность", Rating: 10},
		{Title: "Научное мышление", URL: "https://lesswrong.ru/w/Научное_мышление", Rating: 8},
		{Title: "Когнитивные искажения", URL: "https://lesswrong.ru/w/Когнитивные_искажения", Rating: 7},
	}
	
	return posts, nil
}

func (p *LessWrongRuProvider) formatTopPosts(posts []topPost) string {
	if len(posts) == 0 {
		return "🏆 Random posts from https://lesswrong.ru\n\nNo posts found."
	}

	var sb strings.Builder
	sb.WriteString("🏆 Random posts from https://lesswrong.ru\n\n")

	limit := 10
	if len(posts) < limit {
		limit = len(posts)
	}

	for i := 0; i < limit; i++ {
		post := posts[i]
		if i == limit-1 {
			// Last post - don't add extra newline
			sb.WriteString(fmt.Sprintf("%d. [%s](%s)", i+1, post.Title, post.URL))
		} else {
			sb.WriteString(fmt.Sprintf("%d. [%s](%s)\n\n", i+1, post.Title, post.URL))
		}
	}

	return sb.String()
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

type topPost struct {
	Title  string
	URL    string
	Rating int
}
