package main

import (
	"videofetcher/internal/bot"
	"videofetcher/internal/config"
	"videofetcher/internal/logging"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func main() {
	z, _ := zap.NewProduction()
	Logger := logging.NewLogger(*z.Sugar())
	bot.Logging = Logger

	cfg, err := config.NewConfig()
	if err != nil {
		if err.Error() == "error reading config.yml" {
			config.NewDefaultConfig()
		}

		Logger.Panic(err)
	}

	tgbotapi.SetLogger(Logger)

	api, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		Logger.Panic(err)
	}

	cfg.DownloaderOpts.AdminID = cfg.AdminId
	api.Debug = cfg.Debug

	Logger.Infof("Authorized on account %s", api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := api.GetUpdatesChan(u)

	b := bot.TelegramBot{
		Options: cfg.DownloaderOpts,
	}
	err = b.Inititalize(api)
	if err != nil {
		Logger.Panic(err)
	}

	for update := range updates {
		if update.Message != nil { // If we got a message
			Logger.Infof("received message: %s", update.Message.Text)
			Logger.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			b.ProcessMessage(*update.Message)
		}
	}
}
