package downloader

import (
	"context"
	"videofetcher/internal/downloader/dcontext"
)

type AbstractDownloader interface {
	Download(ctx *dcontext.Context) error
	Close() error
}

type AbstractNotifier interface {
	SendNotify(text string) (err error)
	UpdTextNotify(text string) (err error)
	StartTicker(ctx context.Context) (err error)
}

type iLogger interface {
	Errorf(ctx context.Context, format string, v ...any)
	Warnf(ctx context.Context, format string, v ...any)
	Infof(ctx context.Context, format string, v ...any)
	Debugf(ctx context.Context, format string, v ...any)
}
