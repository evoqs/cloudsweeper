package cost_estimator

import (
	logger "cloudsweep/logging"
	aws_model "cloudsweep/model/aws"
	"cloudsweep/utils"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
)

func GetInstanceTypeRecommendation(pInfo aws_model.ProductInfo[aws_model.ProductAttributesInstance], resourceUsageParams map[string]interface{}) (aws_model.Recommendation[aws_model.InstanceDetails], error) {
	currentCost, err := GetComputeInstanceCost(pInfo)
	if err != nil {
		return aws_model.Recommendation[aws_model.InstanceDetails]{}, err
	}
	recommendation := aws_model.Recommendation[aws_model.InstanceDetails]{
		CurrentResourceDetails: aws_model.InstanceDetails{
			InstanceType: pInfo.Attributes.InstanceType,
		},
		CurrentCost: currentCost,
	}
	newResAttr, err := GetRecommendedResource(pInfo.Attributes, resourceUsageParams)
	if err != nil {
		return recommendation, err
	}
	newCost, err := GetComputeInstanceCost(aws_model.ProductInfo[aws_model.ProductAttributesInstance]{
		Attributes: *newResAttr,
		// TODO: It may be a good idea to get ProductFamily from the GetRecommendedResource function itself.
		//ProductFamily: pInfo.ProductFamily,
	})
	if currentCost.MinPrice > newCost.MinPrice {
		return aws_model.Recommendation[aws_model.InstanceDetails]{
			Source:        aws_model.RECOMMENDATION_AI,
			CloudProvider: aws_model.CLOUD_PROVIDER_AWS,
			CurrentResourceDetails: aws_model.InstanceDetails{
				InstanceType: pInfo.Attributes.InstanceType,
			},
			CurrentCost: currentCost,
			RecommendationItems: []aws_model.RecommendationItem[aws_model.InstanceDetails]{
				{
					Resource: aws_model.InstanceDetails{
						InstanceType: newResAttr.InstanceType,
					},
					Cost:                 newCost,
					EstimatedCostSavings: fmt.Sprintf("%.2f", (currentCost.MinPrice-newCost.MinPrice)*100/currentCost.MinPrice) + "%",
				},
			},
		}, nil
	}
	return recommendation, err
}

func GetRecommendedResource[T any](pAttr T, resourceUsageParams map[string]interface{}) (*T, error) {
	ai_recommendation, err := GetRecommendationFromAi(pAttr, resourceUsageParams)
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`\r?\n`)
	ai_recommendation = re.ReplaceAllString(ai_recommendation, "")
	jsonStrings, err := utils.ExtractJsonFromString(ai_recommendation)
	if err != nil {
		return nil, err
	}

	var recommendedResource T
	err = json.Unmarshal([]byte(jsonStrings[0]), &recommendedResource)
	if err != nil {
		return nil, err
	}
	return &recommendedResource, nil
}

func GetRecommendationFromAi[T any](pAttr T, resourceUsageParams map[string]interface{}) (string, error) {
	logger.NewDefaultLogger().Infof("Getting recommendation from openAI")
	requestMessage := buildRequestMessage(pAttr, resourceUsageParams)
	logger.NewDefaultLogger().Infof("AI request message : %s", requestMessage)
	return utils.QueryOpenAi(requestMessage)
	//return utils.QueryMakerSuite(requestMessage)
}

func buildRequestMessage[T any](pAttr T, resourceUsageParams map[string]interface{}) string {
	requestMessage := `{"role": "user", "content": "Current Resource: { \"Cloud\": \"AWS\",`

	// Use reflection to iterate over fields in pAttr and add them to the request message
	reflectValue := reflect.ValueOf(pAttr)
	for i := 0; i < reflectValue.NumField(); i++ {
		field := reflectValue.Field(i)
		fieldName := reflectValue.Type().Field(i).Name
		fieldValue := field.Interface()

		// Check if the field is non-empty
		if !reflect.ValueOf(fieldValue).IsZero() {
			// Add the field to the request message
			requestMessage += ` \"` + fieldName + `\": \"` + fmt.Sprintf("%v", fieldValue) + `\",`
		}
	}
	requestMessage = requestMessage[:len(requestMessage)-1] // Remove the trailing comma
	requestMessage += `}. Usage Attributes: {`
	// Add key-value pairs from resourceUsageParams
	for key, value := range resourceUsageParams {
		requestMessage += ` \"` + key + `\": \"` + fmt.Sprintf("%v", value) + `\",`
	}

	// Complete the request message
	requestMessage = requestMessage[:len(requestMessage)-1] // Remove the trailing comma
	//requestMessage += ` }. Please provide top two recommendations with cost saving value (max 20 - 25 words) to reduce the cost"}`
	requestMessage += ` }. Recommended resource to reduce the cost in the same format as Current Resource Json with no other details or words"}`

	return requestMessage
}
