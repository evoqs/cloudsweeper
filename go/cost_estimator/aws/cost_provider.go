package cost_estimator

import (
	"fmt"
	"reflect"
	"strings"

	logger "cloudsweep/logging"
	"cloudsweep/model"
	aws_model "cloudsweep/model/aws"
	"cloudsweep/storage"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pricing"
)

func GetComputeInstanceCost(pInfo aws_model.ProductInfo[aws_model.ProductAttributesInstance]) (model.ResourceCost, error) {
	if pInfo.Attributes.RegionCode == "" || pInfo.Attributes.InstanceType == "" || pInfo.Attributes.OperatingSystem == "" {
		return model.ResourceCost{MinPrice: -1}, fmt.Errorf("Unable to get the Cost for ComputeInstance. RegionCode and/or InstanceType values are empty.")
	}
	return getCost("AmazonEC2", pInfo)
}

// =============================================== EBS ================================================================

func GetEbsCost(pInfo aws_model.ProductInfo[aws_model.ProductAttributesEBS]) (model.ResourceCost, error) {
	if pInfo.Attributes.RegionCode == "" || (pInfo.Attributes.StorageMedia == "" && pInfo.Attributes.VolumeApiName == "") {
		return model.ResourceCost{MinPrice: -1}, fmt.Errorf("Unable to get the Cost for EBS. RegionCode and/or StorageMedia/VolumeApiName values are empty.")
	}
	return getCost("AmazonEC2", pInfo)
}

// ==================================== Generic Functions ==================================================

func getCost[T any](serviceCode string, pInfo aws_model.ProductInfo[T]) (model.ResourceCost, error) {
	min, max, err := GetCostFromDB(pInfo)
	if err == nil {
		return model.ResourceCost{
			MinPrice: min.PricePerUnit["USD"],
			MaxPrice: max.PricePerUnit["USD"],
			Unit:     min.Unit,
			Currency: "USD",
		}, nil
	}

	min, max, err = GetCostFromAws(serviceCode, pInfo)
	if err != nil {
		return model.ResourceCost{MinPrice: -1}, fmt.Errorf("Unable to get the cost for the Instance. %v", err)
	}

	opr := storage.GetDefaultDBOperators()
	opr.CostOperator.AddResourceCost(min)
	opr.CostOperator.AddResourceCost(max)

	return model.ResourceCost{
		MinPrice: min.PricePerUnit["USD"],
		MaxPrice: max.PricePerUnit["USD"],
		Unit:     min.Unit,
		Currency: "USD",
	}, nil
}

func buildFilterInput[T interface{}](productInfo aws_model.ProductInfo[T]) []*pricing.Filter {
	var filters []*pricing.Filter
	reflectValue := reflect.ValueOf(productInfo.Attributes)
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
	if productInfo.ProductFamily != "" {
		filters = append(filters, &pricing.Filter{
			Field: aws.String("productFamily"),
			Type:  aws.String("TERM_MATCH"),
			Value: aws.String(productInfo.ProductFamily),
		})
	}

	return filters
}

