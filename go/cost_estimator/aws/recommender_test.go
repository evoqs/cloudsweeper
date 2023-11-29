package cost_estimator

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"testing"

	"cloudsweep/cloud_lib"
	"cloudsweep/config"
	aws_model "cloudsweep/model/aws"
	"cloudsweep/storage"
	"cloudsweep/utils"

	"github.com/aws/aws-sdk-go/service/ec2"
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
		InstanceType:    "t3.2xlarge",
		Memory:          "",
		RegionCode:      "us-east-1",
		OperatingSystem: "Linux",
	}
	resourceUsageParams := map[string]interface{}{
		"PeakCPUUtilization":    80,
		"AverageCPUUtilization": 4,
	}
	recommendation, err := GetInstanceTypeRecommendation(aws_model.ProductInfo[aws_model.ProductAttributesInstance]{
		Attributes: product}, resourceUsageParams)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Recommendation: %v\n", recommendation)
}

func TestRecommenderInstances(t *testing.T) {
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

	//------------------------------------------------------------------------
	file, err := os.Create("/tmp/recommendations.csv")
	if err != nil {
		t.Fatalf("Error creating CSV file: %v", err)
		return
	}
	defer file.Close()

	// Create a CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	header := []string{"InstanceType", "Region", "Memory", "OperatingSystem", "PeakCPUUtilization", "AverageCPUUtilization", "CurrentCost", "Recommended InstanceType", "NewCost", "EstimatedCostSavings"}
	if err := writer.Write(header); err != nil {
		t.Fatalf("Error writing CSV header: %v", err)
		return
	}
	//---------------------------------------------------------------------------

	//instanceTypes := []string{"t2", "t3", "m4", "m5", "m6", "m7"}
	instanceTypes := []string{"t2", "t3", "c5"}
	csAdminAwsClient, err := cloud_lib.GetCSAdminAwsClient()
	if err != nil {
		fmt.Println("Error creating CS Admin AWS Client:", err)
		return
	}
	allInstanceTypes, err := csAdminAwsClient.GetAllInstanceTypes(config.GetConfig().Aws.Creds.Aws_default_region, nil, nil)
	if err != nil {
		fmt.Println("Error getting instance types:", err)
		return
	}

	filteredInstanceTypes := filterInstanceTypes(allInstanceTypes, instanceTypes)

	// Get recommendations for each instance type
	for _, instanceType := range filteredInstanceTypes {
		pAttr := aws_model.ProductAttributesInstance{
			InstanceFamily:  "",
			InstanceType:    *instanceType.InstanceType,
			Memory:          "",
			RegionCode:      "us-east-1",
			OperatingSystem: "Linux",
		}
		cpuUtilizations := []int{10, 20, 40, 60, 80, 90, 98}

		for _, peakCPU := range cpuUtilizations {
			avgCPU := getAvgCPUForPeak(peakCPU)

			resourceUsageParams := map[string]interface{}{
				"PeakCPUUtilization":    peakCPU,
				"AverageCPUUtilization": avgCPU,
			}

			recommendation, err := GetInstanceTypeRecommendation(aws_model.ProductInfo[aws_model.ProductAttributesInstance]{
				Attributes: pAttr}, resourceUsageParams)
			if err != nil {
				fmt.Printf("Error getting recommendation for %s with PeakCPU %d: %v\n", *instanceType.InstanceType, peakCPU, err)
				continue
			}

			data := []string{
				*instanceType.InstanceType,
				pAttr.RegionCode,
				pAttr.Memory,
				pAttr.OperatingSystem,
				fmt.Sprintf("%v", peakCPU),
				fmt.Sprintf("%v", avgCPU),
				fmt.Sprintf("%.2f", recommendation.CurrentCost.MinPrice),
				recommendation.RecommendationItems[0].Resource.InstanceType,
				fmt.Sprintf("%.2f", recommendation.RecommendationItems[0].Cost.MinPrice),
				recommendation.RecommendationItems[0].EstimatedCostSavings,
			}

			fmt.Printf("Data: %v", data)
			if err := writer.Write(data); err != nil {
				fmt.Printf("Error writing CSV data for %s with PeakCPU %d: %v\n", *instanceType.InstanceType, peakCPU, err)
				continue
			}

			fmt.Printf("Recommendation for %s with PeakCPU %d: %v\n", *instanceType.InstanceType, peakCPU, recommendation.RecommendationItems[0].Resource)
		}
		writer.Flush()
	}
}

func filterInstanceTypes(allInstanceTypes []*ec2.InstanceTypeInfo, allowedTypes []string) []*ec2.InstanceTypeInfo {
	var filteredInstanceTypes []*ec2.InstanceTypeInfo

	for _, instanceType := range allInstanceTypes {
		for _, allowedType := range allowedTypes {
			if strings.HasPrefix(*instanceType.InstanceType, allowedType) {
				filteredInstanceTypes = append(filteredInstanceTypes, instanceType)
				break
			}
		}
	}
	return filteredInstanceTypes
}

func getAvgCPUForPeak(peakCPU int) int {
	output := peakCPU / 2
	if peakCPU < 50 {
		output = output - peakCPU/3
	} else {
		output = output + peakCPU/3
	}
	return output
}
