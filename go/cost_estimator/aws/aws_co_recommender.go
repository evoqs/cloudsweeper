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

func GetAWSCOResultForAllEBSVolumes(awsAccessKeyId string, awsSecretAccessKey string, regions []string) (map[string][]*computeoptimizer.VolumeRecommendation, error) {
	resultMap := make(map[string][]*computeoptimizer.VolumeRecommendation)
	var errors error
	if regions == nil || len(regions) == 0 {
		return resultMap, fmt.Errorf("No regions specified")
	}
	for _, region := range regions {
		// Describe recommendations for all EBS volumes
		input := &computeoptimizer.GetEBSVolumeRecommendationsInput{
			// TODO: Add any additional filters or parameters as needed
		}
		awsClient, err := cloud_lib.GetAwsClient(awsAccessKeyId, awsSecretAccessKey, region)
		if err != nil {
			errors = fmt.Errorf("%w\n%w", errors, fmt.Errorf("Region:%s - %w", region, err))
			continue
		}
		coClient := awsClient.GetComputeOptimizerClient()

		var regionRecommendations []*computeoptimizer.VolumeRecommendation
		for {
			result, err := coClient.GetEBSVolumeRecommendations(input)
			if err != nil {
				errors = fmt.Errorf("%w\n%w", errors, fmt.Errorf("Region:%s - %w", region, err))
				continue
			}
			regionRecommendations = append(regionRecommendations, result.VolumeRecommendations...)
			if result.NextToken == nil {
				break
			}
			input.NextToken = result.NextToken
		}
		resultMap[region] = regionRecommendations
	}
	return resultMap, nil
}

func GetAWSCOResultForEBSVolume(awsAccessKeyId string, awsSecretAccessKey string, region string, accountId string, volumeId string) ([]*computeoptimizer.VolumeRecommendation, error) {
	var errors error
	// Describe recommendations for a specific EBS volume
	input := &computeoptimizer.GetEBSVolumeRecommendationsInput{
		VolumeArns: []*string{
			aws.String(fmt.Sprintf("arn:aws:ec2:%s:%s:volume/%s", region, accountId, volumeId)),
		},
	}

	awsClient, err := cloud_lib.GetAwsClient(awsAccessKeyId, awsSecretAccessKey, region)
	if err != nil {
		return nil, fmt.Errorf("%w\n%w", errors, fmt.Errorf("Region:%s - %w", region, err))
	}
	coClient := awsClient.GetComputeOptimizerClient()

	result, err := coClient.GetEBSVolumeRecommendations(input)
	if err != nil {
		return nil, fmt.Errorf("Error getting CO recommendation for volume id:%s. %v\n", volumeId, err)
	}
	return result.VolumeRecommendations, nil
}

// ============================= Recommendations ====================================

func GetAWSRecommendationForAllEC2Instances(awsAccessKeyId string, awsSecretAccessKey string, regions []string) ([]*aws_model.Recommendation[aws_model.InstanceDetails], error) {
	var recommendations []*aws_model.Recommendation[aws_model.InstanceDetails]

	coResult, err := GetAWSCOResultForAllEC2Instances(awsAccessKeyId, awsSecretAccessKey, regions)

	for region, coItems := range coResult {
		for _, coItem := range coItems {
			currentCost, err := GetComputeInstanceCost(aws_model.ProductInfo[aws_model.ProductAttributesInstance]{
				Attributes: aws_model.ProductAttributesInstance{
					InstanceType:    *coItem.CurrentInstanceType,
					RegionCode:      region,
					OperatingSystem: "Linux",
				}})
			if err != nil {
				logger.NewDefaultLogger().Errorf(err.Error())
			}

			recommendationItems := []aws_model.RecommendationItem[aws_model.InstanceDetails]{}
			for _, recom := range coItem.RecommendationOptions {
				newCost, err := GetComputeInstanceCost(aws_model.ProductInfo[aws_model.ProductAttributesInstance]{
					Attributes: aws_model.ProductAttributesInstance{
						InstanceType:    *recom.InstanceType,
						RegionCode:      region,
						OperatingSystem: "Linux",
					}})
				if err != nil {
					logger.NewDefaultLogger().Errorf(err.Error())
				}
				recommendationItems = append(recommendationItems, aws_model.RecommendationItem[aws_model.InstanceDetails]{
					Resource: aws_model.InstanceDetails{
						InstanceType: *recom.InstanceType,
					},
					Cost: newCost,
					// TODO: current or new cost might be nil.
					EstimatedCostSavings: fmt.Sprintf("%.2f", (currentCost.MinPrice-newCost.MinPrice)*100/currentCost.MinPrice) + "%",
				})
			}
			recommendation := &aws_model.Recommendation[aws_model.InstanceDetails]{
				CurrentResourceDetails: aws_model.InstanceDetails{
					InstanceType:  *coItem.CurrentInstanceType,
					InstanceName:  *coItem.InstanceName,
					InstanceState: *coItem.InstanceState,
					InstanceArn:   *coItem.InstanceArn,
					Region:        region,
				},
				CurrentCost:         currentCost,
				RecommendationItems: recommendationItems,
			}
			recommendations = append(recommendations, recommendation)
		}
	}
	return recommendations, err
}

