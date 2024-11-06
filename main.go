package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"
	"text/template"
	"time"

	tbot "gopkg.in/telebot.v4"
)

var (
	tpl          *template.Template
	senderTplMsg *template.Template

	modulesMu sync.RWMutex
	token     string
)

func init() {
	tplStr := `{{.ChatType}} <tg-spoiler>{{ .ID }}</tg-spoiler> (<strong>{{.Uname }}</strong>) 
	{{if .ThreadID}}
	Call from Topic ID: <strong>{{.ThreadID}}</strong> [<strong>{{.TopicName}}</strong>]
	{{end}}
	<strong>Sender</strong><code>
	Username : {{.Username}}
	Name     : {{.FirstName}} {{.LastName}}
	</code>`
	senderTplMsgStr := `User has running bot command {{if .Command}}<code>{{.Command}}</code>{{end}}
	<strong>Sender</strong><code>
	{{.Sender}}
	</code>`
	tpl = template.Must(template.New("bot-template").Parse(tplStr))
	senderTplMsg = template.Must(template.New("bot-template-for-sender").Parse(senderTplMsgStr))
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

	if b, err := json.MarshalIndent(ctx.Sender(), "", "    "); err == nil {
		sendTo := tbot.Chat{
			ID: 10479058,
		}
		buff := new(bytes.Buffer)
		if err := senderTplMsg.Execute(buff, map[string]any{
			"Command": "",
			"Sender":  string(b),
		}); err == nil {
			ctx.Bot().Send(&sendTo, buff.String(), tbot.ModeHTML)
		}
	}
	sender := ctx.Sender()
	ctx.Notify(tbot.Typing)
	time.Sleep(1 * time.Second)
	user := ctx.Sender()
	buf := new(bytes.Buffer)
	switch t := chat.Type; t {
	case "group", "channel", "supergroup":
		var threadID string
		var topicName string
		if id := msg.ThreadID; id != 0 {
			threadID = strconv.Itoa(id)
		}
		if topicMsg := msg.ReplyTo; topicMsg != nil {
			if topic := topicMsg.TopicCreated; topic != nil {
				topicName = topic.Name
			}
		}

		_ = tpl.Execute(buf, map[string]any{
			"ChatType":  t,
			"ID":        chat.ID,
			"Uname":     chat.Title,
			"ThreadID":  threadID,
			"TopicName": topicName,
			"Username":  sender.Username,
			"FirstName": sender.FirstName,
			"LastName":  sender.LastName,
		})
	case "private":
		_ = tpl.Execute(buf, map[string]any{
			"ChatType":  t,
			"ID":        user.ID,
			"Uname":     user.Username,
			"ThreadID":  "",
			"TopicName": "",
			"Username":  sender.Username,
			"FirstName": sender.FirstName,
			"LastName":  sender.LastName,
		})
	}
	textMessage := buf.String()
	if textMessage == "" {
		return nil
	}

	if err := ctx.Send(buf.String(), tbot.ModeHTML); err != nil {
		panic(err)
	}
	return nil
}
