package bot

import (
	"context"
	"fmt"
	"io"
	"time"
	"videofetcher/internal/downloader"
	"videofetcher/internal/notifier"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var Logging iLogger

type TelegramBot struct {
	Options downloader.DownloaderOpts

	bot *tgbotapi.BotAPI
	d   downloader.Downloader
}

type MsgPayload struct {
	Text      string
	SourceMsg *tgbotapi.Message
	Video     io.Reader
}

func (m *TelegramBot) Inititalize(bot *tgbotapi.BotAPI) error {
	m.bot = bot
	m.d = *downloader.NewDownloader(&m.Options)
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
	return m.bot.Send(c)
}

func (m *TelegramBot) SendMsg(payload *MsgPayload) (rmsg tgbotapi.Message) {
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

	rmsg, err := m.Send(msg)
	if err != nil {
		Logging.Errorf("err sending a message: %v", err)
		m.SendMsg(&MsgPayload{Text: GetErrMsg(err), SourceMsg: payload.SourceMsg})
	}
	Logging.Infof("msg successfully sent to: %v", payload.SourceMsg.Chat.ID)
	return
}

func (m *TelegramBot) ProcessMessage(message tgbotapi.Message) {
	go func(msg tgbotapi.Message) {
		ctx := context.Background()

		n := notifier.NewMsgNotifier(m, msg.Chat.ID)

		m.d.SetNotifier(n)

		n.SendNotify("Got link. 👀 at 📼")

		ctx, cancel := context.WithTimeout(ctx, time.Duration(m.Options.Timeout)*time.Second)
		defer cancel()

		videos, err := m.d.Download(ctx, msg.Text)
		if err != nil {
			Logging.Errorf("err download video %s: %v", msg.Text, err)
			n.UpdTextNotify(GetErrMsg(err))
			return
		}

		if len(videos) > 1 {
			n.UpdTextNotify(fmt.Sprintf("📣 I found %v videos. Start sending", len(videos)))
		}

		for _, v := range videos {
			m.SendMsg(&MsgPayload{
				Text:      v.Title,
				Video:     v.Reader,
				SourceMsg: &msg,
			})
			v.Close()
		}

		n.UpdTextNotify("Done! ✅")
		time.Sleep(time.Duration(2 * time.Second))
		n.Close()

	}(message)
}
