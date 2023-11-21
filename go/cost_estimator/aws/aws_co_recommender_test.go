package cost_estimator

import (
	"fmt"
	"testing"

	"cloudsweep/config"
	"cloudsweep/storage"
	"cloudsweep/utils"
)

func TestCOAwsResultAllInstances(t *testing.T) {
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

	recommendations, err := GetAWSCOResultForAllEC2Instances(config.GetConfig().Aws.Creds.Aws_access_key_id, config.GetConfig().Aws.Creds.Aws_secret_access_key, []string{config.GetConfig().Aws.Creds.Aws_default_region, "ap-northeast-1", "ap-southeast-2"})
	if err != nil {
		t.Errorf("Error %v", err)
	}
	t.Logf("Total Number of recommendations: %d", len(recommendations))
	for _, recommendation := range recommendations {
		t.Logf("Recommendation: ====================== %v\n", recommendation)
	}
}

func TestCOAllInstances(t *testing.T) {
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

	recommendations, err := GetAWSRecommendationForAllEC2Instances(config.GetConfig().Aws.Creds.Aws_access_key_id, config.GetConfig().Aws.Creds.Aws_secret_access_key, []string{config.GetConfig().Aws.Creds.Aws_default_region, "ap-northeast-1", "ap-southeast-2"})
	if err != nil {
		t.Errorf("Error %v", err)
	}
	t.Logf("Total Number of recommendations: %d", len(recommendations))
	for _, recommendation := range recommendations {
		t.Logf("Recommendation: %v\n", recommendation)
	}
}
func TestCOInstance(t *testing.T) {
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

	/*recommendations, err := GetAWSCOResultForEC2Instance(config.GetConfig().Aws.Creds.Aws_access_key_id, config.GetConfig().Aws.Creds.Aws_secret_access_key, "ap-northeast-1", "867226238913", "i-068d71557bb611e95")
	if err != nil {
		t.Errorf("Error %v", err)
	}
	t.Logf("Total Number of recommendations: %d", len(recommendations))
	for _, recommendation := range recommendations {
		t.Logf("Instance: %s, Current: %s, New: %s", *recommendation.InstanceName, *recommendation.CurrentInstanceType, *recommendation.RecommendationOptions[0].InstanceType)
	}*/
	recommendation, err := GetAWSRecommendationForEC2Instance(config.GetConfig().Aws.Creds.Aws_access_key_id, config.GetConfig().Aws.Creds.Aws_secret_access_key, "ap-northeast-1", "867226238913", "i-068d71557bb611e95")
	if err != nil {
		t.Errorf("Error %v", err)
	}
	t.Logf("Recommendation: %v\n", recommendation)
}
