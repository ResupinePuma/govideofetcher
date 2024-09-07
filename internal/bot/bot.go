package bot

import (
	"context"
	"fmt"
	"strings"
	"time"
	"videofetcher/internal/downloader"
	"videofetcher/internal/downloader/dcontext"
	"videofetcher/internal/downloader/options"
	"videofetcher/internal/downloader/parsers/instagram"
	"videofetcher/internal/downloader/parsers/tiktok"
	"videofetcher/internal/downloader/parsers/ytdl"
	"videofetcher/internal/downloader/video"
	"videofetcher/internal/notifier"
	"videofetcher/internal/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var Logging iLogger

const (
	TypeVideo    = 1
	TypeFeedback = 2
)

type TelegramBot struct {
	Options options.DownloaderOpts

	vidCache map[string]bool
	fdbcache map[int64]time.Time
	username string
	bot      *tgbotapi.BotAPI
	d        downloader.Downloader
}

type MsgPayload struct {
	Type      int
	Text      string
	SourceMsg *tgbotapi.Message
	Videos    []video.Video
}

func (m *TelegramBot) Inititalize(bot *tgbotapi.BotAPI) error {
	ytdl.Logger = Logging
	tiktok.Logger = Logging
	instagram.Logger = Logging

	m.bot = bot
	m.fdbcache = make(map[int64]time.Time)
	m.vidCache = make(map[string]bool)
	me, err := bot.GetMe()
	if err != nil {
		return err
	}
	m.username = me.UserName
	m.d = *downloader.NewDownloader(m.Options)
	return nil
}

func (m *TelegramBot) NewMessage(chatID int64, text string) tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(chatID, text)
}
func (m *TelegramBot) NewDeleteMessage(chatID int64, messageID int) tgbotapi.DeleteMessageConfig {
	return tgbotapi.NewDeleteMessage(chatID, messageID)
}
func (m *TelegramBot) NewEditMessageText(chatID int64, messageID int, text string) tgbotapi.EditMessageTextConfig {
	return tgbotapi.NewEditMessageText(chatID, messageID, text)
}

func (m *TelegramBot) NewEditMessageTextAndMarkup(chatID int64, messageID int, text string, replyMarkup tgbotapi.InlineKeyboardMarkup) tgbotapi.EditMessageTextConfig {
	return tgbotapi.NewEditMessageText(chatID, messageID, text)
}
func (m *TelegramBot) Send(c tgbotapi.Chattable) ([]tgbotapi.Message, error) {
	return m.bot.Send(c)
}

func (m *TelegramBot) SendMsg(payload *MsgPayload) (rmsg []tgbotapi.Message) {
	var msg tgbotapi.Chattable
	switch payload.Type {
	case TypeVideo:
		switch p := payload; {
		case p.Videos != nil:
			defer func() {
				for _, vid := range p.Videos {
					vid.Close()
				}
			}()

			vids := []interface{}{}
			for _, vid := range p.Videos {
				media := tgbotapi.NewInputMediaVideo(&vid)
				if len(vid.Title) <= 1024 {
					media.Caption = vid.Title
				}
				media.SupportsStreaming = true
				vids = append(vids, media)
			}
			group := tgbotapi.NewMediaGroup(payload.SourceMsg.Chat.ID, vids)
			msg = group
		default:
			msgt := tgbotapi.NewMessage(payload.SourceMsg.Chat.ID, payload.Text)
			msg = msgt
		}
	case TypeFeedback:
		msgt := tgbotapi.NewMessage(payload.SourceMsg.Chat.ID, payload.Text)
		msg = msgt
	}

	rmsg, err := m.Send(msg)
	if err != nil {
		Logging.Errorf("err sending a message: %v", err)
		m.SendMsg(&MsgPayload{Text: GetErrMsg(err), SourceMsg: payload.SourceMsg})
	}
	Logging.Infof("msg successfully sent to: %v", payload.SourceMsg.Chat.ID)
	return
}

func usage(cmd string) string {
	cmd = strings.TrimPrefix(cmd, "/")
	lines := map[string]string{
		"get":      "<URL> <caption (optional)> - Downloаd video from URL",
		"feedback": "send your feedback to admin",
	}
	if _, ok := lines[cmd]; !ok {
		return "Command not exist"
	}

	return fmt.Sprintf("/%s %s", cmd, lines[cmd])
}

func (m *TelegramBot) fetcher(msg tgbotapi.Message) {
	ctx := context.Background()

	url, label, err := utils.ExtractUrlAndText(msg.Text)
	if err != nil {
		return
	}

	if _, ok := m.vidCache[url.String()]; ok {
		return
	}

	m.vidCache[url.String()] = true

	defer delete(m.vidCache, url.String())

	n := notifier.NewMsgNotifier(m, msg.Chat.ID)

	m.d.SetNotifier(n)

	n.SendNotify("Got link. 👀 at 📼")

	ctx, cancel := context.WithTimeout(ctx, time.Duration(m.Options.Timeout)*time.Second)
	defer cancel()

	dctx := dcontext.NewDownloaderContext(ctx, n)

	videos, err := m.d.Download(dctx, url, label)
	if err != nil {
		Logging.Errorf("err download video %s: %v", msg.Text, err)
		n.UpdTextNotify(GetErrMsg(err))
		return
	}

	if len(videos) > 1 {
		n.UpdTextNotify(fmt.Sprintf("📣 I found %v videos. Start sending", len(videos)))
	}

	m.SendMsg(&MsgPayload{
		Type:      TypeVideo,
		Videos:    videos,
		SourceMsg: &msg,
	})

	n.UpdTextNotify("Done! ✅")
	time.Sleep(time.Duration(2 * time.Second))

	n.Close()
}

func (m *TelegramBot) ProcessMessage(message tgbotapi.Message) {
	go func(msg tgbotapi.Message) {
		if msg.From.IsBot {
			return
		}

		// if msg is command and empty
		if msg.IsCommand() {
			switch msg.Command() {
			case "get":
				text := msg.CommandArguments()
				if text == "" {
					sent := m.SendMsg(&MsgPayload{
						Type:      TypeVideo,
						Text:      usage(msg.Command()),
						SourceMsg: &msg,
					})
					if len(sent) == 0 {
						return
					}
					time.Sleep(10 * time.Second)
					m.bot.Send(tgbotapi.NewDeleteMessage(msg.Chat.ID, sent[0].MessageID))
					return
				}
				msg.Text = text

				m.fetcher(msg)
			case "feedback":
				if v, ok := m.fdbcache[msg.Chat.ID]; ok {
					if time.Now().Sub(v).Seconds() <= 30 {
						return
					}
				}

				m.fdbcache[msg.Chat.ID] = time.Now()

				mc := tgbotapi.NewMessage(
					m.Options.AdminID,
					fmt.Sprintf("#feedback from %v\n\n%s", msg.Chat.UserName, msg.CommandArguments()),
				)
				_, err := m.bot.Send(mc)
				if err != nil {
					return
				}

				mc = tgbotapi.NewMessage(msg.Chat.ID, "Thanks for your feedback! ✅")

				sent, err := m.bot.Send(mc)
				if err != nil {
					return
				}
				if len(sent) == 0 {
					return
				}

				time.Sleep(10 * time.Second)
				m.bot.Send(tgbotapi.NewDeleteMessage(msg.Chat.ID, sent[0].MessageID))
			}
		} else {
			m.fetcher(msg)
		}

	}(message)
}
