package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/ndrewnee/lesswrong-bot/models"
)

type LessWrongProvider struct {
	httpClient HTTPClient
	randomInt  func(int) int
}

func NewLessWrongProvider(httpClient HTTPClient, randomInt func(int) int) *LessWrongProvider {
	return &LessWrongProvider{
		httpClient: httpClient,
		randomInt:  randomInt,
	}
}

func (p *LessWrongProvider) GetName() string {
	return "LessWrong.com"
}

func (p *LessWrongProvider) GetCacheKey() string {
	return "posts:lesswrong.com"
}

func (p *LessWrongProvider) GetTopPosts(ctx context.Context) (string, error) {
	posts, err := p.fetchTopPosts(ctx)
	if err != nil {
		return "", fmt.Errorf("fetch top posts failed: %w", err)
	}

	return p.formatTopPosts(posts), nil
}

func (p *LessWrongProvider) fetchTopPosts(ctx context.Context) ([]models.LesswrongResult, error) {
	query := `{
		posts(input: {terms: {view: "top", limit: 10, meta: null}}) {
			results {
				title
				pageUrl
			}
		}
	}`

	requestBody, err := json.Marshal(map[string]string{"query": query})
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	resp, err := p.httpClient.Post(ctx, "https://www.lesswrong.com/graphql", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	if resp.StatusCode != 200 {
		log.Printf("[ERROR] lesswrong.com top posts request failed with status %d: %s", resp.StatusCode, string(resp.Body))
		// Return empty posts to trigger fallback in formatTopPosts
		return []models.LesswrongResult{}, nil
	}

	var response models.LesswrongResponse
	if err := json.Unmarshal(resp.Body, &response); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}

	return response.Data.Posts.Results, nil
}

func (p *LessWrongProvider) formatTopPosts(posts []models.LesswrongResult) string {
	if len(posts) == 0 {
		return "🏆 Top posts from https://www.lesswrong.com\n\nNo posts found."
	}

	var sb strings.Builder
	sb.WriteString("🏆 Top posts from https://www.lesswrong.com\n\n")

	for i, post := range posts {
		if i >= 10 {
			break
		}
		sb.WriteString(fmt.Sprintf("%d. [%s](%s)\n", i+1, post.Title, post.PageURL))
	}

	return sb.String()
}

func (p *LessWrongProvider) GetRandomPost(ctx context.Context) (models.Post, error) {
	query := fmt.Sprintf(`{
		posts(input: {terms: {view: "new", limit: 1, meta: null, offset: %d}}) {
			results {
				title
				pageUrl
				htmlBody
			}
		}
	}`, p.randomInt(models.LesswrongPostsMaxCount))

	request, err := json.Marshal(map[string]string{"query": query})
	if err != nil {
		return models.Post{}, fmt.Errorf("marshal request for lesswrong.com random post failed: %s", err)
	}

	httpResponse, err := p.httpClient.Post(ctx, "https://www.lesswrong.com/graphql", "application/json", bytes.NewBuffer(request))
	if err != nil {
		return models.Post{}, fmt.Errorf("get lesswrong.com random post failed: %s", err)
	}

	var response models.LesswrongResponse

	if err := json.Unmarshal(httpResponse.Body, &response); err != nil {
		return models.Post{}, fmt.Errorf("handle lesswrong.com random post response: %s", err)
	}

	if len(response.Data.Posts.Results) == 0 {
		return models.Post{}, fmt.Errorf("lesswrong.com random post not found")
	}

	result := response.Data.Posts.Results[0]

	return result.AsPost(), nil
}
