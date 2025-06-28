//go:build integration

package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/stretchr/testify/require"

	"github.com/ndrewnee/lesswrong-bot/bot/mocks"
	"github.com/ndrewnee/lesswrong-bot/models"
)

// Individual random post tests - exact same logic as original TestRandomPost
func setupMockHTTPClient(t *testing.T) *mocks.HTTPClient {
	httpClient := &mocks.HTTPClient{}

	httpClient.On("Get", context.TODO(), "https://astralcodexten.substack.com/api/v1/archive?sort=new&limit=12&offset=0").Return(
		&http.Response{
			Body: func() io.ReadCloser {
				file, err := os.ReadFile("testdata/astral_new_posts.json")
				require.NoError(t, err)
				return io.NopCloser(bytes.NewBuffer(file))
			}(),
		},
		nil,
	)

	httpClient.On("Get", context.TODO(), "https://astralcodexten.substack.com/api/v1/archive?sort=new&limit=12&offset=12").Return(
		&http.Response{
			Body: func() io.ReadCloser {
				return io.NopCloser(bytes.NewBufferString("[]"))
			}(),
		},
		nil,
	)

	httpClient.On("Get", context.TODO(), "https://astralcodexten.substack.com/api/v1/posts/open-thread-160").Return(
		&http.Response{
			Body: func() io.ReadCloser {
				file, err := os.ReadFile("testdata/astral_random_post.json")
				require.NoError(t, err)
				return io.NopCloser(bytes.NewBuffer(file))
			}(),
		},
		nil,
	)

	httpClient.On("Get", context.TODO(), "https://astralcodexten.substack.com/api/v1/posts/coronavirus-links-discussion-open").Return(
		&http.Response{
			Body: func() io.ReadCloser {
				file, err := os.ReadFile("testdata/astral_random_post_invalid_cut.json")
				require.NoError(t, err)
				return io.NopCloser(bytes.NewBuffer(file))
			}(),
		},
		nil,
	)

	httpClient.On("Get", context.TODO(), "https://astralcodexten.substack.com/api/v1/posts/open-thread-159").Return(
		&http.Response{
			Body: func() io.ReadCloser {
				file, err := os.ReadFile("testdata/astral_random_post_link_bug.json")
				require.NoError(t, err)
				return io.NopCloser(bytes.NewBuffer(file))
			}(),
		},
		nil,
	)

	query1 := `{
		posts(input: {terms: {view: "new", limit: 1, meta: null, offset: 0}}) {
			results {
				title
				pageUrl
				htmlBody
			}
		}
	}`

	request1, err := json.Marshal(map[string]string{"query": query1})
	require.NoError(t, err)

	httpClient.On("Post", context.TODO(), "https://www.lesswrong.com/graphql", "application/json", bytes.NewBuffer(request1)).Return(
		&http.Response{
			Body: func() io.ReadCloser {
				file, err := os.ReadFile("testdata/lesswrong_random_post.json")
				require.NoError(t, err)
				return io.NopCloser(bytes.NewBuffer(file))
			}(),
		},
		nil,
	)

	query2 := `{
		posts(input: {terms: {view: "new", limit: 1, meta: null, offset: 1}}) {
			results {
				title
				pageUrl
				htmlBody
			}
		}
	}`

	request2, err := json.Marshal(map[string]string{"query": query2})
	require.NoError(t, err)

	httpClient.On("Post", context.TODO(), "https://www.lesswrong.com/graphql", "application/json", bytes.NewBuffer(request2)).Return(
		&http.Response{
			Body: func() io.ReadCloser {
				file, err := os.ReadFile("testdata/lesswrong_random_post_invalid_domain.json")
				require.NoError(t, err)
				return io.NopCloser(bytes.NewBuffer(file))
			}(),
		},
		nil,
	)

	return httpClient
}

