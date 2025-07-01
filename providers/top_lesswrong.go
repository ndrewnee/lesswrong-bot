package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ndrewnee/lesswrong-bot/models"
)

type LessWrongTopProvider struct {
	httpClient HTTPClient
}

func NewLessWrongTopProvider(httpClient HTTPClient) *LessWrongTopProvider {
	return &LessWrongTopProvider{
		httpClient: httpClient,
	}
}

func (p *LessWrongTopProvider) GetName() string {
	return models.SourceLesswrong.Value()
}

func (p *LessWrongTopProvider) GetTopPosts(ctx context.Context) (string, error) {
	posts, err := p.fetchTopPosts(ctx)
	if err != nil {
		return "", fmt.Errorf("fetch top posts failed: %w", err)
	}

	return p.formatTopPosts(posts), nil
}

func (p *LessWrongTopProvider) fetchTopPosts(ctx context.Context) ([]lesswrongPost, error) {
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
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response lesswrongResponse
	if err := json.Unmarshal(resp.Body, &response); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}

	return response.Data.Posts.Results, nil
}

func (p *LessWrongTopProvider) formatTopPosts(posts []lesswrongPost) string {
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

type lesswrongPost struct {
	Title   string `json:"title"`
	PageURL string `json:"pageUrl"`
}

type lesswrongResponse struct {
	Data struct {
		Posts struct {
			Results []lesswrongPost `json:"results"`
		} `json:"posts"`
	} `json:"data"`
}