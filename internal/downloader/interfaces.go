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
	Errorf(format string, v ...any)
	Warnf(format string, v ...any)
	Infof(format string, v ...any)
	Debugf(format string, v ...any)
}
