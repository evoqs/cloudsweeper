package logger

import (
	"os"

	"golang.org/x/exp/slog"
)

type SlogLogger struct {
	*slog.Logger
}

func NewSlogLogger() *SlogLogger {
	l := slog.New(slog.NewTextHandler(os.Stdout, nil))
	// TODO: Add your custom logger configuration here
	return &SlogLogger{l}
}

func (l *SlogLogger) Debugf(format string, args ...interface{}) {
	l.Info(format, args...)
}

func (l *SlogLogger) Infof(format string, args ...interface{}) {
	l.Info(format, args...)
}

func (l *SlogLogger) Errorf(format string, args ...interface{}) {
	l.Error(format, args...)
}

// TODO: Implement other log methods
