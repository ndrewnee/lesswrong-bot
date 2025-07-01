package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ndrewnee/lesswrong-bot/models"
)

type AstralTopProvider struct {
	httpClient HTTPClient
}

func NewAstralTopProvider(httpClient HTTPClient) *AstralTopProvider {
	return &AstralTopProvider{
		httpClient: httpClient,
	}
}

func (p *AstralTopProvider) GetName() string {
	return models.SourceAstral.Value()
}

func (p *AstralTopProvider) GetTopPosts(ctx context.Context) (string, error) {
	posts, err := p.fetchTopPosts(ctx)
	if err != nil {
		return "", fmt.Errorf("fetch top posts failed: %w", err)
	}

	return p.formatTopPosts(posts), nil
}

func (p *AstralTopProvider) fetchTopPosts(ctx context.Context) ([]astralPost, error) {
	resp, err := p.httpClient.Get(ctx, "https://astralcodexten.substack.com/api/v1/archive?sort=top&limit=10")
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var posts []astralPost
	if err := json.Unmarshal(resp.Body, &posts); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}

	return posts, nil
}

func (p *AstralTopProvider) formatTopPosts(posts []astralPost) string {
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

type astralPost struct {
	Title        string `json:"title"`
	CanonicalURL string `json:"canonical_url"`
}