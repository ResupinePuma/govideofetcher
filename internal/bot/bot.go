package bot

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf16"
	"videofetcher/internal/downloader"
	"videofetcher/internal/downloader/dcontext"
	"videofetcher/internal/downloader/media"
	"videofetcher/internal/downloader/options"
	"videofetcher/internal/downloader/parsers/instagram"
	"videofetcher/internal/downloader/parsers/tiktok"
	"videofetcher/internal/downloader/parsers/ytdl"
	"videofetcher/internal/notifier"
	"videofetcher/internal/userdb"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	Logging       Logger
	ErrInvalidUrl = errors.New("Invalid url")
)

const (
	TypeMedia    = 1
	TypeFeedback = 2

	LinkMsg = "🔗Link"
)

type TelegramBot struct {
	Options options.DownloaderOpts
	Userdb  *userdb.UserDB

	vidCache map[string]bool
	fdbcache map[int64]time.Time
	username string
	bot      *tgbotapi.BotAPI
	d        downloader.Downloader
}

type MsgPayload struct {
	Type        int
	Text        string
	OriginalURL *url.URL
	SourceMsg   *tgbotapi.Message
	Media       []media.Media
}

func (m *TelegramBot) Inititalize(bot *tgbotapi.BotAPI) error {
	ytdl.Logging = Logging
	tiktok.Logging = Logging
	instagram.Logging = Logging

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
func (m *TelegramBot) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	res, err := m.bot.Send(c)
	if err != nil {
		return res, nil
	}
	return res, nil
}
func utf16Len(s string) int {
	return len(utf16.Encode([]rune(s)))
}

func (m *TelegramBot) SendMsg(payload *MsgPayload) (rmsg tgbotapi.Message, err error) {
	var msg tgbotapi.Chattable
	var re = regexp.MustCompile(`(?m).(webm|mp4|mkv|gif|flv|avi|mov|wmv|asf)`)
	switch payload.Type {
	case TypeMedia:
		if len(payload.Media) == 0 {
			msgt := tgbotapi.NewMessage(payload.SourceMsg.Chat.ID, payload.Text)
			msg = msgt
			break
		}

		items := make([]interface{}, 0, len(payload.Media))
		for _, m := range payload.Media {
			defer m.Close()
			switch v := m.(type) {
			case *media.Video:
				vid := tgbotapi.NewInputMediaVideo(v)
				if payload.Text != "" {
					v.Title = payload.Text
				}
				if v.Title != "" {
					v.Title = re.ReplaceAllString(v.Title, "")
					if len(v.Title) <= 1024 {
						vid.Caption = v.Title
					}
					vid.Caption += "\n\n" + LinkMsg
				} else {
					vid.Caption += LinkMsg
				}
				vid.CaptionEntities = []tgbotapi.MessageEntity{
					{
						Type:   "text_link",
						Offset: utf16Len(vid.Caption) - utf16Len(LinkMsg),
						Length: utf16Len(LinkMsg),
						URL:    payload.OriginalURL.String(),
					},
				}

				if v.Thumbnail != nil {
					vid.Thumb = tgbotapi.FileReader{
						Name:   "thumbnail.png",
						Reader: v.Thumbnail,
					}
				}
				vid.Duration = int(v.Duration)
				vid.SupportsStreaming = true
				items = append(items, vid)
			case *media.Audio:
				mus := tgbotapi.NewInputMediaAudio(v)
				if payload.Text != "" {
					v.Title = payload.Text
				}
				v.Title = re.ReplaceAllString(v.Title, "")
				mus.Title = v.Title
				mus.Performer = v.Artist
				mus.Duration = int(v.Duration)
				if v.Thumbnail != nil {
					mus.Thumb = tgbotapi.FileReader{
						Name:   "thumbnail.png",
						Reader: v.Thumbnail,
					}
				}
				mus.Caption += LinkMsg
				mus.CaptionEntities = []tgbotapi.MessageEntity{
					{
						Type:   "text_link",
						Offset: utf16Len(mus.Caption) - utf16Len(LinkMsg),
						Length: utf16Len(LinkMsg),
						URL:    payload.OriginalURL.String(),
					},
				}
				items = append(items, mus)
			}
		}
		msgt := tgbotapi.NewMediaGroup(payload.SourceMsg.Chat.ID, items)
		msg = msgt

	default:
		msgt := tgbotapi.NewMessage(payload.SourceMsg.Chat.ID, payload.Text)
		msg = &msgt
	}

	if msg == nil {
		return
	}

	rmsg, err = m.Send(msg)
	if err != nil {
		Logging.Errorf("err sending a message: %v", err)
		_, e := m.SendMsg(&MsgPayload{Text: GetErrMsg(err), SourceMsg: payload.SourceMsg})
		if e != nil {
			return rmsg, e
		}
	}
	Logging.Infof("msg successfully sent to: %v", payload.SourceMsg.Chat.ID)
	return
}

