package cost_estimator

import (
	"encoding/json"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pricing"

	awsutil "cloudsweep/cloud_lib"
	logger "cloudsweep/logging"
	aws_model "cloudsweep/model/aws"
)

/*type Filter struct {
	Field string
	Type  string
	Value string
}

func buildFilterInput(filters []Filter) []*pricing.Filter {
	filtersInput := []*pricing.Filter{}
	for _, filter := range filters {
		filter := &pricing.Filter{
			Field: aws.String(filter.Field),
			Type:  aws.String(filter.Type),
			Value: aws.String(filter.Value),
		}
		filtersInput = append(filtersInput, filter)
	}
	return filtersInput
}*/

// Note: This function calculates the cost for currency USD. filter can not be used to filter out different currencies.
// TODO: Update model and support multiple currencies - future. Enhancement support is provided
func CollectResourceCost[T any](serviceCode string, filters []*pricing.Filter, resultContainer *[]aws_model.AwsResourceCost[T]) error {
	sess, err := awsutil.GetAwsSession()
	if err != nil {
		logger.NewDefaultLogger().Errorf("Error while creating AWS Session %v. Skipping Instance Cost Collection.", err)
		return err
	}
	pricingClient := pricing.New(sess)
	pricingInput := &pricing.GetProductsInput{
		ServiceCode: aws.String(serviceCode),
		Filters:     filters,
		NextToken:   aws.String(""),
	}

	for {
		var pricingData aws_model.PricingData[T]
		pricingResult, err := pricingClient.GetProducts(pricingInput)
		if err != nil {
			return err
		}

		pricingJSON, err := json.Marshal(pricingResult)
		if err != nil {
			return err
		}

		err = json.Unmarshal(pricingJSON, &pricingData)
		if err != nil {
			return err
		}

		for _, priceItem := range pricingData.PriceList {
			for _, term := range priceItem.Terms.OnDemand {
				for _, priceDimension := range term.PriceDimensions {
					prices := make(map[string]float64)
					cost, err := strconv.ParseFloat(priceDimension.PricePerUnit.USD, 64)
					if err != nil {
						prices["USD"] = -1
					}
					prices["USD"] = cost

					// Create a new resource cost instance
					instance := aws_model.AwsResourceCost[T]{
						CloudProvider:     "AWS",
						Version:           priceItem.Version,
						PublicationDate:   priceItem.PublicationDate,
						ProductFamily:     priceItem.Product.ProductFamily,
						PricePerUnit:      prices,
						Unit:              priceDimension.Unit,
						ProductAttributes: priceItem.Product.Attributes,
					}

					// Append the instance to the result container slice
					*resultContainer = append(*resultContainer, instance)
				}
			}
		}
		if pricingResult.NextToken != nil {
			pricingInput.NextToken = pricingResult.NextToken
		} else {
			break
		}
	}
	return nil
}

/*func CollectResourceCost(serviceCode string, filters []*pricing.Filter, resultContainer interface{}) error {
	sess, err := awsutil.GetAwsSession()
	if err != nil {
		logger.NewDefaultLogger().Errorf("Error while creating AWS Session %v. Skipping Instance Cost Collection.", err)
		return err
	}
	pricingClient := pricing.New(sess)
	pricingInput := &pricing.GetProductsInput{
		ServiceCode: aws.String(serviceCode),
		Filters:     filters,
		NextToken:   aws.String(""),
	}

	// Get the element type of the resultContainer slice
	resultElemType := reflect.TypeOf(resultContainer).Elem().Elem()

	for {
		var pricingData aws_model.PricingDataInstance
		pricingResult, err := pricingClient.GetProducts(pricingInput)
		if err != nil {
			return err
		}

		// TODO: Check for JSON encode / decode
		pricingJSON, err := json.Marshal(pricingResult)
		if err != nil {
			return err
		}

		err = json.Unmarshal(pricingJSON, &pricingData)
		if err != nil {
			return err
		}

		for _, priceItem := range pricingData.PriceList {
			for _, term := range priceItem.Terms.OnDemand {
				for _, priceDimension := range term.PriceDimensions {
					prices := make(map[string]float64)
					cost, err := strconv.ParseFloat(priceDimension.PricePerUnit.USD, 64)
					if err != nil {
						prices["USD"] = -1
					}
					prices["USD"] = cost

					// Create an instance of the desired type using reflection
					instanceValue := reflect.New(resultElemType).Elem()

					// Populate the instance
					instanceValue.FieldByName("CloudProvider").SetString("AWS")
					instanceValue.FieldByName("Version").SetString(priceItem.Version)
					instanceValue.FieldByName("PublicationDate").SetString(priceItem.PublicationDate)
					instanceValue.FieldByName("ProductFamily").SetString(priceItem.Product.ProductFamily)
					instanceValue.FieldByName("PricePerUnit").Set(reflect.ValueOf(prices))
					instanceValue.FieldByName("Unit").SetString(priceDimension.Unit)
					instanceValue.FieldByName("ProductAttributes").Set(reflect.ValueOf(priceItem.Product.Attributes))

					// Append the instance to the result container slice
					resultSlice := reflect.ValueOf(resultContainer).Elem()
					resultSlice.Set(reflect.Append(resultSlice, instanceValue))
				}
			}
		}
		if pricingResult.NextToken != nil {
			pricingInput.NextToken = pricingResult.NextToken
		} else {
			break
		}
	}
	return nil
}*/

