package bot

import "context"

type Logger interface {
	Print(v ...any)
	Errorf(ctx context.Context, format string, v ...any)
	Errorw(ctx context.Context, msg string, keysAndVals ...any)
	Warnf(ctx context.Context, format string, v ...any)
	Infof(ctx context.Context, format string, v ...any)
	Infow(ctx context.Context, msg string, keysAndVals ...any)
	Debugf(ctx context.Context, format string, v ...any)
}
