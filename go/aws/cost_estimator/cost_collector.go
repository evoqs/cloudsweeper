package cost_estimator

import (
	"encoding/json"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pricing"

	awsutil "cloudsweep/aws"
	aws_model "cloudsweep/aws/model"
	logger "cloudsweep/logging"
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
func CollectComputeInstanceCost(filters []*pricing.Filter) ([]aws_model.ResourceCostInstance, error) {
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
					cost, err := strconv.ParseFloat(priceDimension.PricePerUnit.USD, 32)
					if err != nil {
						prices["USD"] = -1
					}
					prices["USD"] = cost
					pricingDataList = append(pricingDataList, aws_model.ResourceCostInstance{
						Version:           priceItem.Version,
						PublicationDate:   priceItem.PublicationDate,
						ProductFamily:     priceItem.Product.ProductFamily,
						PricePerUnit:      prices,
						ProductAttributes: priceItem.Product.Attributes,
					})
					//fmt.Printf("USD Cost: %v  -- %s --  %s\n", prices, priceItem.Product.Attributes.RegionCode, priceDimension.Description)
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
