package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.SugaredLogger

// Init initializes the logger
func Init(level, format, output, filePath string) {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	var writeSyncer zapcore.WriteSyncer
	if output == "file" && filePath != "" {
		file, _ := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		writeSyncer = zapcore.AddSync(file)
	} else {
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	core := zapcore.NewCore(encoder, writeSyncer, zapLevel)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	Log = logger.Sugar()
}

// Debug logs a debug message
func Debug(args ...interface{}) { Log.Debug(args...) }

// Info logs an info message
func Info(args ...interface{}) { Log.Info(args...) }

// Warn logs a warning message
func Warn(args ...interface{}) { Log.Warn(args...) }

// Error logs an error message
func Error(args ...interface{}) { Log.Error(args...) }

// Fatal logs a fatal message and exits
func Fatal(args ...interface{}) { Log.Fatal(args...) }

// Debugf logs a formatted debug message
func Debugf(template string, args ...interface{}) { Log.Debugf(template, args...) }

// Infof logs a formatted info message
func Infof(template string, args ...interface{}) { Log.Infof(template, args...) }

// Warnf logs a formatted warning message
func Warnf(template string, args ...interface{}) { Log.Warnf(template, args...) }

// Errorf logs a formatted error message
func Errorf(template string, args ...interface{}) { Log.Errorf(template, args...) }

// Fatalf logs a formatted fatal message and exits
func Fatalf(template string, args ...interface{}) { Log.Fatalf(template, args...) }
