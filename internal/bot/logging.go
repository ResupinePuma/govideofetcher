package bot

import "go.uber.org/zap"

type Logger struct {
	zap.SugaredLogger
}

func NewLogger(z zap.SugaredLogger) *Logger {
	return &Logger{SugaredLogger: z}
}

func (l *Logger) Println(v ...interface{}) {
	l.SugaredLogger.Infoln(v...)
}

func (l *Logger) Printf(format string, v ...interface{}) {
	l.SugaredLogger.Infof(format, v...)
}

func (l *Logger) Print(v ...interface{}) {
	l.SugaredLogger.Info(v...)
}
