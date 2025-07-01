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
	"github.com/ndrewnee/lesswrong-bot/providers"
)

func TestTopPosts(t *testing.T) {
	const userID = 1

	httpClient := &mocks.HTTPClient{}

	httpClient.On("Get", context.TODO(), "https://astralcodexten.substack.com/api/v1/archive?sort=top&limit=10").Return(
		&http.Response{
			StatusCode: 200,
			Body: func() io.ReadCloser {
				file, err := os.ReadFile("testdata/astral_top_posts.json")
				require.NoError(t, err)

				return io.NopCloser(bytes.NewBuffer(file))
			}(),
		},
		nil,
	)

	query := `{
		posts(input: {terms: {view: "top", limit: 10, meta: null}}) {
			results {
				title
				pageUrl
			}
		}
	}`

	request, err := json.Marshal(map[string]string{"query": query})
	require.NoError(t, err)

	httpClient.On("Post", context.TODO(), "https://www.lesswrong.com/graphql", "application/json", bytes.NewBuffer(request)).Return(
		&http.Response{
			StatusCode: 200,
			Body: func() io.ReadCloser {
				file, err := os.ReadFile("testdata/lesswrong_top_posts.json")
				require.NoError(t, err)

				return io.NopCloser(bytes.NewBuffer(file))
			}(),
		},
		nil,
	)

	tgbot, err := New(Options{BotAPI: &tgbotapi.BotAPI{}, HTTPClient: httpClient})
	require.NoError(t, err)
	
	// Update the provider factory to use the same mock HTTP client
	tgbot.providerFactory = providers.NewProviderFactory(
		tgbot.storage,
		httpClient,
		int(tgbot.config.CacheExpire.Seconds()),
		tgbot.randomInt,
	)

	type args struct {
		randomPost int
		source     models.Source
	}

	tests := []struct {
		name    string
		args    args
		want    func(t *testing.T, got string)
		wantErr require.ErrorAssertionFunc
	}{
		{
			name: "Should get top posts from https://lesswrong.ru when source is not set",
			args: args{
				randomPost: 2,
			},
			want: func(t *testing.T, got string) {
				file, err := os.ReadFile("testdata/lesswrong_ru_top_posts.md")
				require.NoError(t, err)
				require.Equal(t, string(file), got)
			},
			wantErr: require.NoError,
		},
		{
			name: "Should get top posts from https://slatestarcodx.com",
			args: args{
				source: models.SourceSlate,
			},
			want: func(t *testing.T, got string) {
				file, err := os.ReadFile("testdata/slate_top_posts.md")
				require.NoError(t, err)
				require.Equal(t, string(file), got)
			},
			wantErr: require.NoError,
		},
		{
			name: "Should get top posts from https://astralcodexten.substack.com",
			args: args{
				source: models.SourceAstral,
			},
			want: func(t *testing.T, got string) {
				file, err := os.ReadFile("testdata/astral_top_posts.md")
				require.NoError(t, err)
				require.Equal(t, string(file), got)
			},
			wantErr: require.NoError,
		},
		{
			name: "Should get top posts from https://lesswrong.ru",
			args: args{
				randomPost: 2,
				source:     models.SourceLesswrongRu,
			},
			want: func(t *testing.T, got string) {
				file, err := os.ReadFile("testdata/lesswrong_ru_top_posts.md")
				require.NoError(t, err)
				require.Equal(t, string(file), got)
			},
			wantErr: require.NoError,
		},
		{
			name: "Should get top posts from https://lesswrong.com",
			args: args{
				source: models.SourceLesswrong,
			},
			want: func(t *testing.T, got string) {
				file, err := os.ReadFile("testdata/lesswrong_top_posts.md")
				require.NoError(t, err)
				require.Equal(t, string(file), got)
			},
			wantErr: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tgbot.randomInt = func(n int) int {
				return tt.args.randomPost
			}

			key := fmt.Sprintf("source:%d", userID)
			err := tgbot.storage.Set(context.TODO(), key, tt.args.source.Value(), 0)
			require.NoError(t, err)

			got, err := tgbot.TopPosts(context.TODO(), userID)
			tt.wantErr(t, err)
			tt.want(t, got)
		})
	}
}
