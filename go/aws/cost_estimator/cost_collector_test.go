package cost_estimator

import (
	"fmt"
	"testing"

	aws_model "cloudsweep/aws/model"
)

func TestCollector(t *testing.T) {
	product := aws_model.ProductAttributesInstance{
		InstanceFamily:  "",
		InstanceType:    "t2.large",
		Memory:          "",
		RegionCode:      "us-east-1",
		OperatingSystem: "Linux",
	}

	min, max, err := GetComputeInstanceCost(product)
	if err != nil {
		fmt.Printf("Error %v", err)
	}
	fmt.Printf("Min: %v Max: %v\n", min, max)
}
