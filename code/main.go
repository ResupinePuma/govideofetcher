package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	logger := InitializeLogger()

	cfg, err := NewConfig()
	if err != nil {
		if err.Error() == "error reading config.yml" {
			NewDefaultConfig()
		}

		log.Panic(err)
	}

	tgbotapi.SetLogger(&logger)

	bot, err := tgbotapi.NewBotAPI(cfg.Base.Token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	mW := NewMsgWorker(&logger, bot, cfg)

	for update := range updates {
		if update.Message != nil { // If we got a message
			logger.Info("received message: %s", update.Message.Text)
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			mW.Process(*update.Message)
		}
	}
}
