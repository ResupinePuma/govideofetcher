package notifier

import (
	"context"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	GREENTILE = "🟩"
	EMPTYTILE = "⬛"
)

type iTelegram interface {
	NewMessage(chatID int64, text string) tgbotapi.MessageConfig
	NewDeleteMessage(chatID int64, messageID int) tgbotapi.DeleteMessageConfig
	NewEditMessageText(chatID int64, messageID int, text string) tgbotapi.EditMessageTextConfig
	//NewInlineKeyboardButtonData(text, data string) tgbotapi.InlineKeyboardButton
	NewEditMessageTextAndMarkup(chatID int64, messageID int, text string, replyMarkup tgbotapi.InlineKeyboardMarkup) tgbotapi.EditMessageTextConfig
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}

type MsgNotifier struct {
	ChatID      int64
	MsgID       int
	ProgressMsg string

	oldMsg string

	bot  iTelegram
	stop chan bool
}

func NewMsgNotifier(bot iTelegram, chatid int64) *MsgNotifier {
	return &MsgNotifier{
		bot:    bot,
		stop:   make(chan bool, 1),
		ChatID: chatid,
	}
}

func (m *MsgNotifier) UpdTextNotify(text string) (err error) {
	msg := tgbotapi.NewEditMessageText(m.ChatID, m.MsgID, text)
	_, err = m.bot.Send(msg)
	if err != nil {
		return err
	}
	m.ProgressMsg = text
	return
}

func (m *MsgNotifier) StartTicker(ctx context.Context) (err error) {
	ticker := time.NewTicker(time.Second)
	offset := 0
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			pbar := []string{EMPTYTILE, EMPTYTILE, EMPTYTILE, EMPTYTILE, EMPTYTILE, EMPTYTILE, EMPTYTILE, EMPTYTILE, EMPTYTILE, EMPTYTILE}
			pbar[offset] = GREENTILE
			offset++
			if offset >= 10 {
				offset = 0
			}

			newMsg := fmt.Sprintf("%s\n|%s|", m.ProgressMsg, strings.Join(pbar, ""))
			if newMsg == m.oldMsg {
				return
			}
			msg := tgbotapi.NewEditMessageText(m.ChatID, m.MsgID, newMsg)
			_, err = m.bot.Send(msg)
			if err != nil {
				return err
			}
			m.oldMsg = newMsg
		}
	}
}

func (m *MsgNotifier) DelProgressBar() (err error) {
	msg := tgbotapi.NewEditMessageText(m.ChatID, m.MsgID, m.ProgressMsg)
	_, err = m.bot.Send(msg)
	if err != nil {
		return err
	}
	return
}

func (m *MsgNotifier) SendNotify(text string) (err error) {
	msg := tgbotapi.NewMessage(m.ChatID, text)
	resp, err := m.bot.Send(msg)
	if err != nil {
		return err
	}
	m.ChatID = resp.Chat.ID
	m.MsgID = resp.MessageID
	return
}

func (m *MsgNotifier) Close() (err error) {
	msg := tgbotapi.NewDeleteMessage(m.ChatID, m.MsgID)
	_, err = m.bot.Send(msg)
	if err != nil {
		return err
	}
	return
}

// func (m *MsgNotifier) DrawKeyboard(params []string) (err error) {
// 	btns := []tgbotapi.InlineKeyboardButton{}
// 	for i, p := range params {
// 		btns = append(btns, tgbotapi.NewInlineKeyboardButtonData(p, strconv.Itoa(i)))
// 	}
// 	mrkup := tgbotapi.NewInlineKeyboardMarkup(btns)
// 	msg := tgbotapi.NewEditMessageTextAndMarkup(
// 		m.ChatID,
// 		m.MsgID,
// 		"Select preffered quality:",
// 		mrkup,
// 	)
// 	_, err = m.Bot.Send(msg)
// 	if err != nil {
// 		return err
// 	}
// 	return
// }
