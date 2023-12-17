package cost_estimator

import (
	"fmt"
	"testing"

	"cloudsweep/config"
	aws_model "cloudsweep/model/aws"
	"cloudsweep/storage"
	"cloudsweep/utils"
)

func preSetUp() {
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
}

func TestCostProviderInstance(t *testing.T) {
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
	product := aws_model.ProductInfo[aws_model.ProductAttributesInstance]{
		Attributes: aws_model.ProductAttributesInstance{
			InstanceFamily:  "",
			InstanceType:    "t2.medium",
			Memory:          "",
			RegionCode:      "us-east-1",
			OperatingSystem: "Linux",
		},
	}

	cost, err := GetComputeInstanceCost(product)
	if err != nil {
		fmt.Printf("Error %v", err)
	}
	fmt.Printf("Cost: %v \n", cost)
}

func TestCostProviderEbs(t *testing.T) {
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
	product := aws_model.ProductInfo[aws_model.ProductAttributesEBS]{
		Attributes: aws_model.ProductAttributesEBS{
			StorageMedia:  "SSD-backed",
			RegionCode:    "us-east-1",
			VolumeApiName: "gp2",
		},
		ProductFamily: "",
	}

	cost, err := GetEbsCost(product)
	if err != nil {
		fmt.Printf("Error %v", err)
	}
	fmt.Printf("Cost: %v \n", cost)
}

func TestCostProviderEbsSnapshot(t *testing.T) {
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
	product := aws_model.ProductInfo[aws_model.ProductAttributesEBSSnapshot]{
		Attributes: aws_model.ProductAttributesEBSSnapshot{
			StorageMedia: "Amazon S3",
			RegionCode:   "us-east-1",
		},
		ProductFamily: "Storage Snapshot",
	}

	cost, err := GetEbsSnapshotCost(product)
	if err != nil {
		fmt.Printf("Error %v", err)
	}
	fmt.Printf("Cost: %v \n", cost)
}

func TestCostProviderElasticIp(t *testing.T) {
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
	product := aws_model.ProductInfo[aws_model.ProductAttributesElasticIp]{
		Attributes: aws_model.ProductAttributesElasticIp{
			RegionCode: "us-east-1",
		},
		ProductFamily: "ElasticIP",
	}

	cost, err := GetElasticIpCost(product)
	if err != nil {
		fmt.Printf("Error %v", err)
	}
	fmt.Printf("Cost: %v \n", cost)
}
