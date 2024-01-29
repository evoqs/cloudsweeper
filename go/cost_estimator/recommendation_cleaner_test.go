package cost_estimator

import (
	"cloudsweep/config"
	"cloudsweep/storage"
	"cloudsweep/utils"
	"time"

	"fmt"
	"testing"
)

func TestRecommendationCleaner(t *testing.T) {
	cfg := config.LoadConfig()
	dbUrl, err := utils.GetDBUrl(&cfg)

	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(dbUrl)
	}

	dbM := storage.GetDBManager()
	dbM.SetDbUrl(dbUrl)
	dbM.SetDatabase(config.GetConfig().Database.Name)

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

	opr := storage.GetDefaultDBOperators()
	before, err := opr.RecommendationOperator.GetQueryResultCount("{}")
	if err != nil {
		fmt.Printf("Error: %v", err)
	}
	fmt.Printf("Total number of Recommendation Entries %d\n", before)

	StartRecommendationPurger()
	//CleanOldRecommendations(2 * time.Minute)

	time.Sleep(90 * time.Second)
	after, err := opr.RecommendationOperator.GetQueryResultCount("{}")
	if err != nil {
		fmt.Printf("Error: %v", err)
	}
	fmt.Printf("Total number of Recommendation Entries Deleted: %d\n", before-after)
	fmt.Printf("Done")
}
