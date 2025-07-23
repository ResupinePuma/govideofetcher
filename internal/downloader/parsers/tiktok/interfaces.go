package tiktok

import "context"

type INotify interface {
	UpdTextNotify(text string) (err error)
	MakeProgressBar(percent float64) (err error)
}

type Logger interface {
	Errorf(ctx context.Context, format string, v ...any)
	Warnf(ctx context.Context, format string, v ...any)
	Infof(ctx context.Context, format string, v ...any)
	Debugf(ctx context.Context, format string, v ...any)
}
