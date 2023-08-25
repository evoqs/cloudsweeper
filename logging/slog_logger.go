package logger

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"

	"golang.org/x/exp/slog" // Change it to log/slog in 1.21
)

type SlogLogger struct {
	*slog.Logger
	logFile *lumberjack.Logger
}

var factory = &LoggerFactory{
	loggers: make(map[string]*SlogLogger),
}

type LoggerFactory struct {
	mu      sync.Mutex
	loggers map[string]*SlogLogger
}

type LoggerConfig struct {
	LogFilePath string
	LogFormat   string
}

// TODO: Read all the values from config
func NewDefaultLogger() *SlogLogger {
	logFilePath := "./cs_runner.log"
	factory.mu.Lock()
	defer factory.mu.Unlock()

	// Check if a logger instance for this log file already exists
	if logger, ok := factory.loggers[logFilePath]; ok {
		return logger
	}
	lumberjack := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    1, // megabytes
		MaxBackups: 5,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	}
	replace := func(groups []string, a slog.Attr) slog.Attr {
		// Remove time.
		if a.Key == slog.TimeKey && len(groups) == 0 {
			a.Key = "Date"
			return a
		}
		// Remove the directory from the source's filename.
		if a.Key == slog.SourceKey {
			source := a.Value.Any().(*slog.Source)
			source.File = filepath.Base(source.File)
		}
		return a
	}
	l := slog.New(slog.NewTextHandler(lumberjack, &slog.HandlerOptions{
		Level:       slog.LevelDebug,
		AddSource:   true,
		ReplaceAttr: replace,
	}))
	logger := &SlogLogger{l, lumberjack}
	factory.loggers[logFilePath] = logger
	return logger
}

func NewLogger(config *LoggerConfig) *SlogLogger {
	factory.mu.Lock()
	defer factory.mu.Unlock()
	var l *slog.Logger
	// Check if a logger instance for this log file already exists
	if logger, ok := factory.loggers[config.LogFilePath]; ok {
		return logger
	}
	lumberjack := &lumberjack.Logger{
		Filename:   config.LogFilePath,
		MaxSize:    1, // megabytes
		MaxBackups: 5,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	}
	replace := func(groups []string, a slog.Attr) slog.Attr {
		// Remove time.
		if a.Key == slog.TimeKey && len(groups) == 0 {
			a.Key = "Date"
			return a
		}
		// Remove the directory from the source's filename.
		if a.Key == slog.SourceKey {
			source := a.Value.Any().(*slog.Source)
			source.File = filepath.Base(source.File)
		}
		return a
	}
	switch config.LogFormat {
	case "text":
		l = slog.New(slog.NewTextHandler(lumberjack, &slog.HandlerOptions{
			Level:       slog.LevelDebug,
			AddSource:   true,
			ReplaceAttr: replace,
		}))
	case "json":
		l = slog.New(slog.NewJSONHandler(lumberjack, &slog.HandlerOptions{
			Level:       slog.LevelDebug,
			AddSource:   true,
			ReplaceAttr: replace,
		}))
	}

	logger := &SlogLogger{l, lumberjack}
	factory.loggers[config.LogFilePath] = logger
	return logger
}

func (l *SlogLogger) Close() {
	// TODO: Check for any other cleanup activities
	l.logFile.Close()
}

func (l *SlogLogger) Debugf(format string, args ...interface{}) {
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:])
	r := slog.NewRecord(time.Now(), slog.LevelDebug, fmt.Sprintf(format, args...), pcs[0])
	_ = l.Handler().Handle(context.Background(), r)
}

func (l *SlogLogger) Infof(format string, args ...interface{}) {
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Infof]
	r := slog.NewRecord(time.Now(), slog.LevelInfo, fmt.Sprintf(format, args...), pcs[0])
	_ = l.Handler().Handle(context.Background(), r)
}

func (l *SlogLogger) Warnf(format string, args ...interface{}) {
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:])
	r := slog.NewRecord(time.Now(), slog.LevelWarn, fmt.Sprintf(format, args...), pcs[0])
	_ = l.Handler().Handle(context.Background(), r)
}

func (l *SlogLogger) Errorf(format string, args ...interface{}) {
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:])
	r := slog.NewRecord(time.Now(), slog.LevelError, fmt.Sprintf(format, args...), pcs[0])
	_ = l.Handler().Handle(context.Background(), r)
}