// Get the Cost From AWS for any resource
func GetCostFromAws[T any](serviceCode string, pInfo aws_model.ProductInfo[T]) (aws_model.AwsResourceCost[T],
	aws_model.AwsResourceCost[T], error) {
	logger.NewDefaultLogger().Infof("Getting Cost for Compute EBS: %v", pInfo)
	/*if err := validateMinimumFilterFieldsEbs(productAttributes); err != nil {
		return aws_model.AwsResourceCost[T]{}, aws_model.AwsResourceCost[T]{}, err
	}*/
	filters := buildFilterInput(pInfo)

	var resourceCosts []aws_model.AwsResourceCost[T]
	err := CollectResourceCost(serviceCode, filters, &resourceCosts)
	if err != nil {
		return aws_model.AwsResourceCost[T]{},
			aws_model.AwsResourceCost[T]{},
			fmt.Errorf("Unable to get the Cost for Compute EBS. Error:  %v", err)
	}

	//fmt.Println("Total Number of Resource Costs: ", len(resourceCosts))
	var minResourceCost aws_model.AwsResourceCost[T]
	var maxResourceCost aws_model.AwsResourceCost[T]
	for _, resource := range resourceCosts {
		if resource.PricePerUnit["USD"] > 0 && !isUsageTypeReservation[T](resource.ProductAttributes) {
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
		return aws_model.AwsResourceCost[T]{},
			aws_model.AwsResourceCost[T]{},
			fmt.Errorf("No Cost Details found for the given filters.")
	}
	//fmt.Printf("Min: %f Max: %f\n", minResourceCost.PricePerUnit["USD"], maxResourceCost.PricePerUnit["USD"])
	return minResourceCost, maxResourceCost, nil
}

func isUsageTypeReservation[T any](productAttributes T) bool {
	attrValue := reflect.ValueOf(productAttributes)
	if attrValue.Kind() == reflect.Struct {
		usageTypeField := attrValue.FieldByName("UsageType")
		if usageTypeField.IsValid() && usageTypeField.Kind() == reflect.String {
			usageType := usageTypeField.String()
			if strings.Contains(strings.ToLower(usageType), "reservation") {
				return true
			}
		}
	}
	return false
}

func GetCostFromDB[T any](pInfo aws_model.ProductInfo[T]) (aws_model.AwsResourceCost[T],
	aws_model.AwsResourceCost[T], error) {
	logger.NewDefaultLogger().Debugf("Query DB for Cost for EBS: %v", pInfo)
	/*if err := validateMinimumFilterFieldsEbs(pAttr); err != nil {
		return aws_model.AwsResourceCost[T]{}, aws_model.AwsResourceCost[T]{}, err
	}*/
	var queryParts []string
	queryParts = append(queryParts, "\"cloudProvider\": \"AWS\"")
	reflectValue := reflect.ValueOf(pInfo.Attributes)
	for i := 0; i < reflectValue.NumField(); i++ {
		field := reflectValue.Field(i)
		//fieldName := reflectValue.Type().Field(i).Name
		fieldValue := field.Interface().(string)
		jsonTagName := reflectValue.Type().Field(i).Tag.Get("json")
		if fieldValue != "" {
			queryParts = append(queryParts, "\"productAttributes."+jsonTagName+"\": \""+fieldValue+"\"")
		}
	}
	// TODO: Need to enable it
	queryParts = append(queryParts, "\"productFamily\": \""+pInfo.ProductFamily+"\"")

	query := "{" + strings.Join(queryParts, ", ") + "}"
	logger.NewDefaultLogger().Debugf("DB Query: %s", query)

	//results,err := storage.GetDefaultCostOperator().RunQuery(query)
	var resourceCosts []aws_model.AwsResourceCost[T]
	opr := storage.GetDefaultDBOperators()
	opr.CostOperator.GetQueryResult(query, &resourceCosts)
	logger.NewDefaultLogger().Debugf("Length of resourceCostInstances from DB: %d\n", len(resourceCosts))

	if len(resourceCosts) >= 2 {
		if resourceCosts[0].PricePerUnit["USD"] < resourceCosts[1].PricePerUnit["USD"] {
			return resourceCosts[0], resourceCosts[1], nil
		}
		return resourceCosts[1], resourceCosts[0], nil
	}
	return aws_model.AwsResourceCost[T]{},
		aws_model.AwsResourceCost[T]{},
		fmt.Errorf("No cost details present in the DB for the given input")
}

// ============================================================================================
/*func GetComputeInstanceCostFromAws(productAttributes aws_model.ProductAttributesInstance) (aws_model.AwsResourceCost[aws_model.ProductAttributesInstance],
	aws_model.AwsResourceCost[aws_model.ProductAttributesInstance], error) {
	logger.NewDefaultLogger().Infof("Getting Cost for Compute Instance: %v", productAttributes)
	if productAttributes.RegionCode == "" || productAttributes.InstanceType == "" {
		return aws_model.AwsResourceCost[aws_model.ProductAttributesInstance]{},
			aws_model.AwsResourceCost[aws_model.ProductAttributesInstance]{},
			fmt.Errorf("Unable to get the Cost for ComputeInstance. RegionCode and/or InstanceType values are empty.")
	}
	filters := buildFilterInput(productAttributes)

	//fmt.Printf("Filters %v", filters)
	//resourceCosts, err := CollectComputeInstanceCost(filters)
	var resourceCosts []aws_model.AwsResourceCost[aws_model.ProductAttributesInstance]
	err := CollectResourceCost("AmazonEC2", filters, &resourceCosts)
	if err != nil {
		return aws_model.AwsResourceCost[aws_model.ProductAttributesInstance]{},
			aws_model.AwsResourceCost[aws_model.ProductAttributesInstance]{},
			fmt.Errorf("Unable to get the Cost for ComputeInstance.Error:  %v", err)
	}

	//fmt.Println("Total Number of Resource Costs: ", len(resourceCosts))
	var minResourceCost aws_model.AwsResourceCost[aws_model.ProductAttributesInstance]
	var maxResourceCost aws_model.AwsResourceCost[aws_model.ProductAttributesInstance]
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
		return aws_model.AwsResourceCost[aws_model.ProductAttributesInstance]{},
			aws_model.AwsResourceCost[aws_model.ProductAttributesInstance]{},
			fmt.Errorf("No Cost Details found for the given filters.")
	}
	//fmt.Printf("Min: %f Max: %f\n", minResourceCost.PricePerUnit["USD"], maxResourceCost.PricePerUnit["USD"])
	return minResourceCost, maxResourceCost, nil
}

func GetEbsCostFromAws(productAttributes aws_model.ProductAttributesEBS) (aws_model.AwsResourceCost[aws_model.ProductAttributesEBS],
	aws_model.AwsResourceCost[aws_model.ProductAttributesEBS], error) {
	logger.NewDefaultLogger().Infof("Getting Cost for Compute EBS: %v", productAttributes)
	if err := validateMinimumFilterFieldsEbs(productAttributes); err != nil {
		return aws_model.AwsResourceCost[aws_model.ProductAttributesEBS]{}, aws_model.AwsResourceCost[aws_model.ProductAttributesEBS]{}, err
	}
	filters := buildFilterInput(productAttributes)

	var resourceCosts []aws_model.AwsResourceCost[aws_model.ProductAttributesEBS]
	err := CollectResourceCost("AmazonEC2", filters, &resourceCosts)
	if err != nil {
		return aws_model.AwsResourceCost[aws_model.ProductAttributesEBS]{},
			aws_model.AwsResourceCost[aws_model.ProductAttributesEBS]{},
			fmt.Errorf("Unable to get the Cost for Compute EBS. Error:  %v", err)
	}

	//fmt.Println("Total Number of Resource Costs: ", len(resourceCosts))
	var minResourceCost aws_model.AwsResourceCost[aws_model.ProductAttributesEBS]
	var maxResourceCost aws_model.AwsResourceCost[aws_model.ProductAttributesEBS]
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
		return aws_model.AwsResourceCost[aws_model.ProductAttributesEBS]{},
			aws_model.AwsResourceCost[aws_model.ProductAttributesEBS]{},
			fmt.Errorf("No Cost Details found for the given filters.")
	}
	//fmt.Printf("Min: %f Max: %f\n", minResourceCost.PricePerUnit["USD"], maxResourceCost.PricePerUnit["USD"])
	return minResourceCost, maxResourceCost, nil
}
func GetComputeInstanceCostFromDB(productAttributes aws_model.ProductAttributesInstance) (aws_model.AwsResourceCost[aws_model.ProductAttributesInstance],
	aws_model.AwsResourceCost[aws_model.ProductAttributesInstance], error) {
	logger.NewDefaultLogger().Debugf("Query DB for Cost for Instance: %v", productAttributes)
	if productAttributes.RegionCode == "" || productAttributes.InstanceType == "" {
		return aws_model.AwsResourceCost[aws_model.ProductAttributesInstance]{},
			aws_model.AwsResourceCost[aws_model.ProductAttributesInstance]{},
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
	var resourceCostInstances []aws_model.AwsResourceCost[aws_model.ProductAttributesInstance]
	opr := storage.GetDefaultDBOperators()
	opr.CostOperator.GetQueryResult(query, &resourceCostInstances)
	fmt.Printf("1st Length of resourceCostInstances from DB: %d\n", len(resourceCostInstances))

	if len(resourceCostInstances) >= 2 {
		if resourceCostInstances[0].PricePerUnit["USD"] < resourceCostInstances[1].PricePerUnit["USD"] {
			return resourceCostInstances[0], resourceCostInstances[1], nil
		}
		return resourceCostInstances[1], resourceCostInstances[0], nil
	}
	return aws_model.AwsResourceCost[aws_model.ProductAttributesInstance]{},
		aws_model.AwsResourceCost[aws_model.ProductAttributesInstance]{},
		fmt.Errorf("No cost details present in the DB for the given input")
}

func GetEbsCostFromDB(productAttributes aws_model.ProductAttributesEBS) (aws_model.AwsResourceCost[aws_model.ProductAttributesEBS],
	aws_model.AwsResourceCost[aws_model.ProductAttributesEBS], error) {
	logger.NewDefaultLogger().Debugf("Query DB for Cost for EBS: %v", productAttributes)
	if err := validateMinimumFilterFieldsEbs(productAttributes); err != nil {
		return aws_model.AwsResourceCost[aws_model.ProductAttributesEBS]{}, aws_model.AwsResourceCost[aws_model.ProductAttributesEBS]{}, err
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
	var resourceCosts []aws_model.AwsResourceCost[aws_model.ProductAttributesEBS]
	opr := storage.GetDefaultDBOperators()
	opr.CostOperator.GetQueryResult(query, &resourceCosts)
	fmt.Printf("1st Length of resourceCostInstances from DB: %d\n", len(resourceCosts))

	if len(resourceCosts) >= 2 {
		if resourceCosts[0].PricePerUnit["USD"] < resourceCosts[1].PricePerUnit["USD"] {
			return resourceCosts[0], resourceCosts[1], nil
		}
		return resourceCosts[1], resourceCosts[0], nil
	}
	return aws_model.AwsResourceCost[aws_model.ProductAttributesEBS]{},
		aws_model.AwsResourceCost[aws_model.ProductAttributesEBS]{},
		fmt.Errorf("No cost details present in the DB for the given input")
}
*/
