# Telebot

`whoami` command for telegram. Use this to get your channel group or your user id. Just send text message with `/whoami` into your bot.  
Read [telegram api bot](https://core.telegram.org/bots/api) documentation to generate telegram bot token.

## Run

Build docker image, parse your telegram token when building the image. 
```shell
$ docker build -t tbot --build-arg token=YOUR_TELEGRAM_TOKEN -f ./Dockerfile .
$ docker run -d tbot:latest
```

## Implementation

Open your telegram app,  
1. Create Group or Channel
2. Invite your bot, make sure `bot` is set as admin if you are creating channel.
3. Send message with command `/whoami` in the text field.

Or you can send `/whoami` directly to your bot. Just create new chat with your bot.