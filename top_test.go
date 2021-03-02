package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/ndrewnee/lesswrong-bot/mocks"
	"github.com/stretchr/testify/require"
)

func TestCommandTop(t *testing.T) {
	httpClient := &mocks.HTTPClient{}

	httpClient.On("Get", "https://astralcodexten.substack.com/api/v1/archive?sort=top&limit=10").Return(
		&http.Response{
			Body: func() io.ReadCloser {
				file, err := ioutil.ReadFile("testdata/astral_top_posts.json")
				require.NoError(t, err)

				return ioutil.NopCloser(bytes.NewBuffer(file))
			}(),
		},
		nil,
	)

	query := fmt.Sprintf(`{
		posts(input: {terms: {view: "top", limit: 12, meta: null, after: "%s"}}) {
			results {
				title
				pageUrl
				user {
					displayName
				}
			}
		}
	}`, time.Now().AddDate(0, 0, -7).Format("2006-01-02"))

	request, err := json.Marshal(map[string]string{"query": query})
	require.NoError(t, err)

	httpClient.On("Post", "https://www.lesswrong.com/graphql", "application/json", bytes.NewBuffer(request)).Return(
		&http.Response{
			Body: func() io.ReadCloser {
				file, err := ioutil.ReadFile("testdata/lesswrong_top_posts.json")
				require.NoError(t, err)

				return ioutil.NopCloser(bytes.NewBuffer(file))
			}(),
		},
		nil,
	)

	bot, err := NewBot(Options{HTTPClient: httpClient})
	require.NoError(t, err)

	type args struct {
		randomPost int
		source     Source
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
				randomPost: 0,
			},
			want: func(t *testing.T, got string) {
				file, err := ioutil.ReadFile("testdata/lesswrong_ru_top_posts.md")
				require.NoError(t, err)
				require.Equal(t, string(file), got)
			},
			wantErr: require.NoError,
		},
		{
			name: "Should get top posts from https://slatestarcodex.com",
			args: args{
				source: SourceSlate,
			},
			want: func(t *testing.T, got string) {
				require.Equal(t, MessageTopSlate, got)
			},
			wantErr: require.NoError,
		},
		{
			name: "Should get top posts from https://astralcodexten.substack.com",
			args: args{
				source: SourceAstral,
			},
			want: func(t *testing.T, got string) {
				file, err := ioutil.ReadFile("testdata/astral_top_posts.md")
				require.NoError(t, err)
				require.Equal(t, string(file), got)
			},
			wantErr: require.NoError,
		},
		{
			name: "Should get top posts from https://lesswrong.ru",
			args: args{
				randomPost: 0,
				source:     SourceLesswrongRu,
			},
			want: func(t *testing.T, got string) {
				file, err := ioutil.ReadFile("testdata/lesswrong_ru_top_posts.md")
				require.NoError(t, err)
				require.Equal(t, string(file), got)
			},
			wantErr: require.NoError,
		},
		{
			name: "Should get top posts from https://lesswrong.com",
			args: args{
				source: SourceLesswrong,
			},
			want: func(t *testing.T, got string) {
				file, err := ioutil.ReadFile("testdata/lesswrong_top_posts.md")
				require.NoError(t, err)
				require.Equal(t, string(file), got)
			},
			wantErr: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot.randomInt = func(n int) int {
				return tt.args.randomPost
			}

			got, err := bot.CommandTop(tt.args.source)
			tt.wantErr(t, err)
			tt.want(t, got)
		})
	}
}
