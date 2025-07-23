package main

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

var regexAnyType = regexp.MustCompile(`^\[.+\]`)

type LogEntry struct {
	Level     string // info, debug, warn, error
	EventType string
	Message   string
	Raw       string
}

func streamLogs(r io.Reader, traceID string, logger *zap.Logger) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		entry := parseYtDlpLogLine(line)
		switch entry.Level {
		case "info":
			logger.Info("yt-dlp", zap.String("event_type", entry.EventType), zap.String("message", entry.Message), zap.String("trace_id", traceID))
		case "debug":
			logger.Debug("yt-dlp", zap.String("event_type", entry.EventType), zap.String("message", entry.Message), zap.String("trace_id", traceID))
		case "error":
			logger.Error("yt-dlp", zap.String("event_type", entry.EventType), zap.String("message", entry.Message), zap.String("trace_id", traceID))
		case "warn":
			logger.Warn("yt-dlp", zap.String("event_type", entry.EventType), zap.String("message", entry.Message), zap.String("trace_id", traceID))
		}

	}
}

// parseYtDlpLogLine превращает строку yt-dlp лога в структуру LogEntry
func parseYtDlpLogLine(line string) LogEntry {
	entry := LogEntry{
		Level:   "info",
		Message: line,
		Raw:     line,
	}

	switch {
	case strings.HasPrefix(line, "[info]"):
		entry.Level = "info"
		entry.Message = strings.TrimPrefix(line, "[info] ")
	case strings.HasPrefix(line, "[debug]"):
		entry.Level = "debug"
		entry.Message = strings.TrimPrefix(line, "[debug] ")
	case strings.HasPrefix(line, "[warning]"):
		entry.Level = "warn"
		entry.Message = strings.TrimPrefix(line, "[warning] ")
	case strings.HasPrefix(line, "[error]"):
		entry.Level = "error"
		entry.Message = strings.TrimPrefix(line, "[error] ")
	case strings.HasPrefix(line, "ERROR:"):
		entry.Level = "error"
		entry.Message = strings.TrimPrefix(line, "ERROR: ")
	case strings.HasPrefix(line, "WARNING:"):
		entry.Level = "error"
		entry.Message = strings.TrimPrefix(line, "WARNING: ")
	default:
		entry.Level = "info"
		entry.EventType = regexAnyType.FindString(line)
		entry.EventType = strings.Replace(entry.EventType, "[", "", 1)
		entry.EventType = strings.Replace(entry.EventType, "]", "", 1)
		entry.Message = regexAnyType.ReplaceAllString(line, "")
	}

	entry.Message = strings.TrimSpace(entry.Message)

	return entry
}
