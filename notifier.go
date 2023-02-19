package main

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	GREENTILE = "🟩"
	EMPTYTILE = "      "
)

type MsgNotifier struct {
	Bot *tgbotapi.BotAPI

	ChatID int64
	MsgID  int

	ProgressMsg string

	oldMsg string
}

func (m *MsgNotifier) UpdTextNotify(text string) (err error) {
	msg := tgbotapi.NewEditMessageText(m.ChatID, m.MsgID, text)
	msg.DisableWebPagePreview = true
	_, err = m.Bot.Send(msg)
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
	_, err = m.Bot.Send(msg)
	if err != nil {
		return err
	}
	m.oldMsg = newMsg
	return
}

func (m *MsgNotifier) DelProgressBar() (err error) {
	msg := tgbotapi.NewEditMessageText(m.ChatID, m.MsgID, m.ProgressMsg)
	_, err = m.Bot.Send(msg)
	if err != nil {
		return err
	}
	return
}

func (m *MsgNotifier) SendNotify(text string) (err error) {
	msg := tgbotapi.NewMessage(m.ChatID, text)
	msg.DisableWebPagePreview = true
	resp, err := m.Bot.Send(msg)
	if err != nil {
		return err
	}
	m.ChatID = resp.Chat.ID
	m.MsgID = resp.MessageID
	return
}

func (m *MsgNotifier) Close() (err error) {
	msg := tgbotapi.NewDeleteMessage(m.ChatID, m.MsgID)
	_, err = m.Bot.Send(msg)
	if err != nil {
		return err
	}
	return
}

// Synonim for MakeProgressBar
func (m *MsgNotifier) Count(percent float64) (err error) {
	return m.MakeProgressBar(percent)
}

// Synonim for UpdProgressMessage
func (m *MsgNotifier) Message(text string) (err error) {
	return m.UpdTextNotify(text)
}

func (m *MsgNotifier) DrawKeyboard(params []string) (err error) {
	btns := []tgbotapi.InlineKeyboardButton{}
	for i, p := range params {
		btns = append(btns, tgbotapi.NewInlineKeyboardButtonData(p, strconv.Itoa(i)))
	}
	mrkup := tgbotapi.NewInlineKeyboardMarkup(btns)
	msg := tgbotapi.NewEditMessageTextAndMarkup(
		m.ChatID,
		m.MsgID,
		"Select preffered quality:",
		mrkup,
	)
	_, err = m.Bot.Send(msg)
	if err != nil {
		return err
	}
	return
}
