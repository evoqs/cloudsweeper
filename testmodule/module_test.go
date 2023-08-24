package testmodule

import (
	logger "cloudsweep/logging"
	"testing"
)

func TestModule(t *testing.T) {
	moduel := NewModule(logger.NewSlogLogger())
	moduel.someFunction()
}
