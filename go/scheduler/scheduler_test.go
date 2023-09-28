package scheduler

import (
	logger "cloudsweep/logging"
	"fmt"
	"testing"
	"time"

	"github.com/go-co-op/gocron"
)

func TestScheduler(t *testing.T) {
	scheduler := Scheduler{Name: "TestOneScheduler",
		taskMap:   make(map[string]Task),
		scheduler: gocron.NewScheduler(time.UTC),
		log:       logger.NewDefaultLogger(),
	}
	scheduler.AddCron("CronOne", "*/5 * * * *", func() { fmt.Println("Scheduled call") })
	scheduler.StartScheduler()
	time.Sleep(20 * time.Second)
	scheduler.StopScheduler()
}
