package main

import (
	"context"
	"strconv"
	"videofetcher/internal/bot"
	"videofetcher/internal/config"
	"videofetcher/internal/downloader"
	"videofetcher/internal/logging"
	"videofetcher/internal/telemetry"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

func main() {
	zapCFG := zap.NewProductionConfig()
	zapCFG.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	zapCFG.DisableCaller = true
	zapCFG.Encoding = "json"
	z, _ := zapCFG.Build()
	Logger := logging.NewLogger(*z.Sugar())
	bot.Logging = Logger
	downloader.Logger = Logger

	cfg, err := config.NewConfig()
	if err != nil {
		if err.Error() == "error reading config.yml" {
			config.NewDefaultConfig()
		}

		Logger.Panic(err)
	}

	telemetry.ServiceName = cfg.Telemetry.ServiceName
	telemetry.TracerEndpoint = cfg.Telemetry.Address

	ctx := context.Background()
	telemetry.InitTracer(ctx)

	tgbotapi.SetLogger(Logger)

	api, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		Logger.Panic(err)
	}

	cfg.DownloaderOpts.AdminID = cfg.AdminId
	api.Debug = cfg.Debug

	Logger.Infof(ctx, "Authorized on account %s", api.Self.UserName)

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

	tracer := otel.Tracer(telemetry.ServiceName)

	for update := range updates {
		if update.Message != nil { // If we got a message

			ctx, span := tracer.Start(ctx, "HandleTelegramMessage")
			span.SetAttributes(
				attribute.String("from.user_id", strconv.Itoa(int(update.Message.From.ID))),
				attribute.String("message.text", update.Message.Text),
			)

			Logger.Infof(ctx, "[%s] received message: %s", update.Message.From.UserName, update.Message.Text)
			go func(update tgbotapi.Update, ctx context.Context) {
				defer span.End()

				b.ProcessMessage(ctx, *update.Message)
			}(update, ctx)
		}
	}
}
