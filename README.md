# lesswrong-bot

[![Go](https://github.com/ndrewnee/lesswrong-bot/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/ndrewnee/lesswrong-bot/actions/workflows/go.yml)

[Telegram bot](https://t.me/lesswrong_bot) for Scott Alexander's blogs, [old](https://slatestarcodex.com) and [new](https://astralcodexten.substack.com).

Maybe in the future I'll add [Lesswrong](https://lesswrong.com) and [Lesswrong.ru](https://lesswrong.ru).

## Usage

🤖 I'm a bot for reading posts from https://slatestarcodex.com (default) and https://astralcodexten.substack.com.

Commands:
 
/top - Top posts

/random - Read random post

/source - Change source (1 - slatestarcodex, 2 - astralcodexten)

/help - Help

## Run

Register new bot at https://t.me/BotFather or use previously created one.

Take bot access token and export it in terminal:

```sh
export TOKEN=<token>
```

Run application locally:

```sh
make run
```

## Testing

Run unit tests

```bash
make test
```

Run integration tests

```bash
make test_integration
```

## Lint

Run linters

```bash
make lint
```

## Deploy

If application is already setup just run:

```sh
make deploy
```

To deploy app on Heroku read [documentation](https://devcenter.heroku.com/articles/getting-started-with-go?singlepage=true).

```sh
brew install heroku/brew/heroku

heroku login
heroku create <app-name>
heroku config:set WEBHOOK=true
heroku config:set TOKEN=<token>

git push heroku main
```

## Environment variables

| Env var      | Type    | Description                   | Default                             |
| ------------ | ------- | ----------------------------- | ----------------------------------- |
| TOKEN        | String  | Telegram bot access token     |                                     |
| DEBUG        | Boolean | Enable debug mode             | false                               |
| WEBHOOK      | Boolean | Enable webhook mode           | false                               |
| PORT         | String  | Port for webhook              | 9999                                |
| WEBHOOK_HOST | String  | Webhook host for telegram bot | https://lesswrong-bot.herokuapp.com |
