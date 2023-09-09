package scheduler

import "testing"

func TestDefaultLogger(t *testing.T) {
	ps := StartPipelineScheduler()
	ps.schedulePipeline()
}