func TestRandomPost_ShouldGetRandomPostFromLessWrongRuWhenSourceNotSet(t *testing.T) {
	const userID = 2
	httpClient := setupMockHTTPClient(t)

	tgbot, err := New(Options{BotAPI: &tgbotapi.BotAPI{}, HTTPClient: httpClient})
	require.NoError(t, err)

	tgbot.randomInt = func(n int) int {
		return 2
	}

	got, err := tgbot.RandomPost(context.TODO(), userID)
	require.NoError(t, err)

	file, err := os.ReadFile("testdata/lesswrong_ru_random_post.md")
	require.NoError(t, err)
	require.Equal(t, string(file), got)
}

func TestRandomPost_ShouldGetRandomPostFromSlateStarCodex(t *testing.T) {
	const userID = 2
	httpClient := setupMockHTTPClient(t)

	tgbot, err := New(Options{BotAPI: &tgbotapi.BotAPI{}, HTTPClient: httpClient})
	require.NoError(t, err)

	tgbot.randomInt = func(n int) int {
		return 0
	}

	key := fmt.Sprintf("source:%d", userID)
	err = tgbot.storage.Set(context.TODO(), key, models.SourceSlate.Value(), 0)
	require.NoError(t, err)

	got, err := tgbot.RandomPost(context.TODO(), userID)
	require.NoError(t, err)

	file, err := os.ReadFile("testdata/slate_random_post.md")
	require.NoError(t, err)
	require.Equal(t, string(file), got)
}

func TestRandomPost_ShouldGetRandomPostFromSlateStarCodexInvalidMarkdownCut(t *testing.T) {
	const userID = 2
	httpClient := setupMockHTTPClient(t)

	tgbot, err := New(Options{BotAPI: &tgbotapi.BotAPI{}, HTTPClient: httpClient})
	require.NoError(t, err)

	tgbot.randomInt = func(n int) int {
		return 563
	}

	key := fmt.Sprintf("source:%d", userID)
	err = tgbot.storage.Set(context.TODO(), key, models.SourceSlate.Value(), 0)
	require.NoError(t, err)

	got, err := tgbot.RandomPost(context.TODO(), userID)
	require.NoError(t, err)

	file, err := os.ReadFile("testdata/slate_random_post_invalid_cut.md")
	require.NoError(t, err)
	require.Equal(t, string(file), got)
}

func TestRandomPost_ShouldGetRandomPostFromSlateStarCodexImageFix(t *testing.T) {
	const userID = 2
	httpClient := setupMockHTTPClient(t)

	tgbot, err := New(Options{BotAPI: &tgbotapi.BotAPI{}, HTTPClient: httpClient})
	require.NoError(t, err)

	tgbot.randomInt = func(n int) int {
		return 191
	}

	key := fmt.Sprintf("source:%d", userID)
	err = tgbot.storage.Set(context.TODO(), key, models.SourceSlate.Value(), 0)
	require.NoError(t, err)

	got, err := tgbot.RandomPost(context.TODO(), userID)
	require.NoError(t, err)

	file, err := os.ReadFile("testdata/slate_random_post_image_fix.md")
	require.NoError(t, err)
	require.Equal(t, string(file), got)
}

func TestRandomPost_ShouldGetRandomPostFromAstralCodexTen(t *testing.T) {
	const userID = 2
	httpClient := setupMockHTTPClient(t)

	tgbot, err := New(Options{BotAPI: &tgbotapi.BotAPI{}, HTTPClient: httpClient})
	require.NoError(t, err)

	tgbot.randomInt = func(n int) int {
		return 0
	}

	key := fmt.Sprintf("source:%d", userID)
	err = tgbot.storage.Set(context.TODO(), key, models.SourceAstral.Value(), 0)
	require.NoError(t, err)

	got, err := tgbot.RandomPost(context.TODO(), userID)
	require.NoError(t, err)

	file, err := os.ReadFile("testdata/astral_random_post.md")
	require.NoError(t, err)
	require.Equal(t, string(file), got)
}

