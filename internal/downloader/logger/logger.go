package logger

import "log"

type Logger struct{}

func (l *Logger) Println(v ...interface{}) {
	log.Println(v...)
}
func (l *Logger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}
func (l *Logger) Errorf(format string, v ...any) {
	log.Printf("[ERROR] "+format, v...)
}
func (l *Logger) Warnf(format string, v ...any) {
	log.Printf("[WARN] "+format, v...)
}
func (l *Logger) Infof(format string, v ...any) {
	log.Printf("[INFO] "+format, v...)
}
func (l *Logger) Debugf(format string, v ...any) {
	log.Printf("[DEBUG] "+format, v...)
}
