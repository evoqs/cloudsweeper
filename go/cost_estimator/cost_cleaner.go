package cost_estimator

import (
	logger "cloudsweep/logging"
	"cloudsweep/scheduler"
	"cloudsweep/storage"
	"fmt"
	"sync"
	"time"
)

var schedulerStore *scheduler.SchedulerStore

func init() {
	// TODO: Implement the lazy construction
	schedulerStore = &scheduler.SchedulerStore{
		Schedulers:  make(map[string]*scheduler.Scheduler),
		ScheduleMux: sync.Mutex{},
		Log:         logger.NewDefaultLogger(),
	}
}
func CleanOldResourceCosts(maxAge time.Duration) error {
	logger.NewDefaultLogger().Infof("Deleting Resource Costs over the age of %v", maxAge)
	oldestTimestamp := time.Now().Add(-maxAge)

	opr := storage.GetDefaultDBOperators()
	count, err := opr.CostOperator.DeleteOldResourceCosts(oldestTimestamp)
	if err != nil {
		return fmt.Errorf("Problem in Deleting Older Resource Costs: %v", err)
	}
	logger.NewDefaultLogger().Infof("Total number of Old Resource Costs purged: %v", count.DeletedCount)
	return nil
}

func StartResourceCostPurger() {
	schedulerStore := &scheduler.SchedulerStore{
		Schedulers: make(map[string]*scheduler.Scheduler),
		Log:        logger.NewDefaultLogger(),
	}
	costScheduler, _ := schedulerStore.CreateScheduler("CostCleaner", logger.NewDefaultLogger())
	// Run everyday, delete over a week old data
	costScheduler.AddCron("cost_cleaner", "0 0 * * *", func() { CleanOldResourceCosts(24 * 7 * time.Hour) })
	//costScheduler.AddCron("cost_cleaner", "* * * * *", func() { CleanOldResourceCosts(2 * time.Minute) })
	costScheduler.StartScheduler()
}

func StopResourceCostPurger() {
	scheduler, err := schedulerStore.GetScheduler("CostCleaner")
	if err != nil {
		logger.NewDefaultLogger().Errorf("Problem in stopping the scheduler \" CostCleaner \"")
	}
	err = scheduler.DeleteCron("cost_cleaner")
	if err != nil {
		logger.NewDefaultLogger().Errorf("Problem in deleting the cron \" cost_cleaner \"")
	}
	scheduler.StopScheduler()
}
