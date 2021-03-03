package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ndrewnee/lesswrong-bot/bot/mocks"
	"github.com/ndrewnee/lesswrong-bot/models"
)

func TestCommandRandom(t *testing.T) {
	httpClient := &mocks.HTTPClient{}

	httpClient.On("Get", context.TODO(), "https://astralcodexten.substack.com/api/v1/archive?sort=new&limit=12&offset=0").Return(
		&http.Response{
			Body: func() io.ReadCloser {
				file, err := ioutil.ReadFile("testdata/astral_new_posts.json")
				require.NoError(t, err)

				return ioutil.NopCloser(bytes.NewBuffer(file))
			}(),
		},
		nil,
	)

	httpClient.On("Get", context.TODO(), "https://astralcodexten.substack.com/api/v1/archive?sort=new&limit=12&offset=12").Return(
		&http.Response{
			Body: func() io.ReadCloser {
				return ioutil.NopCloser(bytes.NewBufferString("[]"))
			}(),
		},
		nil,
	)

	httpClient.On("Get", context.TODO(), "https://astralcodexten.substack.com/api/v1/posts/open-thread-160").Return(
		&http.Response{
			Body: func() io.ReadCloser {
				file, err := ioutil.ReadFile("testdata/astral_random_post.json")
				require.NoError(t, err)

				return ioutil.NopCloser(bytes.NewBuffer(file))
			}(),
		},
		nil,
	)

	httpClient.On("Get", context.TODO(), "https://astralcodexten.substack.com/api/v1/posts/coronavirus-links-discussion-open").Return(
		&http.Response{
			Body: func() io.ReadCloser {
				file, err := ioutil.ReadFile("testdata/astral_random_post_invalid_cut.json")
				require.NoError(t, err)

				return ioutil.NopCloser(bytes.NewBuffer(file))
			}(),
		},
		nil,
	)

	query := `{
		posts(input: {terms: {view: "new", limit: 1, meta: null, offset: 0}}) {
			results {
				title
				pageUrl
				htmlBody
			}
		}
	}`

	request, err := json.Marshal(map[string]string{"query": query})
	require.NoError(t, err)

	httpClient.On("Post", context.TODO(), "https://www.lesswrong.com/graphql", "application/json", bytes.NewBuffer(request)).Return(
		&http.Response{
			Body: func() io.ReadCloser {
				file, err := ioutil.ReadFile("testdata/lesswrong_random_post.json")
				require.NoError(t, err)

				return ioutil.NopCloser(bytes.NewBuffer(file))
			}(),
		},
		nil,
	)

	tgbot, err := New(Options{HTTPClient: httpClient})
	require.NoError(t, err)

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
			name: "Should get random post from https://lesswrong.ru when source is not set",
			args: args{
				randomPost: 2,
			},
			want: func(t *testing.T, got string) {
				file, err := ioutil.ReadFile("testdata/lesswrong_ru_random_post.md")
				require.NoError(t, err)
				require.Equal(t, string(file), got)
			},
			wantErr: require.NoError,
		},
		{
			name: "Should get random post from https://slatestarcodex.com",
			args: args{
				source: models.SourceSlate,
			},
			want: func(t *testing.T, got string) {
				file, err := ioutil.ReadFile("testdata/slate_random_post.md")
				require.NoError(t, err)
				require.Equal(t, string(file), got)
			},
			wantErr: require.NoError,
		},
		{
			name: "Should get random post from https://slatestarcodex.com (invalid markdown cut)",
			args: args{
				randomPost: 563,
				source:     models.SourceSlate,
			},
			want: func(t *testing.T, got string) {
				file, err := ioutil.ReadFile("testdata/slate_random_post_invalid_cut.md")
				require.NoError(t, err)
				require.Equal(t, string(file), got)
			},
			wantErr: require.NoError,
		},
		{
			name: "Should get random post from https://slatestarcodex.com (image fix)",
			args: args{
				randomPost: 191,
				source:     models.SourceSlate,
			},
			want: func(t *testing.T, got string) {
				file, err := ioutil.ReadFile("testdata/slate_random_post_image_fix.md")
				require.NoError(t, err)
				require.Equal(t, string(file), got)
			},
			wantErr: require.NoError,
		},
		{
			name: "Should get random post from https://astralcodexten.substack.com",
			args: args{
				randomPost: 4,
				source:     models.SourceAstral,
			},
			want: func(t *testing.T, got string) {
				file, err := ioutil.ReadFile("testdata/astral_random_post.md")
				require.NoError(t, err)
				require.Equal(t, string(file), got)
			},
			wantErr: require.NoError,
		},
		{
			name: "Should get random post from https://astralcodexten.substack.com (invalid markdown cut)",
			args: args{
				randomPost: 12,
				source:     models.SourceAstral,
			},
			want: func(t *testing.T, got string) {
				file, err := ioutil.ReadFile("testdata/astral_random_post_invalid_cut.md")
				require.NoError(t, err)
				require.Equal(t, string(file), got)
			},
			wantErr: require.NoError,
		},
		{
			name: "Should get random post from https://lesswrong.ru",
			args: args{
				randomPost: 2,
				source:     models.SourceLesswrongRu,
			},
			want: func(t *testing.T, got string) {
				file, err := ioutil.ReadFile("testdata/lesswrong_ru_random_post.md")
				require.NoError(t, err)
				require.Equal(t, string(file), got)
			},
			wantErr: require.NoError,
		},
		{
			name: "Should get random post from https://lesswrong.ru (invalid cut)",
			args: args{
				randomPost: 279,
				source:     models.SourceLesswrongRu,
			},
			want: func(t *testing.T, got string) {
				file, err := ioutil.ReadFile("testdata/lesswrong_ru_random_post_invalid_cut.md")
				require.NoError(t, err)
				require.Equal(t, string(file), got)
			},
			wantErr: require.NoError,
		},
		{
			name: "Should get random post from https://lesswrong.com",
			args: args{
				randomPost: 0,
				source:     models.SourceLesswrong,
			},
			want: func(t *testing.T, got string) {
				file, err := ioutil.ReadFile("testdata/lesswrong_random_post.md")
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

			got, err := tgbot.CommandRandom(context.TODO(), tt.args.source)
			tt.wantErr(t, err)
			tt.want(t, got)
		})
	}
}