func GetAWSRecommendationForEC2Instance(awsAccessKeyId string, awsSecretAccessKey string, region string, accountId string, instanceId string) (*aws_model.Recommendation[aws_model.InstanceDetails], error) {
	var recommendation *aws_model.Recommendation[aws_model.InstanceDetails]

	coResult, err := GetAWSCOResultForEC2Instance(awsAccessKeyId, awsSecretAccessKey, region, accountId, instanceId)

	for _, coItem := range coResult {
		currentCost, err := GetComputeInstanceCost(aws_model.ProductInfo[aws_model.ProductAttributesInstance]{
			Attributes: aws_model.ProductAttributesInstance{
				InstanceType:    *coItem.CurrentInstanceType,
				RegionCode:      *aws.String(region),
				OperatingSystem: "Linux",
			}})
		if err != nil {
			logger.NewDefaultLogger().Errorf(err.Error())
		}

		recommendationItems := []aws_model.RecommendationItem[aws_model.InstanceDetails]{}
		for _, recom := range coItem.RecommendationOptions {
			newCost, err := GetComputeInstanceCost(aws_model.ProductInfo[aws_model.ProductAttributesInstance]{
				Attributes: aws_model.ProductAttributesInstance{
					InstanceType:    *recom.InstanceType,
					RegionCode:      region,
					OperatingSystem: "Linux",
				}})
			if err != nil {
				logger.NewDefaultLogger().Errorf(err.Error())
			}
			recommendationItems = append(recommendationItems, aws_model.RecommendationItem[aws_model.InstanceDetails]{
				Resource: aws_model.InstanceDetails{
					InstanceType: *recom.InstanceType,
				},
				Cost: newCost,
				// TODO: current or new cost might be nil.
				EstimatedCostSavings: fmt.Sprintf("%.2f", (currentCost.MinPrice-newCost.MinPrice)*100/currentCost.MinPrice) + "%",
			})
		}
		recommendation = &aws_model.Recommendation[aws_model.InstanceDetails]{
			CurrentResourceDetails: aws_model.InstanceDetails{
				InstanceType:  *coItem.CurrentInstanceType,
				InstanceName:  *coItem.InstanceName,
				InstanceState: *coItem.InstanceState,
				InstanceArn:   *coItem.InstanceArn,
				Region:        region,
			},
			CurrentCost:         currentCost,
			RecommendationItems: recommendationItems,
		}
	}
	return recommendation, err
}

