package ytdl

type INotify interface {
	UpdTextNotify(text string) (err error)
	MakeProgressBar(percent float64) (err error)
}

type Logger interface {
	Print(v ...any)
	Errorf(format string, v ...any)
	Warnf(format string, v ...any)
	Infof(format string, v ...any)
	Debugf(format string, v ...any)
}
