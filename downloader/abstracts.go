package downloader

import (
	"context"
	"io"
)

type AbstractLogger interface {
	Error(format string, v ...any)
	Warning(format string, v ...any)
	Info(format string, v ...any)
	Debug(format string, v ...any)
}

type AbstractDownloader interface {
	Init(loggger AbstractLogger, notifier AbstractNotifier, opts *DownloaderOpts) error
	Download(ctx context.Context, url string) (string, io.ReadCloser, error)
	Close() error
}

type AbstractNotifier interface {
	Count(percent float64) error
	Message(text string) error
}

type DownloaderOpts struct {
	SizeLimit int // Max allowed size
	Timeout   int
}
