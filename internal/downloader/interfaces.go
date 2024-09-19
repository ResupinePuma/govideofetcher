package downloader

import (
	"context"
	"net/url"
	"videofetcher/internal/downloader/dcontext"
	"videofetcher/internal/downloader/video"
)

type AbstractDownloader interface {
	Download(ctx dcontext.Context, u *url.URL) ([]video.Video, error)
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
