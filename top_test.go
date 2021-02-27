package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

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

	bot := NewLesswrongBot(httpClient, nil)

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
			name: "Should get top posts from slatestarcodex when source is not set",
			want: func(t *testing.T, got string) {
				require.Equal(t, MessageTopSlate, got)
			},
			wantErr: require.NoError,
		},
		{
			name: "Should get top posts from slatestarcodex",
			args: args{
				source: SourceSlate,
			},
			want: func(t *testing.T, got string) {
				require.Equal(t, MessageTopSlate, got)
			},
			wantErr: require.NoError,
		},
		{
			name: "Should get top posts from astralcodexten",
			args: args{
				source: SourceAstral,
			},
			want: func(t *testing.T, got string) {
				file, err := ioutil.ReadFile("testdata/astral_top_want.md")
				require.NoError(t, err)

				require.Equal(t, string(file), got)
			},
			wantErr: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bot.CommandTop(tt.args.source)
			tt.wantErr(t, err)
			tt.want(t, got)
		})
	}
}