func GetAWSRecommendationForAllEBSVolumes(awsAccessKeyId string, awsSecretAccessKey string, regions []string) ([]*aws_model.Recommendation[aws_model.EBSVolumeDetails], error) {
	var recommendations []*aws_model.Recommendation[aws_model.EBSVolumeDetails]

	coResult, err := GetAWSCOResultForAllEBSVolumes(awsAccessKeyId, awsSecretAccessKey, regions)

	for region, coItems := range coResult {
		for _, coItem := range coItems {
			currentCost, err := GetEbsCost(aws_model.ProductInfo[aws_model.ProductAttributesEBS]{
				Attributes: aws_model.ProductAttributesEBS{
					VolumeApiName: *coItem.CurrentConfiguration.VolumeType,
					RegionCode:    region,
					// TODO: Check if possible to add more filters
				},
				ProductFamily: "Storage"})
			if err != nil {
				logger.NewDefaultLogger().Errorf(err.Error())
			}

			recommendationItems := []aws_model.RecommendationItem[aws_model.EBSVolumeDetails]{}
			for _, recom := range coItem.VolumeRecommendationOptions {
				newCost, err := GetEbsCost(aws_model.ProductInfo[aws_model.ProductAttributesEBS]{
					Attributes: aws_model.ProductAttributesEBS{
						VolumeApiName: *recom.Configuration.VolumeType,
						RegionCode:    region,
						// TODO: Check if possible to add more filters
					},
					ProductFamily: "Storage"})
				if err != nil {
					logger.NewDefaultLogger().Errorf(err.Error())
				}
				recommendationItems = append(recommendationItems, aws_model.RecommendationItem[aws_model.EBSVolumeDetails]{
					Resource: aws_model.EBSVolumeDetails{
						VolumeType: *recom.Configuration.VolumeType,
					},
					Cost: newCost,
					// TODO: current or new cost might be nil.
					EstimatedCostSavings: fmt.Sprintf("%.2f", (currentCost.MinPrice-newCost.MinPrice)*100/currentCost.MinPrice) + "%",
				})
			}
			recommendation := &aws_model.Recommendation[aws_model.EBSVolumeDetails]{
				CurrentResourceDetails: aws_model.EBSVolumeDetails{
					VolumeType: *coItem.CurrentConfiguration.VolumeType,
					//VolumeName:  *coItem.,
					VolumeArn: *coItem.VolumeArn,
					Region:    region,
				},
				CurrentCost:         currentCost,
				RecommendationItems: recommendationItems,
			}
			recommendations = append(recommendations, recommendation)
		}
	}
	return recommendations, err
}

func GetAWSRecommendationForEBSVolume(awsAccessKeyId string, awsSecretAccessKey string, region string, accountId string, volumeId string) (*aws_model.Recommendation[aws_model.EBSVolumeDetails], error) {
	var recommendation *aws_model.Recommendation[aws_model.EBSVolumeDetails]

	coResult, err := GetAWSCOResultForEBSVolume(awsAccessKeyId, awsSecretAccessKey, region, accountId, volumeId)

	for _, coItem := range coResult {
		currentCost, err := GetEbsCost(aws_model.ProductInfo[aws_model.ProductAttributesEBS]{
			Attributes: aws_model.ProductAttributesEBS{
				VolumeApiName: *coItem.CurrentConfiguration.VolumeType,
				RegionCode:    region,
			},
			ProductFamily: "Storage"})
		if err != nil {
			logger.NewDefaultLogger().Errorf(err.Error())
		}

		recommendationItems := []aws_model.RecommendationItem[aws_model.EBSVolumeDetails]{}
		for _, recom := range coItem.VolumeRecommendationOptions {
			newCost, err := GetEbsCost(aws_model.ProductInfo[aws_model.ProductAttributesEBS]{
				Attributes: aws_model.ProductAttributesEBS{
					VolumeApiName: *recom.Configuration.VolumeType,
					RegionCode:    region,
				},
				ProductFamily: "Storage"})
			if err != nil {
				logger.NewDefaultLogger().Errorf(err.Error())
			}
			recommendationItems = append(recommendationItems, aws_model.RecommendationItem[aws_model.EBSVolumeDetails]{
				Resource: aws_model.EBSVolumeDetails{
					VolumeType: *recom.Configuration.VolumeType,
				},
				Cost: newCost,
				// TODO: current or new cost might be nil.
				EstimatedCostSavings: fmt.Sprintf("%.2f", (currentCost.MinPrice-newCost.MinPrice)*100/currentCost.MinPrice) + "%",
			})
		}
		recommendation = &aws_model.Recommendation[aws_model.EBSVolumeDetails]{
			CurrentResourceDetails: aws_model.EBSVolumeDetails{
				VolumeType: *coItem.CurrentConfiguration.VolumeType,
				VolumeArn:  *coItem.VolumeArn,
				Region:     region,
			},
			CurrentCost:         currentCost,
			RecommendationItems: recommendationItems,
		}
	}
	return recommendation, err
}
