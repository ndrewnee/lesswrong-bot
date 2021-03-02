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

Take bot access token and export it in terminal:

```sh
export TOKEN=<token>
```

Run application locally:

```sh
make run
```

## 👷 Build

Build binary

```bash
make build
```

## 🧪 Testing

Run unit tests

```bash
make test
```

Run integration tests

```bash
make test_integration
```

## 🖍 Lint

Run linters

```bash
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
| TOKEN        | String  | Telegram bot access token     |                                     |
| DEBUG        | Boolean | Enable debug mode             | false                               |
| WEBHOOK      | Boolean | Enable webhook mode           | false                               |
| PORT         | String  | Port for webhook              | 9999                                |
| WEBHOOK_HOST | String  | Webhook host for telegram bot | https://lesswrong-bot.herokuapp.com |
