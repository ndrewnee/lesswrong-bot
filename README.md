# lesswrong-bot

[Telegram bot](https://t.me/lesswrong_bot) for [Scott Alexander's new blog](https://astralcodexten.substack.com).

Maybe in future I'll add [his old blog](https://slatestarcodex.com) and [Lesswrong](https://lesswrong.com).

## Setup

Register new bot at https://t.me/BotFather or use previously created one.

Take bot access token and export it in terminal:

```sh
export TOKEN=<token>
```

## Run

To run application locally simply run:

```sh
make run
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
