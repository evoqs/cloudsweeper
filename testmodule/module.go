package testmodule

import (
	logger "cloudsweep/logging"
)

type Module struct {
	logger logger.Logger
}

func NewModule(log logger.Logger) *Module {
	return &Module{logger: log}
}

func (m *Module) logDemoSimple() {
	m.logger.Debugf("This is a Debug log message")
	m.logger.Infof("This is a Info log message")
	m.logger.Warnf("This is a Warn log message")
	m.logger.Errorf("This is a Error log message")
}

func (m *Module) logDemoRotation() {
	for i := 0; i < 20000; i++ {
		m.logger.Debugf("This is a Debug log message")
		m.logger.Infof("This is a Info log message")
		m.logger.Warnf("This is a Warn log message")
		m.logger.Errorf("This is a Error log message")
	}
}
