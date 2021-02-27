package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/ndrewnee/lesswrong-bot/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCommandRandom(t *testing.T) {
	httpClient := &mocks.HTTPClient{}

	httpClient.On("Get", "https://astralcodexten.substack.com/api/v1/archive?sort=new&limit=12&offset=0").Return(
		&http.Response{
			Body: func() io.ReadCloser {
				file, err := ioutil.ReadFile("testdata/astral_new_posts.json")
				require.NoError(t, err)

				return ioutil.NopCloser(bytes.NewBuffer(file))
			}(),
		},
		nil,
	)

	httpClient.On("Get", "https://astralcodexten.substack.com/api/v1/archive?sort=new&limit=12&offset=12").Return(
		&http.Response{
			Body: func() io.ReadCloser {
				return ioutil.NopCloser(bytes.NewBufferString("[]"))
			}(),
		},
		nil,
	)

	httpClient.On("Get", mock.MatchedBy(func(uri string) bool {
		return strings.HasPrefix(uri, "https://astralcodexten.substack.com/api/v1/posts/")
	})).Return(
		&http.Response{
			Body: func() io.ReadCloser {
				file, err := ioutil.ReadFile("testdata/astral_random_post.json")
				require.NoError(t, err)

				return ioutil.NopCloser(bytes.NewBuffer(file))
			}(),
		},
		nil,
	)

	bot := NewBot(nil, httpClient, md.NewConverter("", true, nil))

	type args struct {
		source Source
	}

	tests := []struct {
		name    string
		args    args
		want    func(t *testing.T, got string)
		wantErr require.ErrorAssertionFunc
	}{
		{
			name: "Should get random post from https://slatestarcodex.com when source is not set",
			want: func(t *testing.T, got string) {
				// TODO Think about hot to get content of random post.
				require.True(t, strings.HasPrefix(got, "📝"))
			},
			wantErr: require.NoError,
		},
		{
			name: "Should get random post from https://slatestarcodex.com",
			args: args{
				source: SourceSlate,
			},
			want: func(t *testing.T, got string) {
				require.True(t, strings.HasPrefix(got, "📝"))
			},
			wantErr: require.NoError,
		},
		{
			name: "Should get random post from https://astralcodexten.substack.com",
			args: args{
				source: SourceAstral,
			},
			want: func(t *testing.T, got string) {
				file, err := ioutil.ReadFile("testdata/astral_random_want.md")
				require.NoError(t, err)

				require.Equal(t, string(file), got)
			},
			wantErr: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bot.CommandRandom(tt.args.source)
			tt.wantErr(t, err)
			tt.want(t, got)
		})
	}
}
