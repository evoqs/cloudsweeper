package logger

import (
	"cloudsweep/config"
	"sync"
	"testing"
)

type Module struct {
	logger Logger
}

func NewModule(log Logger) *Module {
	return &Module{logger: log}
}

func (m *Module) logDemoSimple() {
	m.logger.Debugf("This is a Debugf log message")
	m.logger.Infof("This is a Infof log message")
	m.logger.Warnf("This is a Warnf log message")
	m.logger.Errorf("This is a Errorf log message")
	m.logger.Debug("This is a Debug log message")
	m.logger.Info("This is a Info log message")
	m.logger.Warnf("This is a Warn log message")
	m.logger.Errorf("This is a Error log message")
}

func (m *Module) logDemoRotation() {
	for i := 0; i < 20000; i++ {
		m.logger.Debugf("This is a Debug log message")
		m.logger.Infof("This is a Info log message")
		m.logger.Warnf("This is a Warn log message")
		m.logger.Errorf("This is a Error log message")
		m.logger.Debugf("This is a Debug log message")
		m.logger.Infof("This is a Info log message")
		m.logger.Warnf("This is a Warn log message")
		m.logger.Errorf("This is a Error log message")
	}
}

func TestDefaultLogger(t *testing.T) {
	config.LoadConfig()

	module := NewModule(NewDefaultLogger())
	module.logDemoSimple()
	module.logDemoRotation()
}
func TestCustomLogger(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(2)
	go customLogPasser(&wg, "cs_run_text.log", "text")
	go customLogPasser(&wg, "cs_run_json.log", "json")
	wg.Wait()
}

func customLogPasser(wg *sync.WaitGroup, logName string, logFormat string) {
	defer wg.Done()
	module := NewModule(NewLogger(&LoggerConfig{LogFilePath: logName, LogFormat: logFormat}))
	module.logDemoRotation()
}
