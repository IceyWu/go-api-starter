package logger

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.SugaredLogger

// customTimeEncoder formats time as "2006-01-16 15:04:05"
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

// customLevelEncoder formats level with color and fixed width
func customLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var levelStr string
	switch level {
	case zapcore.DebugLevel:
		levelStr = "\033[36mDEBUG\033[0m"
	case zapcore.InfoLevel:
		levelStr = "\033[32mINFO \033[0m"
	case zapcore.WarnLevel:
		levelStr = "\033[33mWARN \033[0m"
	case zapcore.ErrorLevel:
		levelStr = "\033[31mERROR\033[0m"
	default:
		levelStr = fmt.Sprintf("%-5s", level.CapitalString())
	}
	enc.AppendString(levelStr)
}

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
		CallerKey:      "",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    customLevelEncoder,
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if format == "json" {
		encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
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
	logger := zap.New(core)
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
