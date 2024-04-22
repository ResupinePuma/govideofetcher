package notifier

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	GREENTILE = "🟩"
	EMPTYTILE = "      "
)

type iTelegram interface {
	NewMessage(chatID int64, text string) tgbotapi.MessageConfig
	NewDeleteMessage(chatID int64, messageID int) tgbotapi.DeleteMessageConfig
	NewEditMessageText(chatID int64, messageID int, text string) tgbotapi.EditMessageTextConfig
	//NewInlineKeyboardButtonData(text, data string) tgbotapi.InlineKeyboardButton
	NewEditMessageTextAndMarkup(chatID int64, messageID int, text string, replyMarkup tgbotapi.InlineKeyboardMarkup) tgbotapi.EditMessageTextConfig
	Send(c tgbotapi.Chattable) ([]tgbotapi.Message, error)
}

type MsgNotifier struct {
	ChatID      int64
	MsgID       int
	ProgressMsg string

	oldMsg string

	bot iTelegram
}

func NewMsgNotifier(bot iTelegram, chatid int64) *MsgNotifier {
	return &MsgNotifier{
		bot:    bot,
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

func (m *MsgNotifier) MakeProgressBar(percent float64) (err error) {
	percent = percent / 10

	if percent > 100 {
		percent = 100
	} else if percent < 0 {
		percent = 0
	}

	greens := strings.Repeat(GREENTILE, int(percent))
	emptys := strings.Repeat(EMPTYTILE, 10-int(percent))
	newMsg := fmt.Sprintf("%s\n|%s%s| %d", m.ProgressMsg, greens, emptys, int(percent)*10) + "%"
	if newMsg == m.oldMsg {
		return
	}
	msg := tgbotapi.NewEditMessageText(m.ChatID, m.MsgID, newMsg)
	_, err = m.bot.Send(msg)
	if err != nil {
		return err
	}
	m.oldMsg = newMsg
	return
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
	m.ChatID = resp[0].Chat.ID
	m.MsgID = resp[0].MessageID
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