func TestRandomPost_ShouldGetRandomPostFromAstralCodexTenInvalidCut(t *testing.T) {
	const userID = 2
	httpClient := setupMockHTTPClient(t)

	tgbot, err := New(Options{BotAPI: &tgbotapi.BotAPI{}, HTTPClient: httpClient})
	require.NoError(t, err)

	tgbot.randomInt = func(n int) int {
		return 1
	}

	key := fmt.Sprintf("source:%d", userID)
	err = tgbot.storage.Set(context.TODO(), key, models.SourceAstral.Value(), 0)
	require.NoError(t, err)

	got, err := tgbot.RandomPost(context.TODO(), userID)
	require.NoError(t, err)

	file, err := os.ReadFile("testdata/astral_random_post_invalid_cut.md")
	require.NoError(t, err)
	require.Equal(t, string(file), got)
}

func TestRandomPost_ShouldGetRandomPostFromAstralCodexTenLinkBug(t *testing.T) {
	const userID = 2
	httpClient := setupMockHTTPClient(t)

	tgbot, err := New(Options{BotAPI: &tgbotapi.BotAPI{}, HTTPClient: httpClient})
	require.NoError(t, err)

	tgbot.randomInt = func(n int) int {
		return 2
	}

	key := fmt.Sprintf("source:%d", userID)
	err = tgbot.storage.Set(context.TODO(), key, models.SourceAstral.Value(), 0)
	require.NoError(t, err)

	got, err := tgbot.RandomPost(context.TODO(), userID)
	require.NoError(t, err)

	file, err := os.ReadFile("testdata/astral_random_post_link_bug.md")
	require.NoError(t, err)
	require.Equal(t, string(file), got)
}

func TestRandomPost_ShouldGetRandomPostFromLessWrongRuInvalidCut(t *testing.T) {
	const userID = 2
	httpClient := setupMockHTTPClient(t)

	tgbot, err := New(Options{BotAPI: &tgbotapi.BotAPI{}, HTTPClient: httpClient})
	require.NoError(t, err)

	tgbot.randomInt = func(n int) int {
		return 1
	}

	key := fmt.Sprintf("source:%d", userID)
	err = tgbot.storage.Set(context.TODO(), key, models.SourceLesswrongRu.Value(), 0)
	require.NoError(t, err)

	got, err := tgbot.RandomPost(context.TODO(), userID)
	require.NoError(t, err)

	file, err := os.ReadFile("testdata/lesswrong_ru_random_post_invalid_cut.md")
	require.NoError(t, err)
	require.Equal(t, string(file), got)
}

func TestRandomPost_ShouldGetRandomPostFromLessWrongCom(t *testing.T) {
	const userID = 2
	httpClient := setupMockHTTPClient(t)

	tgbot, err := New(Options{BotAPI: &tgbotapi.BotAPI{}, HTTPClient: httpClient})
	require.NoError(t, err)

	tgbot.randomInt = func(n int) int {
		return 0
	}

	key := fmt.Sprintf("source:%d", userID)
	err = tgbot.storage.Set(context.TODO(), key, models.SourceLesswrong.Value(), 0)
	require.NoError(t, err)

	got, err := tgbot.RandomPost(context.TODO(), userID)
	require.NoError(t, err)

	file, err := os.ReadFile("testdata/lesswrong_random_post.md")
	require.NoError(t, err)
	require.Equal(t, string(file), got)
}

func TestRandomPost_ShouldGetRandomPostFromLessWrongComInvalidDomain(t *testing.T) {
	const userID = 2
	httpClient := setupMockHTTPClient(t)

	tgbot, err := New(Options{BotAPI: &tgbotapi.BotAPI{}, HTTPClient: httpClient})
	require.NoError(t, err)

	tgbot.randomInt = func(n int) int {
		return 1
	}

	key := fmt.Sprintf("source:%d", userID)
	err = tgbot.storage.Set(context.TODO(), key, models.SourceLesswrong.Value(), 0)
	require.NoError(t, err)

	got, err := tgbot.RandomPost(context.TODO(), userID)
	require.NoError(t, err)

	file, err := os.ReadFile("testdata/lesswrong_random_post_invalid_domain.md")
	require.NoError(t, err)
	require.Equal(t, string(file), got)
}