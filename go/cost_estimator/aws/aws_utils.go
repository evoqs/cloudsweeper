package cost_estimator

import (
	logger "cloudsweep/logging"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
)

func getResourceIdFromArn(arnString string) (string, error) {
	parsedARN, err := arn.Parse(arnString)
	if err != nil {
		return "", err
	}
	resourceId := strings.Split(parsedARN.Resource, "/")[1]
	return resourceId, nil
}

func GetResourceIdFromArn(arnString string) string {
	id, err := getResourceIdFromArn(arnString)
	if err != nil {
		logger.NewDefaultLogger().Errorf("AWS Arn parse error: %v", err)
	}
	return id
}