// Note: This function calculates the cost for currency USD. filter can not be used to filter out different currencies.
// TODO: Update model and support multiple currencies - future. Enhancement support is provided
/*func CollectComputeInstanceCost(filters []*pricing.Filter) ([]aws_model.ResourceCostInstance, error) {
	var pricingDataList []aws_model.ResourceCostInstance
	sess, err := awsutil.GetAwsSession()
	if err != nil {
		logger.NewDefaultLogger().Errorf("Error while creating AWS Session %v. Skipping Instance Cost Collection.", err)
		return pricingDataList, err
	}
	pricingClient := pricing.New(sess)
	pricingInput := &pricing.GetProductsInput{
		ServiceCode: aws.String("AmazonEC2"),
		Filters:     filters,
		NextToken:   aws.String(""),
	}

	for {
		var pricingData aws_model.PricingDataInstance
		pricingResult, err := pricingClient.GetProducts(pricingInput)
		if err != nil {
			panic(err)
		}

		pricingJSON, err := json.Marshal(pricingResult)
		if err != nil {
			panic(err)
		}

		err = json.Unmarshal(pricingJSON, &pricingData)
		if err != nil {
			panic(err)
		}
		//fmt.Println("Size of Price List = ", len(pricingData.PriceList))
		for _, priceItem := range pricingData.PriceList {
			// Loop through OnDemand terms and PriceDimensions for each entry
			for _, term := range priceItem.Terms.OnDemand {
				for _, priceDimension := range term.PriceDimensions {
					prices := make(map[string]float64)

					// Filters - Price should not be 0
					//var cost float64
					cost, err := strconv.ParseFloat(priceDimension.PricePerUnit.USD, 64)
					if err != nil {
						prices["USD"] = -1
					}
					prices["USD"] = cost
					pricingDataList = append(pricingDataList, aws_model.ResourceCostInstance{
						CloudProvider:     "AWS",
						Version:           priceItem.Version,
						PublicationDate:   priceItem.PublicationDate,
						ProductFamily:     priceItem.Product.ProductFamily,
						PricePerUnit:      prices,
						ProductAttributes: priceItem.Product.Attributes,
						Unit:              priceDimension.Unit,
					})
					//fmt.Printf("USD Cost: %v  -- %s --  %s\n", prices, priceItem.Product.Attributes.InstanceType, priceDimension.Description)
				}
			}
		}
		if pricingResult.NextToken != nil {
			pricingInput.NextToken = pricingResult.NextToken
		} else {
			// No more pages, break the loop
			break
		}
	}
	return pricingDataList, nil
}

// Note used, just for reference
func CollectResourceCostNoReflect(serviceCode string, filters []*pricing.Filter, resultContainer interface{}, factory func() interface{}) error {
	sess, err := awsutil.GetAwsSession()
	if err != nil {
		logger.NewDefaultLogger().Errorf("Error while creating AWS Session %v. Skipping Instance Cost Collection.", err)
		return err
	}
	pricingClient := pricing.New(sess)
	pricingInput := &pricing.GetProductsInput{
		ServiceCode: aws.String(serviceCode),
		Filters:     filters,
		NextToken:   aws.String(""),
	}

	for {
		var pricingData aws_model.PricingDataInstance
		pricingResult, err := pricingClient.GetProducts(pricingInput)
		if err != nil {
			return err
		}

		pricingJSON, err := json.Marshal(pricingResult)
		if err != nil {
			return err
		}

		err = json.Unmarshal(pricingJSON, &pricingData)
		if err != nil {
			return err
		}

		for _, priceItem := range pricingData.PriceList {
			for _, term := range priceItem.Terms.OnDemand {
				for _, priceDimension := range term.PriceDimensions {
					prices := make(map[string]float64)
					cost, err := strconv.ParseFloat(priceDimension.PricePerUnit.USD, 64)
					if err != nil {
						prices["USD"] = -1
					}
					prices["USD"] = cost

					// Create an instance of the desired type using the factory function
					instance := factory()

					// Populate the instance with the received data
					if rci, ok := instance.(*aws_model.ResourceCostInstance); ok {
						rci.CloudProvider = "AWS"
						rci.Version = priceItem.Version
						rci.PublicationDate = priceItem.PublicationDate
						rci.ProductFamily = priceItem.Product.ProductFamily
						rci.PricePerUnit = prices
						rci.ProductAttributes = priceItem.Product.Attributes
						rci.Unit = priceDimension.Unit
					} else if rce, ok := instance.(*aws_model.ResourceCostEBS); ok {
						rce.CloudProvider = "AWS"
						rce.Version = priceItem.Version
						rce.PublicationDate = priceItem.PublicationDate
						rce.ProductFamily = priceItem.Product.ProductFamily
						rce.PricePerUnit = prices
						rce.ProductAttributes = priceItem.Product.Attributes
						rce.Unit = priceDimension.Unit
					}

					// Append the populated instance to the result container slice
					reflect.ValueOf(resultContainer).Elem().Set(reflect.Append(reflect.ValueOf(resultContainer).Elem(), reflect.ValueOf(instance).Elem()))
				}
			}
		}
		if pricingResult.NextToken != nil {
			pricingInput.NextToken = pricingResult.NextToken
		} else {
			break
		}
	}
	return nil
}*/
