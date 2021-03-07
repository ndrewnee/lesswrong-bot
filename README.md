# 🤖 lesswrong-bot

[![Go](https://github.com/ndrewnee/lesswrong-bot/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/ndrewnee/lesswrong-bot/actions/workflows/go.yml)

[Telegram bot](https://t.me/lesswrong_bot) for reading posts from:

- [Lesswrong.ru](https://lesswrong.ru) (default)
- [Slate Star Codex](https://slatestarcodex.com)
- [Astral Codex Ten](https://astralcodexten.substack.com)
- [Lesswrong.com](https://lesswrong.com)

## 😎 Usage

Commands:

/top - Top posts

/random - Read random post

/source - Change source:

  1. [Lesswrong.ru](https://lesswrong.ru) (default)
  2. [Slate Star Codex](https://slatestarcodex.com)
  3. [Astral Codex Ten](https://astralcodexten.substack.com).
  4. [Lesswrong.com](https://lesswrong.com)

/help - Help

## 🧑‍💻 Run locally

Register new bot at https://t.me/BotFather or use previously created one.

Take bot access token.

Before running copy sample file and replace env vars with your credentials

```sh
cp .env.dev.sample .env.dev
source .env.dev
```

Run application locally:

```sh
make run
```

Redis is used as cache. If redis isn't available fallbacks to memory cache.

Also you can run bot with redis in docker compose:

```sh
docker-compose up
```

## 👷 Build

Build binary

```sh
make build
```

## 🧪 Testing

Run unit tests

```sh
make test
```

Before running tests copy example env file and replace env vars with your credentials

```sh
cp .env.test.sample .env.test
source .env.test
```

and then run integration tests

```sh
make test_integration
```

## 🖍 Lint

Run linters

```sh
make lint
```

## 🛥 Deployment

Automatic CI/CD pipelines are building and testing the bot on each PR.

Demo bot is deployed to production on Heroku on merge to master.

To deploy your app on Heroku read [documentation](https://devcenter.heroku.com/articles/getting-started-with-go?singlepage=true).

```sh
brew install heroku/brew/heroku

heroku login
heroku create <app-name>
heroku config:set WEBHOOK=true
heroku config:set TOKEN=<token>

git push heroku main
```

If application is already setup just run:

```sh
make deploy
```

## 🛠 Environment variables

| Env var      | Type    | Description                   | Default                             |
| ------------ | ------- | ----------------------------- | ----------------------------------- |
| REDIS_URL    | String  | Redis connection string       | redis://localhost:6379/1            |
| TOKEN        | String  | Telegram bot access token     |                                     |
| DEBUG        | Boolean | Enable debug mode             | false                               |
| WEBHOOK      | Boolean | Enable webhook mode           | false                               |
| PORT         | String  | Port for webhook              | 9999                                |
| WEBHOOK_HOST | String  | Webhook host for telegram bot | https://lesswrong-bot.herokuapp.com |
| TIMEOUT      | Integer | Request timeout in seconds    | 15                                  |
| CACHE_EXPIRE | Integer | Posts cache expire in hours   | 24                                  |