func usage(cmd string) string {
	cmd = strings.TrimPrefix(cmd, "/")
	lines := map[string]string{
		"get": "<URL> <caption (optional)> - Downloаd video from URL",
	}
	if _, ok := lines[cmd]; !ok {
		return "Command not exist"
	}

	return fmt.Sprintf("/%s %s", cmd, lines[cmd])
}

func extractUrlAndText(str string) (addr *url.URL, label string, err error) {
	re := regexp.MustCompile(`(?m)(https?:\/\/[^\s]+)`)
	ustr := re.FindString(str)
	if ustr == "" {
		err = errors.New("invalid url")
		return
	}
	addr, errp := url.ParseRequestURI(ustr)
	if errp != nil {
		err = errors.New("invalid url")
		return
	}
	label = strings.TrimSpace(strings.Replace(str, addr.String(), "", -1))
	return
}

func (m *TelegramBot) fetcher(msg tgbotapi.Message) {
	ctx := context.Background()

	url, label, err := extractUrlAndText(msg.Text)
	if err != nil {
		return
	}

	if _, ok := m.vidCache[url.String()]; ok {
		return
	}

	m.vidCache[url.String()] = true

	if strings.HasSuffix(url.Hostname(), "ru") || strings.Contains(url.Hostname(), "vk") {
		m.SendMsg(&MsgPayload{
			SourceMsg: &msg,
			Text:      "❌service is not supported",
		})
		return
	}

	defer delete(m.vidCache, url.String())

	n := notifier.NewMsgNotifier(m, msg.Chat.ID)

	m.d.SetNotifier(n)

	n.SendNotify("Got link. 👀 at 📼")

	ctx, cancel := context.WithTimeout(ctx, time.Duration(m.Options.Timeout)*time.Second)
	defer cancel()

	dctx := dcontext.NewDownloaderContext(ctx, n)

	dctx.SetUrl(url)
	dctx.SetLang(msg.From.LanguageCode)

	m.Userdb.Add(int(msg.Chat.ID), msg.Chat.UserName, msg.Text)

	media, err := m.d.Download(dctx, url)
	if err != nil {
		Logging.Errorf("err download video %s: %v", msg.Text, err)
		n.Close()

		_, err = m.bot.Send(tgbotapi.NewAnimation(msg.Chat.ID, tgbotapi.FileURL(RandomGif())))
		if err != nil {
			Logging.Errorf("err sending gif %s: %v", msg.Text, err)
		}
		return
	}

	if len(media) > 1 {
		n.UpdTextNotify(fmt.Sprintf("📣 I found %v media. Start sending", len(media)))
	}

	m.SendMsg(&MsgPayload{
		Type:        TypeMedia,
		Text:        label,
		Media:       media,
		OriginalURL: url,
		SourceMsg:   &msg,
	})

	n.UpdTextNotify("Done! ✅")
	cancel()
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
					sent, e := m.SendMsg(&MsgPayload{
						Type:      TypeMedia,
						Text:      usage(msg.Command()),
						SourceMsg: &msg,
					})
					if e != nil {
						return
					}
					time.Sleep(10 * time.Second)
					m.bot.Send(tgbotapi.NewDeleteMessage(msg.Chat.ID, sent.MessageID))
					return
				}
				msg.Text = text

				m.fetcher(msg)
			}
		} else {
			m.fetcher(msg)
		}

	}(message)
}
