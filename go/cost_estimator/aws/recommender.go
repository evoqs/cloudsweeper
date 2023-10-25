package cost_estimator

import (
	logger "cloudsweep/logging"
	"cloudsweep/utils"
	"encoding/json"
	"fmt"
	"reflect"
)

type Recommendation struct {
	CurrentCost          ResourceCost
	Recommendation       string `json:"recommendation"`
	NewCost              ResourceCost
	EstimatedCostSavings string `json:"estimated_cost_savings"`
}

func GetRecommendation[T any](pAttr T, resourceUsageParams map[string]interface{}) (string, error) {
	logger.NewDefaultLogger().Infof("Getting recommendation from openAI")
	requestMessage := buildRequestMessage(pAttr, resourceUsageParams)
	logger.NewDefaultLogger().Infof("AI request message : %s", requestMessage)
	postResponse, err := utils.QueryOpenAI(requestMessage)
	if err != nil {
		return "", err
	}

	// Unmarshal the response JSON
	fmt.Printf("postResponse.Body:%v", string(postResponse.Body))
	var responseMap map[string]interface{}
	if err := json.Unmarshal(postResponse.Body, &responseMap); err != nil {
		return "", err
	}

	// Extract the "content" from the response
	content := ""
	if choices, ok := responseMap["choices"].([]interface{}); ok {
		if len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if message, ok := choice["message"].(map[string]interface{}); ok {
					if contentStr, ok := message["content"].(string); ok {
						content = contentStr
					}
				}
			}
		}
	}
	return content, nil
	// Unmarshal the response JSON
	//var recommendationMap map[string]interface{}
	/*var recommendation Recommendation
	if err := json.Unmarshal([]byte(content), &recommendation); err != nil {
		return nil, err
	}

	return &recommendation, nil*/
}

func buildRequestMessage[T any](pAttr T, resourceUsageParams map[string]interface{}) string {
	// Start building the request message with common information
	requestMessage := `{"role": "user", "content": "Current resource details { Cloud: AWS,`

	// Use reflection to iterate over fields in pAttr and add them to the request message
	reflectValue := reflect.ValueOf(pAttr)
	for i := 0; i < reflectValue.NumField(); i++ {
		field := reflectValue.Field(i)
		fieldName := reflectValue.Type().Field(i).Name
		fieldValue := field.Interface()

		// Check if the field is non-empty
		if !reflect.ValueOf(fieldValue).IsZero() {
			// Add the field to the request message
			requestMessage += ` ` + fieldName + `: ` + fmt.Sprintf("%v", fieldValue) + `,`
		}
	}

	// Add key-value pairs from resourceUsageParams
	for key, value := range resourceUsageParams {
		requestMessage += ` ` + key + `: ` + fmt.Sprintf("%v", value) + `,`
	}

	// Complete the request message
	requestMessage = requestMessage[:len(requestMessage)-1] // Remove the trailing comma
	requestMessage += ` }. Please provide top two recommendations with cost saving value (max 20 - 25 words) to reduce the cost"}`

	return requestMessage
}
