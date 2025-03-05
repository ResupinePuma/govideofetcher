package instagram

type INotify interface {
	UpdTextNotify(text string) (err error)
	MakeProgressBar(percent float64) (err error)
}

type Logger interface {
	Errorf(format string, v ...any)
	Warnf(format string, v ...any)
	Infof(format string, v ...any)
	Debugf(format string, v ...any)
}
