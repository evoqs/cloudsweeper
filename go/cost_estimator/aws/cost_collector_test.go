package cost_estimator

import (
	"fmt"
	"testing"

	aws_model "cloudsweep/model/aws"
	"cloudsweep/storage"
	"cloudsweep/utils"
)

func preSetUp() {
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
}

func TestCostProviderInstance(t *testing.T) {
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
	product := aws_model.ProductAttributesInstance{
		InstanceFamily:  "",
		InstanceType:    "t3.large",
		Memory:          "",
		RegionCode:      "us-east-1",
		OperatingSystem: "Linux",
	}

	cost, err := GetComputeInstanceCost(product)
	if err != nil {
		fmt.Printf("Error %v", err)
	}
	fmt.Printf("Cost: %v \n", cost)
}

func TestCostProviderEbs(t *testing.T) {
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
	product := aws_model.ProductAttributesEBS{
		StorageMedia:  "SSD-backed",
		RegionCode:    "us-east-1",
		VolumeApiName: "gp2",
	}

	cost, err := GetEbsCost(product)
	if err != nil {
		fmt.Printf("Error %v", err)
	}
	fmt.Printf("Cost: %v \n", cost)
}
