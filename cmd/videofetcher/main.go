package main

import (
	"videofetcher/internal/bot"
	"videofetcher/internal/config"
	"videofetcher/internal/downloader"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func main() {
	z, _ := zap.NewProduction()
	Logger := bot.NewLogger(*z.Sugar())
	bot.Logging = Logger

	cfg, err := config.NewConfig()
	if err != nil {
		if err.Error() == "error reading config.yml" {
			config.NewDefaultConfig()
		}

		Logger.Panic(err)
	}

	tgbotapi.SetLogger(Logger)

	api, err := tgbotapi.NewBotAPI(cfg.Base.Token)
	if err != nil {
		Logger.Panic(err)
	}

	api.Debug = cfg.Base.Debug

	Logger.Infof("Authorized on account %s", api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := api.GetUpdatesChan(u)

	b := bot.TelegramBot{
		Options: downloader.DownloaderOpts{
			Timeout:   cfg.Base.Timeout,
			SizeLimit: cfg.Base.SizeLimit,
		},
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
