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

func (m *Module) someFunction() {
	m.logger.Infof("This is a log message")
	m.logger.Errorf("This is a log message")
	m.logger.Debugf("This is a log message")
}
