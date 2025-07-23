package logging

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/samber/lo"
	"go.uber.org/zap"
)

type Logger struct {
	zap.SugaredLogger
}

func NewLogger(z zap.SugaredLogger) *Logger {
	return &Logger{SugaredLogger: z}
}

func (l *Logger) Println(v ...interface{}) {
	l.SugaredLogger.Infoln(v...)
}

var ingnoredEndpoints []string = []string{"editMessageText", "getUpdates", "getMe"}

func (l *Logger) Printf(format string, v ...interface{}) {
	switch {
	case strings.HasPrefix(format, "Endpoint: %s, params: %v"):
		if lo.Contains(ingnoredEndpoints, fmt.Sprint(v[0])) {
			return
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(fmt.Sprint(v[1])), &parsed); err == nil {
			l.SugaredLogger.Debugw("tg request", "endpoint", v[0], "request", parsed)
		} else {
			l.SugaredLogger.Debugw("tg request", "endpoint", v[0], "raw_request", v[1])
		}
		return
	case strings.HasPrefix(format, "Endpoint: %s, response: %s"):
		if lo.Contains(ingnoredEndpoints, fmt.Sprint(v[0])) {
			return
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(fmt.Sprint(v[1])), &parsed); err == nil {
			l.SugaredLogger.Debugw("tg response", "endpoint", v[0], "response", parsed)
		} else {
			l.SugaredLogger.Debugw("tg response", "endpoint", v[0], "raw_response", v[1])
		}
		return
	}
	l.SugaredLogger.Debugf(format, v...)
}

func (l *Logger) Print(v ...interface{}) {
	l.SugaredLogger.Info(v...)
}

func (l *Logger) Errorf(format string, v ...any) {
	l.SugaredLogger.Errorf(format, v...)
}
func (l *Logger) Warnf(format string, v ...any) {
	l.SugaredLogger.Warnf(format, v...)
}
func (l *Logger) Infof(format string, v ...any) {
	l.SugaredLogger.Infof(format, v...)
}
func (l *Logger) Debugf(format string, v ...any) {
	l.SugaredLogger.Debugf(format, v...)
}

func (l *Logger) Fatalf(format string, v ...any) {}
