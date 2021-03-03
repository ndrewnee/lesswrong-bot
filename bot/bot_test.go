// +build integration

package bot

import (
	"context"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/stretchr/testify/require"

	"github.com/ndrewnee/lesswrong-bot/config"
	"github.com/ndrewnee/lesswrong-bot/storage/redis"
)

func TestBot_GetUpdatesChan(t *testing.T) {
	type args struct {
		config config.Config
	}

	tests := []struct {
		name    string
		args    args
		want    require.ValueAssertionFunc
		wantErr require.ErrorAssertionFunc
	}{
		{
			name: "Shouldn't get webhook chan because webhook host is empty",
			args: args{
				config: config.Config{
					Webhook:     true,
					WebhookHost: "",
				},
			},
			want:    require.Nil,
			wantErr: require.Error,
		},
		{
			name: "Should get webhook chan",
			args: args{
				config: config.Config{
					Webhook:     true,
					WebhookHost: "https://lesswrong-bot.herokuapp.com",
				},
			},
			want:    require.NotNil,
			wantErr: require.NoError,
		},
		{
			name:    "Should get polling chan",
			want:    require.NotNil,
			wantErr: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tgbot, err := New()
			require.NoError(t, err)

			tgbot.config = tt.args.config

			got, err := tgbot.GetUpdatesChan()
			tt.wantErr(t, err)
			tt.want(t, got)
			// To avoid error "Too Many Requests: retry after 1"
			time.Sleep(time.Second)
		})
	}
}

