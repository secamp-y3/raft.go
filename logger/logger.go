package logger

import (
	"fmt"
	"log"
	"runtime"
	"time"
)

const (
	level_info  = "INFO"
	level_error = "ERROR"
	level_fatal = "FATAL"
)

// Logger handler
type Logger struct {
	Name string
}

func timestamp() string {
	return time.Now().String()
}

func (l *Logger) display(level, format string, args ...any) {
	log.Println(l.format(level, format, args...))
}

func (l *Logger) format(level, format string, args ...any) string {
	return fmt.Sprintf("[%s] (%s) %s: %s", timestamp(), l.Name, level, fmt.Sprintf(format, args...))
}

func New(name string) *Logger {
	return &Logger{Name: name}
}

// Info displays basic information
func (l *Logger) Info(format string, args ...any) {
	l.display(level_info, format, args...)
}

// Error displays error message.
//
// Unlike Logger#Fatal, this function does not cause panic.
func (l *Logger) Error(format string, args ...any) {
	l.display(level_error, format, args...)
}

// Fatal displays error message, then causes panic.
func (l *Logger) Fatal(format string, args ...any) {
	l.display(level_fatal, format, args...)
	pc, file, line, ok := runtime.Caller(1)
	if ok {
		panic(fmt.Sprintf("PANIC [%s] (%s): {%s, %d, %s}", timestamp(), l.Name, file, line, runtime.FuncForPC(pc)))
	} else {
		panic(fmt.Sprintf("!!PANIC!! [%s] (%s)", timestamp(), l.Name))
	}
}
