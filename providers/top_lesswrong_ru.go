package providers

import (
	"context"
	"fmt"
	"strings"

	"github.com/ndrewnee/lesswrong-bot/models"
)

type LessWrongRuTopProvider struct {
	storage     Storage
	cacheExpire int
}

func NewLessWrongRuTopProvider(storage Storage, cacheExpire int) *LessWrongRuTopProvider {
	return &LessWrongRuTopProvider{
		storage:     storage,
		cacheExpire: cacheExpire,
	}
}

func (p *LessWrongRuTopProvider) GetName() string {
	return models.SourceLesswrongRu.Value()
}

func (p *LessWrongRuTopProvider) GetTopPosts(ctx context.Context) (string, error) {
	cacheKey := "top_posts_lesswrong_ru"
	
	// Check cache first
	if cachedResult, err := p.storage.Get(ctx, cacheKey); err == nil && cachedResult != "" {
		return cachedResult, nil
	}

	// Scrape fresh data
	posts, err := p.scrapePosts(ctx)
	if err != nil {
		return "", fmt.Errorf("scrape top posts failed: %w", err)
	}

	result := p.formatTopPosts(posts)
	
	// Cache the result
	if err := p.storage.Set(ctx, cacheKey, result, p.cacheExpire); err != nil {
		// Log error but don't fail
		// log.Printf("Failed to cache top posts: %v", err)
	}

	return result, nil
}

func (p *LessWrongRuTopProvider) scrapePosts(ctx context.Context) ([]topPost, error) {
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

func (p *LessWrongRuTopProvider) formatTopPosts(posts []topPost) string {
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
		sb.WriteString(fmt.Sprintf("%d. [%s](%s)\n\n", i+1, post.Title, post.URL))
	}

	return sb.String()
}

type topPost struct {
	Title  string
	URL    string
	Rating int
}