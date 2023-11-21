package cost_estimator

import (
	"cloudsweep/cloud_lib"
	logger "cloudsweep/logging"
	aws_model "cloudsweep/model/aws"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/computeoptimizer"
)

// =========================== Cost Optimizer ===================================

func GetAWSCOResultForAllEC2Instances(awsAccessKeyId string, awsSecretAccessKey string, regions []string) (map[string][]*computeoptimizer.InstanceRecommendation, error) {
	resultMap := make(map[string][]*computeoptimizer.InstanceRecommendation)
	var errors error
	if regions == nil || len(regions) == 0 {
		return resultMap, fmt.Errorf("No regions specified")
	}
	for _, region := range regions {
		// Describe recommendations for a specific EC2 instance
		input := &computeoptimizer.GetEC2InstanceRecommendationsInput{
			//https://docs.aws.amazon.com/compute-optimizer/latest/APIReference/API_Filter.html
			Filters: []*computeoptimizer.Filter{
				{
					Name:   aws.String("RecommendationSourceType"),
					Values: []*string{aws.String("Ec2Instance")},
				},
			},
			NextToken: nil,
		}
		awsClient, err := cloud_lib.GetAwsClient(awsAccessKeyId, awsSecretAccessKey, region)
		if err != nil {
			errors = fmt.Errorf("%w\n%w", errors, fmt.Errorf("Region:%s - %w", region, err))
			continue
		}
		coClient := awsClient.GetComputeOptimizerClient()

		var regionRecommendations []*computeoptimizer.InstanceRecommendation
		for {
			result, err := coClient.GetEC2InstanceRecommendations(input)
			if err != nil {
				errors = fmt.Errorf("%w\n%w", errors, fmt.Errorf("Region:%s - %w", region, err))
				continue
				//return allRecommendations, fmt.Errorf("Error getting CO recommendation for account id:%s. %v\n", accountId, err)
			}
			regionRecommendations = append(regionRecommendations, result.InstanceRecommendations...)
			if result.NextToken == nil {
				break
			}
			input.NextToken = result.NextToken
		}
		resultMap[region] = regionRecommendations
	}
	return resultMap, nil
}

func GetAWSCOResultForEC2Instance(awsAccessKeyId string, awsSecretAccessKey string, region string, accountId string, instanceId string) ([]*computeoptimizer.InstanceRecommendation, error) {
	var errors error
	// Describe recommendations for a specific EC2 instance
	input := &computeoptimizer.GetEC2InstanceRecommendationsInput{
		InstanceArns: []*string{
			aws.String(fmt.Sprintf("arn:aws:ec2:%s:%s:instance/%s", region, accountId, instanceId)),
		},
	}

	awsClient, err := cloud_lib.GetAwsClient(awsAccessKeyId, awsSecretAccessKey, region)
	if err != nil {
		return nil, fmt.Errorf("%w\n%w", errors, fmt.Errorf("Region:%s - %w", region, err))
	}
	coClient := awsClient.GetComputeOptimizerClient()

	result, err := coClient.GetEC2InstanceRecommendations(input)
	if err != nil {
		return nil, fmt.Errorf("Error getting CO recommendation for account id:%s, instance id:%s. %v\n", accountId, instanceId, err)
	}
	return result.InstanceRecommendations, fmt.Errorf("No recommendations available for the given instance %s", instanceId)
}

// ============================= Recommendations ====================================

