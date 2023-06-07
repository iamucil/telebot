# Telebot

`whoami` command for telegram. Use this to get your channel group or your user id. Just send text message with `/whoami` into your bot.  
Read [telegram api bot](https://core.telegram.org/bots/api) documentation to generate telegram bot token.

## Run

Build docker image, parse your telegram token when building the image. 
```shell
$ docker build -t tbot --build-arg token=YOUR_TELEGRAM_TOKEN -f ./Dockerfile .
$ docker run -d tbot:latest
```