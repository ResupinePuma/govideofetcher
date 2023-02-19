package main

import (
	"go.uber.org/zap"
)

type VFLogger struct {
	s zap.SugaredLogger
}

func (l *VFLogger) Println(v ...interface{}) {
	l.s.Infoln(v...)
}

func (l *VFLogger) Printf(format string, v ...interface{}) {
	l.s.Infof(format, v...)
}

func (l *VFLogger) Error(format string, v ...any) {
	l.s.Errorf(format, v...)
}
func (l *VFLogger) Warning(format string, v ...any) {
	l.s.Warnf(format, v...)
}
func (l *VFLogger) Info(format string, v ...any) {
	l.s.Infof(format, v...)
}
func (l *VFLogger) Debug(format string, v ...any) {
	l.s.Debugf(format, v...)
}

func InitializeLogger() VFLogger {
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()
	return VFLogger{s: *sugar}
}
