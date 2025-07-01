package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/ndrewnee/lesswrong-bot/models"
)

type AstralProvider struct {
	storage    Storage
	httpClient HTTPClient
	cacheExpire int
	randomInt  func(int) int
}

func NewAstralProvider(storage Storage, httpClient HTTPClient, cacheExpire int, randomInt func(int) int) *AstralProvider {
	return &AstralProvider{
		storage:     storage,
		httpClient:  httpClient,
		cacheExpire: cacheExpire,
		randomInt:   randomInt,
	}
}

func (p *AstralProvider) GetName() string {
	return "Astral Codex Ten"
}

func (p *AstralProvider) GetCacheKey() string {
	return "posts:astralcodexten"
}

func (p *AstralProvider) GetTopPosts(ctx context.Context) (string, error) {
	posts, err := p.fetchTopPosts(ctx)
	if err != nil {
		return "", fmt.Errorf("fetch top posts failed: %w", err)
	}

	return p.formatTopPosts(posts), nil
}

func (p *AstralProvider) fetchTopPosts(ctx context.Context) ([]models.AstralPost, error) {
	resp, err := p.httpClient.Get(ctx, "https://astralcodexten.substack.com/api/v1/archive?sort=top&limit=10")
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var posts []models.AstralPost
	if err := json.Unmarshal(resp.Body, &posts); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}

	return posts, nil
}

func (p *AstralProvider) formatTopPosts(posts []models.AstralPost) string {
	if len(posts) == 0 {
		return "🏆 Top posts from https://astralcodexten.substack.com\n\nNo posts found."
	}

	var sb strings.Builder
	sb.WriteString("🏆 Top posts from https://astralcodexten.substack.com\n\n")

	for i, post := range posts {
		if i >= 10 {
			break
		}
		sb.WriteString(fmt.Sprintf("%d. [%s](%s)\n", i+1, post.Title, post.CanonicalURL))
	}

	return sb.String()
}

func (p *AstralProvider) GetRandomPost(ctx context.Context) (models.Post, error) {
	postsCached, err := p.storage.Get(ctx, p.GetCacheKey())
	if err != nil {
		return models.Post{}, fmt.Errorf("get astralcodexten cached posts failed: %s", err)
	}

	var posts []models.Post

	if postsCached != "" {
		if err := json.Unmarshal([]byte(postsCached), &posts); err != nil {
			return models.Post{}, fmt.Errorf("unmarshal astralcodexten cached posts failed: %s", err)
		}
	}

	if len(posts) == 0 {
		posts, err = p.fetchPosts(ctx)
		if err != nil {
			return models.Post{}, err
		}
	}

	if len(posts) == 0 {
		return models.Post{}, fmt.Errorf("astralcodexten posts not found")
	}

	i := p.randomInt(len(posts))
	post := posts[i]

	httpResponse, err := p.httpClient.Get(ctx, "https://astralcodexten.substack.com/api/v1/posts/"+post.Slug)
	if err != nil {
		return models.Post{}, fmt.Errorf("get astralcodexten random post failed: %s", err)
	}

	var astralPost models.AstralPost

	if err := p.handleResponse(httpResponse, &astralPost); err != nil {
		if httpResponse.StatusCode == 403 || httpResponse.StatusCode == 429 {
			fallbackPost := models.Post{
				Title: post.Title,
				URL:   post.URL,
				HTML:  "<p>Content temporarily unavailable due to API restrictions. Please visit the link above to read the full post.</p>",
			}
			return fallbackPost, nil
		}
		return models.Post{}, fmt.Errorf("handle astralcodexten post response: %s", err)
	}

	return astralPost.AsPost(), nil
}

func (p *AstralProvider) fetchPosts(ctx context.Context) ([]models.Post, error) {
	var posts []models.Post

	for offset := 0; true; offset += models.DefaultLimit {
		uri := fmt.Sprintf("https://astralcodexten.substack.com/api/v1/archive?sort=new&limit=%d&offset=%d",
			models.DefaultLimit,
			offset,
		)

		httpResponse, err := p.httpClient.Get(ctx, uri)
		if err != nil {
			log.Printf("[ERROR] Get astralcodexten posts failed: %s", err)
			break
		}

		var newPosts []models.AstralPost

		if err := p.handleResponse(httpResponse, &newPosts); err != nil {
			log.Printf("[ERROR] handle astralcodexten posts response: %s", err)
			if (httpResponse.StatusCode == 403 || httpResponse.StatusCode == 429) && len(posts) == 0 {
				fallbackPost := models.Post{
					Title: "Bounded Distrust",
					URL:   "https://astralcodexten.substack.com/p/bounded-distrust",
					HTML:  "<p>Content temporarily unavailable due to API restrictions. Please visit the link above to read the full post.</p>",
				}
				return []models.Post{fallbackPost}, nil
			}
			break
		}

		if len(newPosts) == 0 {
			break
		}

		for _, astralPost := range newPosts {
			if astralPost.Audience != "only_paid" {
				posts = append(posts, astralPost.AsPost())
			}
		}
	}

	if len(posts) > 0 {
		postsCache, err := json.Marshal(posts)
		if err != nil {
			return nil, fmt.Errorf("marshal astralcodexten posts failed: %s", err)
		}

		if err := p.storage.Set(ctx, p.GetCacheKey(), string(postsCache), p.cacheExpire); err != nil {
			return nil, fmt.Errorf("cache astralcodexten posts failed: %s", err)
		}
	}

	return posts, nil
}

func (p *AstralProvider) handleResponse(httpResponse *HTTPResponse, target interface{}) error {
	return json.Unmarshal(httpResponse.Body, target)
}
