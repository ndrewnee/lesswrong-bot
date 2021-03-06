package bot

import (
	"context"
	"fmt"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/ndrewnee/lesswrong-bot/models"
	"github.com/stretchr/testify/require"
)

func TestChangeSource(t *testing.T) {
	const userID = 3

	tgbot, err := New(Options{BotAPI: &tgbotapi.BotAPI{}})
	require.NoError(t, err)

	type args struct {
		newSource models.Source
	}

	tests := []struct {
		name    string
		args    args
		want    func(t *testing.T, text string, keyboard interface{})
		wantErr require.ErrorAssertionFunc
	}{
		{
			name: "Should return current source and source keyboard when newSource is empty",
			args: args{
				newSource: "",
			},
			want: func(t *testing.T, text string, keyboard interface{}) {
				require.Equal(t, "Current source is "+models.SourceLesswrongRu.String(), text)
				require.Equal(t, sourceKeyboard, keyboard)

				source, err := tgbot.storage.Get(context.TODO(), fmt.Sprintf("source:%d", userID))
				require.NoError(t, err)
				require.Empty(t, source)
			},
			wantErr: require.NoError,
		},
		{
			name: "Should return current source and source keyboard when newSource is invalid",
			args: args{
				newSource: "invalid",
			},
			want: func(t *testing.T, text string, keyboard interface{}) {
				require.Equal(t, "New source is invalid. Current source is "+models.SourceLesswrongRu.String(), text)
				require.Equal(t, sourceKeyboard, keyboard)

				source, err := tgbot.storage.Get(context.TODO(), fmt.Sprintf("source:%d", userID))
				require.NoError(t, err)
				require.Empty(t, source)
			},
			wantErr: require.NoError,
		},
		{
			name: "Should return new source and empty keyboard when newSource is valid",
			args: args{
				newSource: models.SourceAstral,
			},
			want: func(t *testing.T, text string, keyboard interface{}) {
				require.Equal(t, "Changed source to "+models.SourceAstral.String(), text)
				require.Nil(t, keyboard)

				source, err := tgbot.storage.Get(context.TODO(), fmt.Sprintf("source:%d", userID))
				require.NoError(t, err)
				require.Equal(t, models.SourceAstral.Value(), source)
			},
			wantErr: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, keyboard, err := tgbot.ChangeSource(context.TODO(), userID, tt.args.newSource)
			tt.wantErr(t, err)
			tt.want(t, text, keyboard)
		})
	}
}
