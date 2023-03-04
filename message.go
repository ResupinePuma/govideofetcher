package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	nurl "net/url"
	"regexp"
	"strings"
	"time"
	"videofetcher/downloader"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const urlRegex = `(https?:\/\/(?:www\.|(!www))[a-zA-Z0-9][a-zA-Z0-9-]+[a-zA-Z0-9]\.[^\s]{2,}|www\.[a-zA-Z0-9][a-zA-Z0-9-]+[a-zA-Z0-9]\.[^\s]{2,}|https?:\/\/(?:www\.|(!www))[a-zA-Z0-9]+\.[^\s]{2,}|www\.[a-zA-Z0-9]+\.[^\s]{2,})`

type MsgWorker struct {
	Log *VFLogger
	bot *tgbotapi.BotAPI
	cfg *Config
}

type MsgPayload struct {
	Text      string
	SourceMsg *tgbotapi.Message
	Video     io.Reader
}

func NewMsgWorker(logger *VFLogger, bot *tgbotapi.BotAPI, cfg *Config) MsgWorker {
	return MsgWorker{
		Log: logger,
		bot: bot,
		cfg: cfg,
	}
}

func extractUrlAndText(str string) (url string, label string, err error) {
	re := regexp.MustCompile(`(?m)(https?:\/\/[^\s]+)`)
	url = re.FindString(str)
	if url == "" {
		err = errors.New("invalid url")
		return
	}
	u, errp := nurl.ParseRequestURI(url)
	if errp != nil {
		err = errors.New("invalid url")
		return
	}
	url = u.String()
	label = strings.TrimSpace(strings.Replace(str, url, "", -1))
	return
}

func (m *MsgWorker) SendMsg(payload *MsgPayload) (rmsg tgbotapi.Message) {
	var msg tgbotapi.Chattable
	switch p := payload; {
	case p.Video != nil:
		vidRdr := tgbotapi.FileReader{
			Name:   payload.Text,
			Reader: payload.Video,
		}
		msgv := tgbotapi.NewVideo(payload.SourceMsg.Chat.ID, vidRdr)
		msgv.Caption = payload.Text
		msg = msgv
	default:
		msgt := tgbotapi.NewMessage(payload.SourceMsg.Chat.ID, payload.Text)
		msgt.DisableWebPagePreview = true
		msg = msgt
	}

	rmsg, err := m.bot.Send(msg)
	if err != nil {
		nerr := err.(*nurl.Error)
		m.Log.Error("err sending a message: %v", err)
		m.SendMsg(&MsgPayload{Text: GetErrMsg(nerr.Err), SourceMsg: payload.SourceMsg})
	}
	m.Log.Info("msg successfully sent to: %v", payload.SourceMsg.Chat.ID)
	return
}

func (m *MsgWorker) Process(message tgbotapi.Message) {
	go func(msg tgbotapi.Message, cfg Config) {
		url, label, err := extractUrlAndText(msg.Text)
		if err != nil {
			m.Log.Error("err extracting url: %v", err)
			m.SendMsg(&MsgPayload{Text: GetErrMsg(err), SourceMsg: &msg})
			return
		}

		notifier := MsgNotifier{Bot: m.bot, ChatID: msg.Chat.ID}
		notifier.SendNotify(fmt.Sprintf("Got %s. 👀 at 📼", url))

		opts := downloader.DownloaderOpts{
			SizeLimit: cfg.Base.SizeLimit,
			Timeout:   cfg.Base.Timeout,
		}

		var d downloader.AbstractDownloader
		switch u := url; {
		case strings.Contains(u, "tiktok.com") || strings.Contains(u, "tt.com"):
			d = &cfg.TT
		default:
			d = &cfg.YTDL
		}
		err = d.Init(m.Log, &notifier, &opts)
		if err != nil {
			m.Log.Error("err init downloader %T: %v", d, err)
			notifier.UpdTextNotify(GetErrMsg(err))
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(30*time.Second))
		defer cancel()

		defer d.Close()
		alabel, vrdr, err := d.Download(ctx, url)
		if err != nil {
			m.Log.Error("err download video %s: %v", url, err)
			notifier.UpdTextNotify(GetErrMsg(err))
			return
		}

		switch l := label; {
		case l == "" && alabel != "":
			label = alabel
		case l == "" && alabel == "":
			label = url
		}

		var re = regexp.MustCompile(`(?m).(webm|mp4|mkv|gif|flv|avi|mov|wmv|asf)`)
		label = re.ReplaceAllString(label, "")

		m.SendMsg(&MsgPayload{
			Text:      label,
			Video:     vrdr,
			SourceMsg: &msg,
		})

		notifier.UpdTextNotify("Done! ✅")
		time.Sleep(time.Duration(2 * time.Second))
		notifier.Close()
	}(message, *m.cfg)
}