func GetAWSRecommendationForAllEC2Instances(awsAccessKeyId string, awsSecretAccessKey string, regions []string) ([]*aws_model.Recommendation[aws_model.CurrentInstanceDetails], error) {
	var recommendations []*aws_model.Recommendation[aws_model.CurrentInstanceDetails]

	coResult, err := GetAWSCOResultForAllEC2Instances(awsAccessKeyId, awsSecretAccessKey, regions)

	for region, coItems := range coResult {
		for _, coItem := range coItems {
			currentCost, err := GetComputeInstanceCost(aws_model.ProductAttributesInstance{
				InstanceType:    *coItem.CurrentInstanceType,
				RegionCode:      region,
				OperatingSystem: "Linux",
			})
			if err != nil {
				logger.NewDefaultLogger().Errorf(err.Error())
			}

			recommendationItems := []aws_model.RecommendationItem{}
			for _, recom := range coItem.RecommendationOptions {
				newCost, err := GetComputeInstanceCost(aws_model.ProductAttributesInstance{
					InstanceType:    *recom.InstanceType,
					RegionCode:      region,
					OperatingSystem: "Linux",
				})
				if err != nil {
					logger.NewDefaultLogger().Errorf(err.Error())
				}
				recommendationItems = append(recommendationItems, aws_model.RecommendationItem{
					Resource: *recom.InstanceType,
					NewCost:  newCost,
					// TODO: current or new cost might be nil.
					EstimatedCostSavings: fmt.Sprintf("%.2f", (currentCost.MinPrice-newCost.MinPrice)*100/currentCost.MinPrice) + "%",
				})
			}
			recommendation := &aws_model.Recommendation[aws_model.CurrentInstanceDetails]{
				CurrentResourceDetails: aws_model.CurrentInstanceDetails{
					InstanceType:  *coItem.CurrentInstanceType,
					InstanceName:  *coItem.InstanceName,
					InstanceState: *coItem.InstanceState,
					InstanceArn:   *coItem.InstanceArn,
					Region:        region,
					CurrentCost:   currentCost,
				},
				RecommendationItems: recommendationItems,
			}
			recommendations = append(recommendations, recommendation)
		}
	}
	return recommendations, err
}

func GetAWSRecommendationForEC2Instance(awsAccessKeyId string, awsSecretAccessKey string, region string, accountId string, instanceId string) (*aws_model.Recommendation[aws_model.CurrentInstanceDetails], error) {
	var recommendation *aws_model.Recommendation[aws_model.CurrentInstanceDetails]

	coResult, err := GetAWSCOResultForEC2Instance(awsAccessKeyId, awsSecretAccessKey, region, accountId, instanceId)

	for _, coItem := range coResult {
		currentCost, err := GetComputeInstanceCost(aws_model.ProductAttributesInstance{
			InstanceType:    *coItem.CurrentInstanceType,
			RegionCode:      *aws.String(region),
			OperatingSystem: "Linux",
		})
		if err != nil {
			logger.NewDefaultLogger().Errorf(err.Error())
		}

		recommendationItems := []aws_model.RecommendationItem{}
		for _, recom := range coItem.RecommendationOptions {
			newCost, err := GetComputeInstanceCost(aws_model.ProductAttributesInstance{
				InstanceType:    *recom.InstanceType,
				RegionCode:      region,
				OperatingSystem: "Linux",
			})
			if err != nil {
				logger.NewDefaultLogger().Errorf(err.Error())
			}
			recommendationItems = append(recommendationItems, aws_model.RecommendationItem{
				Resource: *recom.InstanceType,
				NewCost:  newCost,
				// TODO: current or new cost might be nil.
				EstimatedCostSavings: fmt.Sprintf("%.2f", (currentCost.MinPrice-newCost.MinPrice)*100/currentCost.MinPrice) + "%",
			})
		}
		recommendation = &aws_model.Recommendation[aws_model.CurrentInstanceDetails]{
			CurrentResourceDetails: aws_model.CurrentInstanceDetails{
				InstanceType:  *coItem.CurrentInstanceType,
				InstanceName:  *coItem.InstanceName,
				InstanceState: *coItem.InstanceState,
				InstanceArn:   *coItem.InstanceArn,
				Region:        region,
				CurrentCost:   currentCost,
			},
			RecommendationItems: recommendationItems,
		}
	}
	return recommendation, err
}
