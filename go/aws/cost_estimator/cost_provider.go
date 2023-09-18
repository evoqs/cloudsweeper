package cost_estimator

import (
	"fmt"
	"reflect"
	"strings"

	aws_model "cloudsweep/aws/model"
	logger "cloudsweep/logging"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pricing"
)

type ResourceCost struct {
	Price    float64
	Unit     string
	Currency string
}

func GetCostForComputeInstance(productAttributes aws_model.ProductAttributesInstance) (ResourceCost, ResourceCost, error) {
	logger.NewDefaultLogger().Infof("Getting Cost for Compute Instance. Region: %s, InstanceType: %s", productAttributes.RegionCode, productAttributes.InstanceType)
	if productAttributes.RegionCode == "" || productAttributes.InstanceType == "" {
		return ResourceCost{Price: -1}, ResourceCost{Price: -1}, fmt.Errorf("Unable to get the Cost for ComputeInstance. RegionCode and/or InstanceType values are empty.")
	}
	var filters []*pricing.Filter
	reflectValue := reflect.ValueOf(productAttributes)
	for i := 0; i < reflectValue.NumField(); i++ {
		field := reflectValue.Field(i)
		fieldName := reflectValue.Type().Field(i).Name
		fieldValue := field.Interface().(string)

		if fieldValue != "" {
			filters = append(filters, &pricing.Filter{
				Field: aws.String(fieldName),
				Type:  aws.String("TERM_MATCH"),
				Value: aws.String(fieldValue),
			})
		}
	}
	//fmt.Printf("Filters %v", filters)
	resourceCosts, err := CollectComputeInstanceCost(filters)
	if err != nil {
		return ResourceCost{Price: -1}, ResourceCost{Price: -1}, fmt.Errorf("Unable to get the Cost for ComputeInstance.Error:  %v", err)
	}

	//fmt.Println("Total Number of Resource Costs: ", len(resourceCosts))

	var minResourceCost aws_model.ResourceCostInstance
	var maxResourceCost aws_model.ResourceCostInstance
	for _, resource := range resourceCosts {
		if resource.PricePerUnit["USD"] > 0 && !strings.Contains(strings.ToLower(resource.ProductAttributes.UsageType), "reservation") {
			if minResourceCost.PricePerUnit == nil && maxResourceCost.PricePerUnit == nil {
				minResourceCost = resource
				maxResourceCost = resource
			}
			if resource.PricePerUnit["USD"] < minResourceCost.PricePerUnit["USD"] {
				minResourceCost = resource
			}
			if resource.PricePerUnit["USD"] > maxResourceCost.PricePerUnit["USD"] {
				maxResourceCost = resource
			}
		}
	}
	if minResourceCost.PricePerUnit == nil {
		return ResourceCost{}, ResourceCost{}, fmt.Errorf("No Cost Details found for the given filters.")
	}
	//fmt.Printf("Min: %f Max: %f\n", minResourceCost.PricePerUnit["USD"], maxResourceCost.PricePerUnit["USD"])
	return ResourceCost{
			Price:    minResourceCost.PricePerUnit["USD"],
			Unit:     minResourceCost.Unit,
			Currency: "USD",
		}, ResourceCost{
			Price:    maxResourceCost.PricePerUnit["USD"],
			Unit:     maxResourceCost.Unit,
			Currency: "USD",
		}, nil
}
