package logger

type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Debug(message string)
	Info(message string)
	Warn(message string)
	Error(message string)
}
