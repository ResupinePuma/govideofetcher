package logging

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

func (l *Logger) Errorf(format string, v ...any) {
	l.SugaredLogger.Errorf("[ERROR] "+format, v...)
}
func (l *Logger) Warnf(format string, v ...any) {
	l.SugaredLogger.Warnf("[WARN] "+format, v...)
}
func (l *Logger) Infof(format string, v ...any) {
	l.SugaredLogger.Infof("[INFO] "+format, v...)
}
func (l *Logger) Debugf(format string, v ...any) {
	l.SugaredLogger.Debugf("[DEBUG] "+format, v...)
}

func (l *Logger) Fatalf (format string, v ...any) { }

