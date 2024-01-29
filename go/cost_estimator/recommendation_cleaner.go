package cost_estimator

import (
	logger "cloudsweep/logging"
	"cloudsweep/scheduler"
	"cloudsweep/storage"
	"fmt"
	"sync"
	"time"
)

var rcmdtnSchedulerStore *scheduler.SchedulerStore

func init() {
	// TODO: Implement the lazy construction
	rcmdtnSchedulerStore = &scheduler.SchedulerStore{
		Schedulers:  make(map[string]*scheduler.Scheduler),
		ScheduleMux: sync.Mutex{},
		Log:         logger.NewDefaultLogger(),
	}
}

func CleanOldRecommendations(maxAge time.Duration) error {
	logger.NewDefaultLogger().Infof("Deleting Recommendations over the age of %v", maxAge)
	oldestTimestamp := time.Now().Add(-maxAge)

	opr := storage.GetDefaultDBOperators()
	count, err := opr.RecommendationOperator.DeleteOldRecommendations(oldestTimestamp)
	if err != nil {
		return fmt.Errorf("Problem in Deleting Older Recommendations: %v", err)
	}
	logger.NewDefaultLogger().Infof("Total number of Old Recommendations purged: %v", count.DeletedCount)
	return nil
}

func StartRecommendationPurger() {
	recomScheduler, _ := schedulerStore.CreateScheduler("RecommendationCleaner", logger.NewDefaultLogger())
	// Run everyday, delete over a week old data
	//costScheduler.AddCron("recommendation_cleaner", "0 0 * * *", func() { CleanOldRecommendations(24 * 7 * time.Hour) })
	recomScheduler.AddCron("cost_cleaner", "* * * * *", func() { CleanOldRecommendations(2 * time.Minute) })
	recomScheduler.StartScheduler()
}

func StopRecommendationPurger() {
	recomScheduler, err := schedulerStore.GetScheduler("RecommendationCleaner")
	if err != nil {
		logger.NewDefaultLogger().Errorf("Problem in stopping the scheduler \" RecommendationCleaner \"")
	}
	err = recomScheduler.DeleteCron("recommendation_cleaner")
	if err != nil {
		logger.NewDefaultLogger().Errorf("Problem in deleting the cron \" recommendation_cleaner \"")
	}
	recomScheduler.StopScheduler()
	schedulerStore.DeleteScheduler(recomScheduler.Name)
}
