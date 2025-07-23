package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"videofetcher/internal/telemetry"

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

func (l *Logger) Errorf(ctx context.Context, format string, v ...any) {
	id := telemetry.TraceID(ctx)
	if id != "none" {
		l.SugaredLogger.Errorw(fmt.Sprintf(format, v...), "trace_id", telemetry.TraceID(ctx))
	} else {
		l.SugaredLogger.Errorf(format, v...)
	}
}
func (l *Logger) Errorw(ctx context.Context, msg string, keysAndValues ...any) {
	id := telemetry.TraceID(ctx)
	if id != "none" {
		keysAndValues = append(keysAndValues, "trace_id", telemetry.TraceID(ctx))
		l.SugaredLogger.Errorw(msg, keysAndValues...)
	} else {
		l.SugaredLogger.Errorw(msg, keysAndValues...)
	}
}
func (l *Logger) Warnf(ctx context.Context, format string, v ...any) {
	id := telemetry.TraceID(ctx)
	if id != "none" {
		l.SugaredLogger.Warnw(fmt.Sprintf(format, v...), "trace_id", telemetry.TraceID(ctx))
	} else {
		l.SugaredLogger.Warnf(format, v...)
	}
}
func (l *Logger) Infof(ctx context.Context, format string, v ...any) {
	id := telemetry.TraceID(ctx)
	if id != "none" {
		l.SugaredLogger.Infow(fmt.Sprintf(format, v...), "trace_id", telemetry.TraceID(ctx))
	} else {
		l.SugaredLogger.Infof(format, v...)
	}
}
func (l *Logger) Infow(ctx context.Context, msg string, keysAndValues ...any) {
	id := telemetry.TraceID(ctx)
	if id != "none" {
		keysAndValues = append(keysAndValues, "trace_id", telemetry.TraceID(ctx))
		l.SugaredLogger.Infow(msg, keysAndValues...)
	} else {
		l.SugaredLogger.Infow(msg, keysAndValues...)
	}
}
func (l *Logger) Debugf(ctx context.Context, format string, v ...any) {
	id := telemetry.TraceID(ctx)
	if id != "none" {
		l.SugaredLogger.Debugw(fmt.Sprintf(format, v...), "trace_id", telemetry.TraceID(ctx))
	} else {
		l.SugaredLogger.Debugf(format, v...)
	}
}

func (l *Logger) Fatalf(format string, v ...any) {}