func TestBot_MessageHandler(t *testing.T) {
	chatID, err := strconv.ParseInt(os.Getenv("CHAT_ID"), 10, 64)
	require.NoError(t, err, "Env var CHAT_ID should be set")

	userID, err := strconv.Atoi(os.Getenv("USER_ID"))
	require.NoError(t, err, "Env var USER_ID should be set")

	config := config.ParseConfig()

	storage, err := redis.NewStorage(config.RedisURL)
	require.NoError(t, err, "Connect to redis failed")

	tgbot, err := New(Options{Config: config, Storage: storage})
	require.NoError(t, err)

	type args struct {
		update tgbotapi.Update
	}

	tests := []struct {
		name    string
		args    args
		check   func(t *testing.T, got tgbotapi.Message)
		wantErr require.ErrorAssertionFunc
	}{
		{
			name:    "Should handle nil message",
			wantErr: require.NoError,
		},
		{
			name: "Should handle command with nil chat",
			args: args{
				update: tgbotapi.Update{
					Message: &tgbotapi.Message{
						Entities: &[]tgbotapi.MessageEntity{
							{
								Offset: 0,
								Type:   "bot_command",
							},
						},
					},
				},
			},
			wantErr: require.NoError,
		},
		{
			name: "Should handle non-command message",
			args: args{
				update: tgbotapi.Update{
					Message: &tgbotapi.Message{
						Chat: &tgbotapi.Chat{
							ID: chatID,
						},
					},
				},
			},
			check: func(t *testing.T, got tgbotapi.Message) {
				require.Equal(t, "I don't know that command", got.Text)
			},
			wantErr: require.NoError,
		},
		{
			name: "Should handle unknown command",
			args: args{
				update: tgbotapi.Update{
					Message: &tgbotapi.Message{
						Entities: &[]tgbotapi.MessageEntity{
							{
								Offset: 0,
								Type:   "bot_command",
								Length: 8,
							},
						},
						Chat: &tgbotapi.Chat{
							ID: chatID,
						},
						Text: "/unknown",
					},
				},
			},
			check: func(t *testing.T, got tgbotapi.Message) {
				require.Equal(t, "I don't know that command", got.Text)
			},
			wantErr: require.NoError,
		},
		{
			name: "Should handle command /help",
			args: args{
				update: tgbotapi.Update{
					Message: &tgbotapi.Message{
						Entities: &[]tgbotapi.MessageEntity{
							{
								Offset: 0,
								Type:   "bot_command",
								Length: 5,
							},
						},
						Chat: &tgbotapi.Chat{
							ID: chatID,
						},
						Text: "/help",
					},
				},
			},
			check: func(t *testing.T, got tgbotapi.Message) {
				want := `🤖 I'm a bot for reading posts:

Commands:

/top - Top posts

/random - Read random post

/source - Change source:

  1. Lesswrong.ru (default)
  2. Slate Star Codex
  3. Astral Codex Ten
  4. Lesswrong.com

/help - Help`
				require.Equal(t, want, got.Text)
			},
			wantErr: require.NoError,
		},
		{
			name: "Should change source to https://slatestarcodex.com",
			args: args{
				update: tgbotapi.Update{
					Message: &tgbotapi.Message{
						From: &tgbotapi.User{
							ID: userID,
						},
						Entities: &[]tgbotapi.MessageEntity{
							{
								Offset: 0,
								Type:   "bot_command",
								Length: 7,
							},
						},
						Chat: &tgbotapi.Chat{
							ID: chatID,
						},
						Text: "/source 2",
					},
				},
			},
			check: func(t *testing.T, got tgbotapi.Message) {
				require.Equal(t, "Changed source to https://slatestarcodex.com", got.Text)
			},
			wantErr: require.NoError,
		},
		{
			name: "Should get top posts from https://slatestarcodex.com",
			args: args{
				update: tgbotapi.Update{
					Message: &tgbotapi.Message{
						From: &tgbotapi.User{
							ID: userID,
						},
						Entities: &[]tgbotapi.MessageEntity{
							{
								Offset: 0,
								Type:   "bot_command",
								Length: 4,
							},
						},
						Chat: &tgbotapi.Chat{
							ID: chatID,
						},
						Text: "/top",
					},
				},
			},
			check: func(t *testing.T, got tgbotapi.Message) {
				want := `🏆 Top posts from https://slatestarcodex.com

1. Beware The Man Of One Study

2. Meditations on Moloch

3. I Can Tolerate Anything Except The Outgroup

4. Book Review: Albion’s Seed

5. Nobody Is Perfect, Everything Is Commensurable

6. The Control Group Is Out Of Control

7. Considerations On Cost Disease

8. Archipelago And Atomic Communitarianism

9. The Categories Were Made For Man, Not Man For The Categories

10. Who By Very Slow Decay`
				require.Equal(t, want, got.Text)
			},
			wantErr: require.NoError,
		},
		{
			name: "Should get random post from https://slatestarcodex.com",
			args: args{
				update: tgbotapi.Update{
					Message: &tgbotapi.Message{
						From: &tgbotapi.User{
							ID: userID,
						},
						Entities: &[]tgbotapi.MessageEntity{
							{
								Offset: 0,
								Type:   "bot_command",
								Length: 7,
							},
						},
						Chat: &tgbotapi.Chat{
							ID: chatID,
						},
						Text: "/random",
					},
				},
			},
			check: func(t *testing.T, got tgbotapi.Message) {
				require.True(t, strings.HasPrefix(got.Text, "📝"))
			},
			wantErr: require.NoError,
		},
		{
			name: "Should change source to https://astralcodexten.substack.com",
			args: args{
				update: tgbotapi.Update{
					Message: &tgbotapi.Message{
						From: &tgbotapi.User{
							ID: userID,
						},
						Entities: &[]tgbotapi.MessageEntity{
							{
								Offset: 0,
								Type:   "bot_command",
								Length: 7,
							},
						},
						Chat: &tgbotapi.Chat{
							ID: chatID,
						},
						Text: "/source 3",
					},
				},
			},
			check: func(t *testing.T, got tgbotapi.Message) {
				require.Equal(t, "Changed source to https://astralcodexten.substack.com", got.Text)
			},
			wantErr: require.NoError,
		},
		{
			name: "Should get top posts from https://astralcodexten.substack.com",
			args: args{
				update: tgbotapi.Update{
					Message: &tgbotapi.Message{
						From: &tgbotapi.User{
							ID: userID,
						},
						Entities: &[]tgbotapi.MessageEntity{
							{
								Offset: 0,
								Type:   "bot_command",
								Length: 4,
							},
						},
						Chat: &tgbotapi.Chat{
							ID: chatID,
						},
						Text: "/top",
					},
				},
			},
			check: func(t *testing.T, got tgbotapi.Message) {
				require.True(t, strings.HasPrefix(got.Text, "🏆 Top posts from https://astralcodexten.substack.com"))
			},
			wantErr: require.NoError,
		},
		{
			name: "Should get random post from https://astralcodexten.substack.com",
			args: args{
				update: tgbotapi.Update{
					Message: &tgbotapi.Message{
						From: &tgbotapi.User{
							ID: userID,
						},
						Entities: &[]tgbotapi.MessageEntity{
							{
								Offset: 0,
								Type:   "bot_command",
								Length: 7,
							},
						},
						Chat: &tgbotapi.Chat{
							ID: chatID,
						},
						Text: "/random",
					},
				},
			},
			check: func(t *testing.T, got tgbotapi.Message) {
				require.True(t, strings.HasPrefix(got.Text, "📝"))
			},
			wantErr: require.NoError,
		},
		{
			name: "Should change source to https://lesswrong.ru",
			args: args{
				update: tgbotapi.Update{
					Message: &tgbotapi.Message{
						From: &tgbotapi.User{
							ID: userID,
						},
						Entities: &[]tgbotapi.MessageEntity{
							{
								Offset: 0,
								Type:   "bot_command",
								Length: 7,
							},
						},
						Chat: &tgbotapi.Chat{
							ID: chatID,
						},
						Text: "/source 1",
					},
				},
			},
			check: func(t *testing.T, got tgbotapi.Message) {
				require.Equal(t, "Changed source to https://lesswrong.ru", got.Text)
			},
			wantErr: require.NoError,
		},
		{
			name: "Should get top posts from https://lesswrong.ru",
			args: args{
				update: tgbotapi.Update{
					Message: &tgbotapi.Message{
						From: &tgbotapi.User{
							ID: userID,
						},
						Entities: &[]tgbotapi.MessageEntity{
							{
								Offset: 0,
								Type:   "bot_command",
								Length: 4,
							},
						},
						Chat: &tgbotapi.Chat{
							ID: chatID,
						},
						Text: "/top",
					},
				},
			},
			check: func(t *testing.T, got tgbotapi.Message) {
				require.True(t, strings.HasPrefix(got.Text, "🏆 Random posts from https://lesswrong.ru"))
			},
			wantErr: require.NoError,
		},
		{
			name: "Should get random post from https://lesswrong.ru",
			args: args{
				update: tgbotapi.Update{
					Message: &tgbotapi.Message{
						From: &tgbotapi.User{
							ID: userID,
						},
						Entities: &[]tgbotapi.MessageEntity{
							{
								Offset: 0,
								Type:   "bot_command",
								Length: 7,
							},
						},
						Chat: &tgbotapi.Chat{
							ID: chatID,
						},
						Text: "/random",
					},
				},
			},
			check: func(t *testing.T, got tgbotapi.Message) {
				require.True(t, strings.HasPrefix(got.Text, "📝"))
			},
			wantErr: require.NoError,
		},
		{
			name: "Should change source to https://lesswrong.com",
			args: args{
				update: tgbotapi.Update{
					Message: &tgbotapi.Message{
						From: &tgbotapi.User{
							ID: userID,
						},
						Entities: &[]tgbotapi.MessageEntity{
							{
								Offset: 0,
								Type:   "bot_command",
								Length: 7,
							},
						},
						Chat: &tgbotapi.Chat{
							ID: chatID,
						},
						Text: "/source 4",
					},
				},
			},
			check: func(t *testing.T, got tgbotapi.Message) {
				require.Equal(t, "Changed source to https://lesswrong.com", got.Text)
			},
			wantErr: require.NoError,
		},
		{
			name: "Should get top posts from https://lesswrong.com",
			args: args{
				update: tgbotapi.Update{
					Message: &tgbotapi.Message{
						From: &tgbotapi.User{
							ID: userID,
						},
						Entities: &[]tgbotapi.MessageEntity{
							{
								Offset: 0,
								Type:   "bot_command",
								Length: 4,
							},
						},
						Chat: &tgbotapi.Chat{
							ID: chatID,
						},
						Text: "/top",
					},
				},
			},
			check: func(t *testing.T, got tgbotapi.Message) {
				require.True(t, strings.HasPrefix(got.Text, "🏆 Top posts this week from https://lesswrong.com"))
			},
			wantErr: require.NoError,
		},
		{
			name: "Should get random post from https://lesswrong.com",
			args: args{
				update: tgbotapi.Update{
					Message: &tgbotapi.Message{
						From: &tgbotapi.User{
							ID: userID,
						},
						Entities: &[]tgbotapi.MessageEntity{
							{
								Offset: 0,
								Type:   "bot_command",
								Length: 7,
							},
						},
						Chat: &tgbotapi.Chat{
							ID: chatID,
						},
						Text: "/random",
					},
				},
			},
			check: func(t *testing.T, got tgbotapi.Message) {
				require.True(t, strings.HasPrefix(got.Text, "📝"))
			},
			wantErr: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tgbot.MessageHandler(context.TODO(), tt.args.update)
			tt.wantErr(t, err)

			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}
