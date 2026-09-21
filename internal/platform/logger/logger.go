package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

// Logger is a compatibility facade over the standard library slog logger.
type Logger struct {
	*slog.Logger
}

var Log = NewNop()

// NewNop returns a logger that discards all output.
func NewNop() *Logger {
	return &Logger{slog.New(slog.NewTextHandler(io.Discard, nil))}
}

// Init configures the process logger.
func Init(level, format, output, filePath string) {
	var writer io.Writer = os.Stdout
	if output == "file" && filePath != "" {
		file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err == nil {
			writer = file
		}
	}

	logLevel := slog.LevelInfo
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	}

	options := &slog.HandlerOptions{Level: logLevel}
	var handler slog.Handler
	if format == "json" {
		handler = slog.NewJSONHandler(writer, options)
	} else {
		handler = slog.NewTextHandler(writer, options)
	}
	Log = &Logger{slog.New(handler)}
}

func (l *Logger) Infof(template string, args ...interface{}) {
	l.Info(fmt.Sprintf(template, args...))
}

func (l *Logger) Warnf(template string, args ...interface{}) {
	l.Warn(fmt.Sprintf(template, args...))
}

func (l *Logger) Errorf(template string, args ...interface{}) {
	l.Error(fmt.Sprintf(template, args...))
}

func (l *Logger) Debugf(template string, args ...interface{}) {
	l.Debug(fmt.Sprintf(template, args...))
}

func (l *Logger) Infow(msg string, args ...interface{}) {
	l.Info(msg, args...)
}

func (l *Logger) Errorw(msg string, args ...interface{}) {
	l.Error(msg, args...)
}

func Debug(args ...interface{}) { Log.Debug(fmt.Sprint(args...)) }
func Info(args ...interface{})  { Log.Info(fmt.Sprint(args...)) }
func Warn(args ...interface{})  { Log.Warn(fmt.Sprint(args...)) }
func Error(args ...interface{}) { Log.Error(fmt.Sprint(args...)) }
func Fatal(args ...interface{}) { Log.Error(fmt.Sprint(args...)); os.Exit(1) }

func Debugf(template string, args ...interface{}) { Log.Debugf(template, args...) }
func Infof(template string, args ...interface{})  { Log.Infof(template, args...) }
func Warnf(template string, args ...interface{})  { Log.Warnf(template, args...) }
func Errorf(template string, args ...interface{}) { Log.Errorf(template, args...) }
func Fatalf(template string, args ...interface{}) {
	Log.Error(fmt.Sprintf(template, args...))
	os.Exit(1)
}
