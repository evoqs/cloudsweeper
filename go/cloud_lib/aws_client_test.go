package cloud_lib

import (
	"fmt"
	"testing"

	"cloudsweep/config"
	"cloudsweep/storage"
	"cloudsweep/utils"
)

func TestAdminAwsClient(t *testing.T) {
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
		t.Logf("Failed to connect to DB %s\n", err)
	}
	err = dbM.CheckConnection()
	if err != nil {
		t.Log("Connection Check failed\n")
	} else {
		t.Log("Successfully Connected\n")
		defer dbM.Disconnect()
	}
	storage.MakeDBOperators(dbM)

	//awsClient, err := NewAwsClient(config.GetConfig().Aws.Creds.Aws_access_key_id, config.GetConfig().Aws.Creds.Aws_secret_access_key, config.GetConfig().Aws.Creds.Aws_default_region)
	awsClient, err := GetCSAdminAwsClient()
	_, err = GetCSAdminAwsClient()
	if !awsClient.ValidateCredentials() {
		t.Errorf("Invalid Credentials..")
	}
	if err != nil {
		t.Errorf("ERROR: Problem Creating Aws Client: %s\n", err)
		return
	}
	regionCodes, err := awsClient.GetSubscribedRegionCodes()
	if err != nil {
		t.Logf("Error getting subscribed region codes: %s\n", err)
		return
	}
	t.Logf("Subscribed Region Codes: ")
	for _, regionCode := range regionCodes {
		t.Logf("%s ", regionCode)
	}

	instanceTypes, err := awsClient.GetAllInstanceTypes(config.GetConfig().Aws.Creds.Aws_default_region, nil, nil)
	if err != nil {
		t.Logf("Error getting subscribed instance types: %s\n", err)
		return
	}
	t.Logf("Instance Types: ")
	for _, instanceType := range instanceTypes {
		t.Logf("%s ", *instanceType.InstanceType)
	}
}

func TestAwsClient(t *testing.T) {
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
		t.Logf("Failed to connect to DB %s\n", err)
	}
	err = dbM.CheckConnection()
	if err != nil {
		t.Log("Connection Check failed\n")
	} else {
		t.Log("Successfully Connected\n")
		defer dbM.Disconnect()
	}
	storage.MakeDBOperators(dbM)

	//awsClient, err := NewAwsClient(config.GetConfig().Aws.Creds.Aws_access_key_id, config.GetConfig().Aws.Creds.Aws_secret_access_key, config.GetConfig().Aws.Creds.Aws_default_region)
	awsClient, err := GetAwsClient(config.GetConfig().Aws.Creds.Aws_access_key_id, config.GetConfig().Aws.Creds.Aws_secret_access_key, "")
	if !awsClient.ValidateCredentials() {
		t.Errorf("Invalid Credentials..")
	}
	if err != nil {
		t.Errorf("ERROR: Problem Creating Aws Client: %s\n", err)
		return
	}
	regionCodes, err := awsClient.GetSubscribedRegionCodes()
	if err != nil {
		t.Logf("Error getting subscribed region codes: %s\n", err)
		return
	}
	t.Logf("Subscribed Region Codes: ")
	for _, regionCode := range regionCodes {
		t.Logf("%s ", regionCode)
	}

	instanceTypes, err := awsClient.GetAllInstanceTypes(config.GetConfig().Aws.Creds.Aws_default_region, nil, nil)
	if err != nil {
		t.Logf("Error getting subscribed instance types: %s\n", err)
		return
	}
	t.Logf("Instance Types: ")
	for _, instanceType := range instanceTypes {
		t.Logf("%s ", *instanceType.InstanceType)
	}
}

func TestAwsGetResourceDetails(t *testing.T) {
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
		t.Logf("Failed to connect to DB %s\n", err)
	}
	err = dbM.CheckConnection()
	if err != nil {
		t.Log("Connection Check failed\n")
	} else {
		t.Log("Successfully Connected\n")
		defer dbM.Disconnect()
	}
	storage.MakeDBOperators(dbM)

	//awsClient, err := NewAwsClient(config.GetConfig().Aws.Creds.Aws_access_key_id, config.GetConfig().Aws.Creds.Aws_secret_access_key, config.GetConfig().Aws.Creds.Aws_default_region)
	awsClient, err := GetAwsClient(config.GetConfig().Aws.Creds.Aws_access_key_id, config.GetConfig().Aws.Creds.Aws_secret_access_key, "")
	instance, err := awsClient.GetInstanceDetails("i-030581259aa42a982", "ap-southeast-2")
	if err != nil {
		t.Logf("Error getting instance details: %s\n", err)
		return
	}
	t.Logf("Instance: %+v ", instance)

	volume, err := awsClient.GetEbsVolumeDetails("vol-0ba99949c5f8009cb", "ap-northeast-1")
	if err != nil {
		t.Logf("Error getting instance details: %s\n", err)
		return
	}
	t.Logf("EBS Volume: %+v ", volume)
}
