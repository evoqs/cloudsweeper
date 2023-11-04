package cost_estimator

import (
	"fmt"
	"testing"

	"cloudsweep/config"
	aws_model "cloudsweep/model/aws"
	"cloudsweep/storage"
	"cloudsweep/utils"
)

func TestRecommenderInstance(t *testing.T) {
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
	product := aws_model.ProductAttributesInstance{
		InstanceFamily:  "",
		InstanceType:    "t3.large",
		Memory:          "",
		RegionCode:      "us-east-1",
		OperatingSystem: "Linux",
	}
	resourceUsageParams := map[string]interface{}{
		"PeakCPUUtilization":    80,
		"AverageCPUUtilization": 4,
	}
	recommendation, err := GetRecommendation(product, resourceUsageParams)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Recommendation: %s\n", recommendation)
}
