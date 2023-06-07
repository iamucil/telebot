package main

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"os"
	"regexp"
	"sync"
	"text/template"
	"time"

	tbot "gopkg.in/telebot.v3"
)

var (
	tpl *template.Template

	modulesMu sync.RWMutex
	token     string
)

func init() {
	tplStr := `<tg-spoiler>{{ .ID }}</tg-spoiler> (<strong>{{.Uname }}</strong>)`
	tpl = template.Must(template.New("bot-template").Parse(tplStr))
}

func main() {
	for _, flag := range os.Args[:] {
		if flag == "--version" {
			return
		}
	}

	if telegramTokenStr, found := os.LookupEnv("TELEGRAM_TOKEN"); found {
		if telegramTokenStr != "" {
			token = telegramTokenStr
		}
	}
	if token == "" {
		panic("bad token")
	}
	var defaultTransport http.RoundTripper = &http.Transport{
		Proxy: nil,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          30,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   15 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	client := &http.Client{
		Transport: defaultTransport,
	}

	pref := tbot.Settings{
		Token:  token,
		Poller: &tbot.LongPoller{Timeout: 10 * time.Second},
		Client: client,
	}

	bot, err := tbot.NewBot(pref)
	if err != nil {
		panic(err)
	}

	if profile := bot.Me; profile != nil {
		fmt.Fprintf(os.Stdout, "Telegram bot running: %s\r\n", profile.Username)
	}

	bot.Handle(tbot.OnText, func(ctx tbot.Context) error {
		modulesMu.RLock()
		_ = respond(ctx)
		modulesMu.RUnlock()
		return nil
	})
	bot.Handle(tbot.OnAddedToGroup, func(ctx tbot.Context) error {
		modulesMu.RLock()
		_ = respond(ctx)
		modulesMu.RUnlock()
		return nil
	})
	bot.Handle(tbot.OnChannelPost, func(ctx tbot.Context) error {
		modulesMu.RLock()
		_ = respond(ctx)
		modulesMu.RUnlock()
		return nil
	})
	bot.Start()

}

func respond(ctx tbot.Context) error {
	chat := ctx.Chat()
	if chat == nil {
		return nil
	}
	msg := ctx.Message()
	if msg == nil {
		return nil
	}

	entities := msg.Entities
	if len(entities) == 0 {
		return nil
	}

	if entities[0].Type != "bot_command" {
		return nil
	}
	text := ctx.Text()
	re := regexp.MustCompile(".*whoami.*")
	if !re.MatchString(text) {
		return nil
	}

	ctx.Notify(tbot.Typing)
	time.Sleep(2 * time.Second)
	user := ctx.Sender()
	buf := new(bytes.Buffer)
	switch t := chat.Type; t {
	case "group", "channel", "supergroup":
		_ = tpl.Execute(buf, map[string]interface{}{
			"ID":    chat.ID,
			"Uname": chat.Title,
		})
	case "private":
		_ = tpl.Execute(buf, map[string]interface{}{
			"ID":    user.ID,
			"Uname": user.Username,
		})
	}
	textMessage := buf.String()
	if textMessage == "" {
		return nil
	}
	return ctx.Send(buf.String(), tbot.ModeHTML)
}
