package testmodule

import (
	logger "cloudsweep/logging"
	"sync"
	"testing"
)

func TestDefaultLogger(t *testing.T) {
	module := NewModule(logger.NewDefaultLogger())
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
	module := NewModule(logger.NewLogger(&logger.LoggerConfig{LogFilePath: logName, LogFormat: logFormat}))
	module.logDemoRotation()
}
