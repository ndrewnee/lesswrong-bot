package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

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

	if err := p.handleResponse(httpResponse, &response); err != nil {
		return models.Post{}, fmt.Errorf("handle lesswrong.com random post response: %s", err)
	}

	if len(response.Data.Posts.Results) == 0 {
		return models.Post{}, fmt.Errorf("lesswrong.com random post not found")
	}

	result := response.Data.Posts.Results[0]

	return result.AsPost(), nil
}

func (p *LessWrongProvider) handleResponse(httpResponse *HTTPResponse, target interface{}) error {
	return json.Unmarshal(httpResponse.Body, target)
}