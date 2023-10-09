package cost_estimator

import (
	"cloudsweep/storage"
	"cloudsweep/utils"

	"fmt"
	"testing"
	"time"
)

func TestCostCleaner(t *testing.T) {
	cfg := utils.LoadConfig()
	dbUrl, err := utils.GetDBUrl(&cfg)

	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(dbUrl)
	}

	dbM := storage.GetDBManager()
	dbM.SetDbUrl(dbUrl)
	dbM.SetDatabase(utils.GetConfig().Database.Name)

	_, err = dbM.Connect()
	if err != nil {
		fmt.Println("Failed to connect to DB " + err.Error())
	}
	err = dbM.CheckConnection()
	if err != nil {
		fmt.Println("Connection Check failed")
	} else {
		fmt.Println("Successfully Connected")
		defer dbM.Disconnect()
	}
	storage.MakeDBOperators(dbM)

	//var schedulerStore scheduler.SchedulerStore
	/*schedulerStore := &scheduler.SchedulerStore{
		Schedulers: make(map[string]*scheduler.Scheduler),
		Log:        logger.NewDefaultLogger(),
	}
	costScheduler, _ := schedulerStore.CreateScheduler("Cost Cleaner", logger.NewDefaultLogger())
	costScheduler.AddCron("cost_cleaner", "* * * * *", func() { CleanOldResourceCosts(24 * time.Hour) })*/

	opr := storage.GetDefaultDBOperators()
	before, err := opr.CostOperator.GetQueryResultCount("{}")
	if err != nil {
		fmt.Printf("Error: %v", err)
	}
	fmt.Printf("Total number of Cost Entries %d\n", before)

	StartResourceCostPurger()
	//CleanOldResourceCosts(2 * time.Minute)

	time.Sleep(90 * time.Second)
	after, err := opr.CostOperator.GetQueryResultCount("{}")
	if err != nil {
		fmt.Printf("Error: %v", err)
	}
	fmt.Printf("Total number of Cost Entries Deleted: %d\n", before-after)
	fmt.Printf("Done")
}
