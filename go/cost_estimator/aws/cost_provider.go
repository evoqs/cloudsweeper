package cost_estimator

import (
	"fmt"
	"reflect"
	"strings"

	logger "cloudsweep/logging"
	aws_model "cloudsweep/model/aws"
	"cloudsweep/storage"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pricing"
)

type ResourceCost struct {
	MinPrice float64
	MaxPrice float64
	Unit     string
	Currency string
}

func GetComputeInstanceCost(productAttributes aws_model.ProductAttributesInstance) (ResourceCost, error) {
	min, max, err := GetComputeInstanceCostFromDB(productAttributes)
	if err == nil {
		return ResourceCost{
			MinPrice: min.PricePerUnit["USD"],
			MaxPrice: max.PricePerUnit["USD"],
			Unit:     min.Unit,
			Currency: "USD",
		}, nil
	}

	min, max, err = GetComputeInstanceCostFromAws(productAttributes)
	if err != nil {
		return ResourceCost{MinPrice: -1}, fmt.Errorf("Unable to get the cost for the Instance. %v", err)
	}

	opr := storage.GetDefaultDBOperators()
	opr.CostOperator.AddResourceCost(min)
	opr.CostOperator.AddResourceCost(max)

	return ResourceCost{
		MinPrice: min.PricePerUnit["USD"],
		MaxPrice: max.PricePerUnit["USD"],
		Unit:     min.Unit,
		Currency: "USD",
	}, nil
}

func GetComputeInstanceCostFromDB(productAttributes aws_model.ProductAttributesInstance) (aws_model.ResourceCostInstance, aws_model.ResourceCostInstance, error) {
	logger.NewDefaultLogger().Debugf("Query DB for Cost for Instance: " + productAttributes.InstanceType)
	if productAttributes.RegionCode == "" || productAttributes.InstanceType == "" {
		return aws_model.ResourceCostInstance{},
			aws_model.ResourceCostInstance{},
			fmt.Errorf("Unable to get the Cost for ComputeInstance. RegionCode and/or InstanceType values are empty.")
	}
	var queryParts []string
	queryParts = append(queryParts, "\"cloudProvider\": \"AWS\"")
	reflectValue := reflect.ValueOf(productAttributes)
	for i := 0; i < reflectValue.NumField(); i++ {
		field := reflectValue.Field(i)
		//fieldName := reflectValue.Type().Field(i).Name
		fieldValue := field.Interface().(string)
		jsonTagName := reflectValue.Type().Field(i).Tag.Get("json")
		if fieldValue != "" {
			queryParts = append(queryParts, "\"productAttributes."+jsonTagName+"\": \""+fieldValue+"\"")
		}
	}
	query := "{" + strings.Join(queryParts, ", ") + "}"
	fmt.Printf("DB Query: %s", query)

	//results,err := storage.GetDefaultCostOperator().RunQuery(query)
	var resourceCostInstances []aws_model.ResourceCostInstance
	opr := storage.GetDefaultDBOperators()
	opr.CostOperator.GetQueryResult(query, &resourceCostInstances)
	fmt.Printf("1st Length of resourceCostInstances from DB: %d\n", len(resourceCostInstances))

	if len(resourceCostInstances) >= 2 {
		if resourceCostInstances[0].PricePerUnit["USD"] < resourceCostInstances[1].PricePerUnit["USD"] {
			return resourceCostInstances[0], resourceCostInstances[1], nil
		}
		return resourceCostInstances[1], resourceCostInstances[0], nil
	}
	return aws_model.ResourceCostInstance{}, aws_model.ResourceCostInstance{}, fmt.Errorf("No cost details present in the DB for the given input")
}

func GetComputeInstanceCostFromAws(productAttributes aws_model.ProductAttributesInstance) (aws_model.ResourceCostInstance, aws_model.ResourceCostInstance, error) {
	logger.NewDefaultLogger().Infof("Getting Cost for Compute Instance. Region: %s, InstanceType: %s", productAttributes.RegionCode, productAttributes.InstanceType)
	if productAttributes.RegionCode == "" || productAttributes.InstanceType == "" {
		return aws_model.ResourceCostInstance{},
			aws_model.ResourceCostInstance{},
			fmt.Errorf("Unable to get the Cost for ComputeInstance. RegionCode and/or InstanceType values are empty.")
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
		return aws_model.ResourceCostInstance{},
			aws_model.ResourceCostInstance{},
			fmt.Errorf("Unable to get the Cost for ComputeInstance.Error:  %v", err)
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
		return aws_model.ResourceCostInstance{},
			aws_model.ResourceCostInstance{},
			fmt.Errorf("No Cost Details found for the given filters.")
	}
	//fmt.Printf("Min: %f Max: %f\n", minResourceCost.PricePerUnit["USD"], maxResourceCost.PricePerUnit["USD"])
	return minResourceCost, maxResourceCost, nil
}
