# 🤖 lesswrong-bot

[![Go](https://github.com/ndrewnee/lesswrong-bot/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/ndrewnee/lesswrong-bot/actions/workflows/go.yml)

[Telegram bot](https://t.me/lesswrong_bot) for reading posts from [Lesswrong.ru](https://lesswrong.ru) and Scott Alexander's blog, [the old](https://slatestarcodex.com) and [new one](https://astralcodexten.substack.com).

Maybe in the future I'll add [Lesswrong.com](https://lesswrong.com).

## 😎 Usage

🤖 I'm a bot for reading posts from https://slatestarcodex.com, https://astralcodexten.substack.com and https://lesswrong.ru.

Commands:

/top - Top posts

/random - Read random post

/source - Change source (1 - slatestarcodex, 2 - astralcodexten, 3 - lesswrong.ru)

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
