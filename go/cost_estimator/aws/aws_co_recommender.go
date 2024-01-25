package cost_estimator

import (
	"cloudsweep/cloud_lib"
	"reflect"
	"strings"
	"time"

	logger "cloudsweep/logging"
	"cloudsweep/model"
	aws_model "cloudsweep/model/aws"
	"cloudsweep/storage"
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
	return resultMap, errors
}

func GetAWSCOResultForEC2Instance(awsAccessKeyId string, awsSecretAccessKey string, region string, accountId string, instanceId string) ([]*computeoptimizer.InstanceRecommendation, error) {
	// Describe recommendations for a specific EC2 instance
	input := &computeoptimizer.GetEC2InstanceRecommendationsInput{
		InstanceArns: []*string{
			aws.String(fmt.Sprintf("arn:aws:ec2:%s:%s:instance/%s", region, accountId, instanceId)),
		},
	}
	awsClient, err := cloud_lib.GetAwsClient(awsAccessKeyId, awsSecretAccessKey, region)
	if err != nil {
		return nil, fmt.Errorf("%w\n", fmt.Errorf("Region:%s - %w", region, err))
	}
	coClient := awsClient.GetComputeOptimizerClient()

	result, err := coClient.GetEC2InstanceRecommendations(input)
	if err != nil {
		return nil, fmt.Errorf("Error getting CO recommendation for account id:%s, instance id:%s. %v\n", accountId, instanceId, err)
	}
	return result.InstanceRecommendations, nil
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
	return resultMap, errors
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

// =============================================================================================================func buildFilter(filter interface{}, parentKey string) string {
func buildFilter(filter interface{}) string {
	return buildFilterWithParentTag(filter, "")
}

func buildFilterWithParentTag(filter interface{}, parentKey string) string {
	var queryParts []string
	reflectValue := reflect.ValueOf(filter)
	buildFilterRecursive(reflectValue, &queryParts, parentKey)
	return "{" + strings.Join(queryParts, ", ") + "}"
}

func buildFilterRecursive(value reflect.Value, queryParts *[]string, parentKey string) {
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldValue := field.Interface()

		// Skip empty or zero values
		if reflect.DeepEqual(fieldValue, reflect.Zero(field.Type()).Interface()) {
			continue
		}

		jsonTagName := value.Type().Field(i).Tag.Get("json")
		fieldKind := field.Kind()
		key := jsonTagName
		if parentKey != "" {
			key = parentKey + "." + jsonTagName
		}

		switch fieldKind {
		case reflect.Struct:
			buildFilterRecursive(field, queryParts, key)
		default:
			*queryParts = append(*queryParts, fmt.Sprintf("\"%s\": \"%v\"", key, fieldValue))
		}
	}
}

// Read from DB
// TODO: return value to be part of the parameter of the function. so that we can verify the type and build the query
// OR the resource ID should represent the InstanceDetails or EBSVolumeDetails - sometimes it can be empty
func GetRecommendationsFromDB[T aws_model.InstanceDetails | aws_model.EBSVolumeDetails](recommendationFilter aws_model.Recommendation[T]) ([]*aws_model.Recommendation[T], error) {
	logger.NewDefaultLogger().Debugf("Query DB for Recommendation for Resource. Filter: %v", recommendationFilter)
	var recommendations []*aws_model.Recommendation[T]
	opr := storage.GetDefaultDBOperators()
	query := buildFilter(recommendationFilter)
	logger.NewDefaultLogger().Debugf("DB Query: %s", query)
	err := opr.RecommendationOperator.GetQueryResult(query, &recommendations)
	logger.NewDefaultLogger().Debugf("Length of Recommendations from DB: %d\n", len(recommendations))
	return recommendations, err
}

// ============================= Recommendations ====================================

func GetAWSRecommendationForAllEC2Instances(awsAccessKeyId string, awsSecretAccessKey string, accountId string, regions []string) ([]*aws_model.Recommendation[aws_model.InstanceDetails], error) {
	logger.NewDefaultLogger().Debugf("[GetAWSRecommendationForAllEC2Instances] Running the function")
	var recommendations []*aws_model.Recommendation[aws_model.InstanceDetails]
	// TODO: Account Id shouldn't be fetched from AWS everytime, we should use customer account object here
	/*awsClient, err := cloud_lib.GetAwsClient(awsAccessKeyId, awsSecretAccessKey, "")
	if err != nil {
		return nil, err
	}
	accountId, err := awsClient.GetAwsAccountID()
	if err != nil {
		return nil, err
	}*/
	// ======================== Get From DB =========================
	var errs []error
	for _, region := range regions {
		regionRecommendations, err := GetRecommendationsFromDB[aws_model.InstanceDetails](aws_model.Recommendation[aws_model.InstanceDetails]{
			Source:        aws_model.RECOMMENDATION_AWSCO,
			CloudProvider: aws_model.CLOUD_PROVIDER_AWS,
			AccountId:     accountId,
			CurrentResourceDetails: aws_model.InstanceDetails{
				Region: region,
			},
		})
		if err != nil {
			logger.NewDefaultLogger().Errorf("Error while getting recommendations from DB: %v", err)
			errs = append(errs, err)
		}
		recommendations = append(recommendations, regionRecommendations...)
	}
	// TODO: We may need to just ignore the errors here? or any case of error get the recommendations from AWS
	if len(errs) == 0 && len(recommendations) > 0 {
		logger.NewDefaultLogger().Infof("Total Number of Recommendation returned from DB: %d", len(recommendations))
		return recommendations, nil
	}

	// ======================== Get From AWS =========================
	coResult, err := GetAWSCOResultForAllEC2Instances(awsAccessKeyId, awsSecretAccessKey, regions)
	logger.NewDefaultLogger().Debugf("Error: %v", err)

	for region, coItems := range coResult {
		for _, coItem := range coItems {
			var currentCost model.ResourceCost
			recommendationItems := []aws_model.RecommendationItem[aws_model.InstanceDetails]{}
			for _, recom := range coItem.RecommendationOptions {
				var newCost model.ResourceCost

				if *recom.SavingsOpportunity.SavingsOpportunityPercentage > 0 && *recom.SavingsOpportunity.EstimatedMonthlySavings.Value > 0 {
					if currentCost == (model.ResourceCost{}) {
						currentCost = model.ResourceCost{
							MinPrice: (*recom.SavingsOpportunity.EstimatedMonthlySavings.Value * 100) / *recom.SavingsOpportunity.SavingsOpportunityPercentage,
							Currency: *recom.SavingsOpportunity.EstimatedMonthlySavings.Currency,
							// TODO: Check if you can avoid recomputing
							MaxPrice: (*recom.SavingsOpportunity.EstimatedMonthlySavings.Value * 100) / *recom.SavingsOpportunity.SavingsOpportunityPercentage,
							Unit:     "Per Month",
						}
					}
					newCost = model.ResourceCost{
						MinPrice: currentCost.MinPrice - *recom.SavingsOpportunity.EstimatedMonthlySavings.Value,
						Currency: *recom.SavingsOpportunity.EstimatedMonthlySavings.Currency,
						MaxPrice: currentCost.MinPrice - *recom.SavingsOpportunity.EstimatedMonthlySavings.Value,
						Unit:     "Per Month",
					}
				}
				// Assign new Cost estimate from general Cost Provider
				if newCost == (model.ResourceCost{}) {
					newCost, err = GetComputeInstanceCost(aws_model.ProductInfo[aws_model.ProductAttributesInstance]{
						Attributes: aws_model.ProductAttributesInstance{
							InstanceType:    *recom.InstanceType,
							RegionCode:      region,
							OperatingSystem: "Linux",
						}})
					if err != nil {
						logger.NewDefaultLogger().Errorf(err.Error())
					}
				}

				recommendationItems = append(recommendationItems, aws_model.RecommendationItem[aws_model.InstanceDetails]{
					Resource: aws_model.InstanceDetails{
						InstanceType: *recom.InstanceType,
					},
					Cost: newCost,
					//EstimatedCostSavings: fmt.Sprintf("%.2f", (currentCost.MinPrice-newCost.MinPrice)*100/currentCost.MinPrice) + "%",
					EstimatedCostSavings:    fmt.Sprintf("%.2f", *recom.SavingsOpportunity.SavingsOpportunityPercentage) + "%",
					EstimatedMonthlySavings: fmt.Sprintf("%f %s", *recom.SavingsOpportunity.EstimatedMonthlySavings.Value, *recom.SavingsOpportunity.EstimatedMonthlySavings.Currency),
				})
			}
			// Assign current Cost estimate from general Cost Provider
			if currentCost == (model.ResourceCost{}) {
				currentCost, err = GetComputeInstanceCost(aws_model.ProductInfo[aws_model.ProductAttributesInstance]{
					Attributes: aws_model.ProductAttributesInstance{
						InstanceType:    *coItem.CurrentInstanceType,
						RegionCode:      region,
						OperatingSystem: "Linux",
					}})
				if err != nil {
					logger.NewDefaultLogger().Errorf(err.Error())
				}

			}
			recommendation := &aws_model.Recommendation[aws_model.InstanceDetails]{
				Source:        aws_model.RECOMMENDATION_AWSCO,
				CloudProvider: aws_model.CLOUD_PROVIDER_AWS,
				AccountId:     accountId,
				CurrentResourceDetails: aws_model.InstanceDetails{
					InstaneId:     GetResourceIdFromArn(*coItem.InstanceArn),
					InstanceType:  *coItem.CurrentInstanceType,
					InstanceName:  *coItem.InstanceName,
					InstanceState: *coItem.InstanceState,
					InstanceArn:   *coItem.InstanceArn,
					Region:        region,
				},
				CurrentCost:         currentCost,
				RecommendationItems: recommendationItems,
				TimeStamp:           time.Now().Unix(),
			}

			opr := storage.GetDefaultDBOperators()
			opr.RecommendationOperator.AddRecommendation(recommendation)
			recommendations = append(recommendations, recommendation)
		}
	}
	/*recommendation := aws_model.Recommendation[aws_model.InstanceDetails]{
		Source:        aws_model.RECOMMENDATION_AWSCO,
		CloudProvider: aws_model.CLOUD_PROVIDER_AWS,
		AccountId:     "867226238913",
		CurrentResourceDetails: aws_model.InstanceDetails{
			InstaneId:     "12345670",
			InstanceType:  "dummyw",
			InstanceName:  "dummy33",
			InstanceState: "dum3e23my",
			InstanceArn:   "dumwdwqsdcsmy:3e3e:sds:3e2e23:instance/i-23e21dedw",
			Region:        "ap-northeast-1",
		},
		CurrentCost: model.ResourceCost{
			MinPrice: 10,
		},
		RecommendationItems: []aws_model.RecommendationItem[aws_model.InstanceDetails]{
			{
				Resource: aws_model.InstanceDetails{
					InstanceType:  "t2.micro",
					InstanceName:  "SampleInstance",
					InstanceState: "Running",
					Region:        "us-east-1",
					InstanceArn:   "arn:aws:ec2:us-east-1:123456789012:instance/i-0123456789abcdef0",
				},
				Cost: model.ResourceCost{
					// Fill in cost details as needed
				},
				EstimatedCostSavings:    "100 USD",
				EstimatedMonthlySavings: "50 USD",
			},
			// Add more items if needed
		},
		TimeStamp: time.Now().Unix(),
	}

	opr := storage.GetDefaultDBOperators()
	logger.NewDefaultLogger().Infof("Object %+v\n", recommendation)
	opr.RecommendationOperator.AddRecommendation(recommendation)*/
	logger.NewDefaultLogger().Infof("Total Number of Recommendation returned from AWS: %d", len(recommendations))
	return recommendations, err
}

func GetAWSRecommendationForEC2Instance(awsAccessKeyId string, awsSecretAccessKey string, region string, accountId string, instanceId string) (*aws_model.Recommendation[aws_model.InstanceDetails], error) {
	logger.NewDefaultLogger().Debugf("[GetAWSRecommendationForEC2Instance] Running the function")
	var recommendation *aws_model.Recommendation[aws_model.InstanceDetails]
	// TODO: Account Id shouldn't be fetched from AWS everytime, we should use customer account object here
	/*awsClient, err := cloud_lib.GetAwsClient(awsAccessKeyId, awsSecretAccessKey, "")
	if err != nil {
		return nil, err
	}
	accounId, err := awsClient.GetAwsAccountID()
	if err != nil {
		return nil, err
	}*/
	// ======================== Get From DB =========================
	regionRecommendations, err := GetRecommendationsFromDB[aws_model.InstanceDetails](aws_model.Recommendation[aws_model.InstanceDetails]{
		Source:        aws_model.RECOMMENDATION_AWSCO,
		CloudProvider: aws_model.CLOUD_PROVIDER_AWS,
		AccountId:     accountId,
		CurrentResourceDetails: aws_model.InstanceDetails{
			Region:    region,
			InstaneId: instanceId,
		},
	})
	if err != nil {
		logger.NewDefaultLogger().Errorf("Error while getting recommendations from DB: %v", err)
	} else if len(regionRecommendations) > 0 {
		// Assuming there should be alway single recommendation entry for an instanceId
		logger.NewDefaultLogger().Infof("Returning Recommendation from DB for Instance.")
		return regionRecommendations[0], nil
	}

	// ======================== Get From AWS =========================
	coResult, err := GetAWSCOResultForEC2Instance(awsAccessKeyId, awsSecretAccessKey, region, accountId, instanceId)
	logger.NewDefaultLogger().Debugf("Error: %v", err)

	for _, coItem := range coResult {
		var currentCost model.ResourceCost
		recommendationItems := []aws_model.RecommendationItem[aws_model.InstanceDetails]{}
		for _, recom := range coItem.RecommendationOptions {
			var newCost model.ResourceCost

			if *recom.SavingsOpportunity.SavingsOpportunityPercentage > 0 && *recom.SavingsOpportunity.EstimatedMonthlySavings.Value > 0 {
				if currentCost == (model.ResourceCost{}) {
					currentCost = model.ResourceCost{
						MinPrice: (*recom.SavingsOpportunity.EstimatedMonthlySavings.Value * 100) / *recom.SavingsOpportunity.SavingsOpportunityPercentage,
						Currency: *recom.SavingsOpportunity.EstimatedMonthlySavings.Currency,
						// TODO: Check if you can avoid recomputing
						MaxPrice: (*recom.SavingsOpportunity.EstimatedMonthlySavings.Value * 100) / *recom.SavingsOpportunity.SavingsOpportunityPercentage,
						Unit:     "Per Month",
					}
				}
				newCost = model.ResourceCost{
					MinPrice: currentCost.MinPrice - *recom.SavingsOpportunity.EstimatedMonthlySavings.Value,
					Currency: *recom.SavingsOpportunity.EstimatedMonthlySavings.Currency,
					MaxPrice: currentCost.MinPrice - *recom.SavingsOpportunity.EstimatedMonthlySavings.Value,
					Unit:     "Per Month",
				}
			}
			// Assign new Cost estimate from general Cost Provider
			if newCost == (model.ResourceCost{}) {
				newCost, err = GetComputeInstanceCost(aws_model.ProductInfo[aws_model.ProductAttributesInstance]{
					Attributes: aws_model.ProductAttributesInstance{
						InstanceType:    *recom.InstanceType,
						RegionCode:      region,
						OperatingSystem: "Linux",
					}})
				if err != nil {
					logger.NewDefaultLogger().Errorf(err.Error())
				}
			}

			recommendationItems = append(recommendationItems, aws_model.RecommendationItem[aws_model.InstanceDetails]{
				Resource: aws_model.InstanceDetails{
					InstanceType: *recom.InstanceType,
				},
				Cost: newCost,
				//EstimatedCostSavings: fmt.Sprintf("%.2f", (currentCost.MinPrice-newCost.MinPrice)*100/currentCost.MinPrice) + "%",
				EstimatedCostSavings:    fmt.Sprintf("%.2f", *recom.SavingsOpportunity.SavingsOpportunityPercentage) + "%",
				EstimatedMonthlySavings: fmt.Sprintf("%f %s", *recom.SavingsOpportunity.EstimatedMonthlySavings.Value, *recom.SavingsOpportunity.EstimatedMonthlySavings.Currency),
			})
		}
		// Assign current Cost estimate from general Cost Provider
		if currentCost == (model.ResourceCost{}) {
			currentCost, err = GetComputeInstanceCost(aws_model.ProductInfo[aws_model.ProductAttributesInstance]{
				Attributes: aws_model.ProductAttributesInstance{
					InstanceType:    *coItem.CurrentInstanceType,
					RegionCode:      region,
					OperatingSystem: "Linux",
				}})
			if err != nil {
				logger.NewDefaultLogger().Errorf(err.Error())
			}

		}
		recommendation = &aws_model.Recommendation[aws_model.InstanceDetails]{
			Source:        aws_model.RECOMMENDATION_AWSCO,
			CloudProvider: aws_model.CLOUD_PROVIDER_AWS,
			AccountId:     accountId,
			CurrentResourceDetails: aws_model.InstanceDetails{
				InstaneId:     GetResourceIdFromArn(*coItem.InstanceArn),
				InstanceType:  *coItem.CurrentInstanceType,
				InstanceName:  *coItem.InstanceName,
				InstanceState: *coItem.InstanceState,
				InstanceArn:   *coItem.InstanceArn,
				Region:        region,
			},
			CurrentCost:         currentCost,
			RecommendationItems: recommendationItems,
			TimeStamp:           time.Now().Unix(),
		}
		opr := storage.GetDefaultDBOperators()
		opr.RecommendationOperator.AddRecommendation(recommendation)
		// For single resource, there will always be single result. We can break/return here.
		break
	}
	logger.NewDefaultLogger().Infof("Returning Recommendation from AWS for Instance.")
	return recommendation, err
}

func GetAWSRecommendationForAllEBSVolumes(awsAccessKeyId string, awsSecretAccessKey string, regions []string) ([]*aws_model.Recommendation[aws_model.EBSVolumeDetails], error) {
	logger.NewDefaultLogger().Debugf("[GetAWSRecommendationForAllEBSVolumes] Running the function")
	var recommendations []*aws_model.Recommendation[aws_model.EBSVolumeDetails]
	coResult, err := GetAWSCOResultForAllEBSVolumes(awsAccessKeyId, awsSecretAccessKey, regions)
	logger.NewDefaultLogger().Debugf("Error: %v", err)

	// TODO: Account Id shouldn't be fetched from AWS everytime, we should use customer account object here
	awsClient, err := cloud_lib.GetAwsClient(awsAccessKeyId, awsSecretAccessKey, "")
	if err != nil {
		return nil, err
	}
	accounId, err := awsClient.GetAwsAccountID()
	if err != nil {
		return nil, err
	}

	for region, coItems := range coResult {
		for _, coItem := range coItems {
			var currentCost model.ResourceCost
			recommendationItems := []aws_model.RecommendationItem[aws_model.EBSVolumeDetails]{}
			for _, recom := range coItem.VolumeRecommendationOptions {
				var newCost model.ResourceCost

				if *recom.SavingsOpportunity.SavingsOpportunityPercentage > 0 && *recom.SavingsOpportunity.EstimatedMonthlySavings.Value > 0 {
					if currentCost == (model.ResourceCost{}) {
						currentCost = model.ResourceCost{
							MinPrice: (*recom.SavingsOpportunity.EstimatedMonthlySavings.Value * 100) / *recom.SavingsOpportunity.SavingsOpportunityPercentage,
							Currency: *recom.SavingsOpportunity.EstimatedMonthlySavings.Currency,
							// TODO: Check if you can avoid recomputing
							MaxPrice: (*recom.SavingsOpportunity.EstimatedMonthlySavings.Value * 100) / *recom.SavingsOpportunity.SavingsOpportunityPercentage,
							Unit:     "Per Month",
						}
					}
					newCost = model.ResourceCost{
						MinPrice: currentCost.MinPrice - *recom.SavingsOpportunity.EstimatedMonthlySavings.Value,
						Currency: *recom.SavingsOpportunity.EstimatedMonthlySavings.Currency,
						MaxPrice: currentCost.MinPrice - *recom.SavingsOpportunity.EstimatedMonthlySavings.Value,
						Unit:     "Per Month",
					}
				}
				// Assign new Cost estimate from general Cost Provider
				if newCost == (model.ResourceCost{}) {
					newCost, err = GetEbsCost(aws_model.ProductInfo[aws_model.ProductAttributesEBS]{
						Attributes: aws_model.ProductAttributesEBS{
							VolumeApiName: *recom.Configuration.VolumeType,
							RegionCode:    region,
						},
						ProductFamily: "Storage"})
					if err != nil {
						logger.NewDefaultLogger().Errorf(err.Error())
					}
				}
				recommendationItems = append(recommendationItems, aws_model.RecommendationItem[aws_model.EBSVolumeDetails]{
					Resource: aws_model.EBSVolumeDetails{
						VolumeType: *recom.Configuration.VolumeType,
					},
					Cost: newCost,
					//EstimatedCostSavings: fmt.Sprintf("%.2f", (currentCost.MinPrice-newCost.MinPrice)*100/currentCost.MinPrice) + "%",
					EstimatedCostSavings:    fmt.Sprintf("%.2f", *recom.SavingsOpportunity.SavingsOpportunityPercentage) + "%",
					EstimatedMonthlySavings: fmt.Sprintf("%f %s", *recom.SavingsOpportunity.EstimatedMonthlySavings.Value, *recom.SavingsOpportunity.EstimatedMonthlySavings.Currency),
				})
			}
			// Assign current Cost estimate from general Cost Provider
			if currentCost == (model.ResourceCost{}) {
				currentCost, err = GetEbsCost(aws_model.ProductInfo[aws_model.ProductAttributesEBS]{
					Attributes: aws_model.ProductAttributesEBS{
						VolumeApiName: *coItem.CurrentConfiguration.VolumeType,
						RegionCode:    region,
					},
					ProductFamily: "Storage"})
				if err != nil {
					logger.NewDefaultLogger().Errorf(err.Error())
				}
			}
			recommendation := &aws_model.Recommendation[aws_model.EBSVolumeDetails]{
				Source:        aws_model.RECOMMENDATION_AWSCO,
				CloudProvider: aws_model.CLOUD_PROVIDER_AWS,
				AccountId:     accounId,
				CurrentResourceDetails: aws_model.EBSVolumeDetails{
					VolumeId:   GetResourceIdFromArn(*coItem.VolumeArn),
					VolumeType: *coItem.CurrentConfiguration.VolumeType,
					//VolumeName:  *coItem.,
					VolumeSize:               *coItem.CurrentConfiguration.VolumeSize,
					VolumeBaselineIOPS:       *coItem.CurrentConfiguration.VolumeBaselineIOPS,
					VolumeBaselineThroughput: *coItem.CurrentConfiguration.VolumeBaselineThroughput,
					VolumeBurstIOPS:          *coItem.CurrentConfiguration.VolumeBurstIOPS,
					VolumeBurstThroughput:    *coItem.CurrentConfiguration.VolumeBurstThroughput,
					VolumeArn:                *coItem.VolumeArn,
					Region:                   region,
				},
				CurrentCost:         currentCost,
				RecommendationItems: recommendationItems,
				TimeStamp:           time.Now().Unix(),
			}
			opr := storage.GetDefaultDBOperators()
			opr.RecommendationOperator.AddRecommendation(recommendation)
			recommendations = append(recommendations, recommendation)
		}
	}
	return recommendations, err
}

// Get the AWS CO Recommendations for a single EBS Volume using the id
func GetAWSRecommendationForEBSVolume(awsAccessKeyId string, awsSecretAccessKey string, region string, accountId string, volumeId string) (*aws_model.Recommendation[aws_model.EBSVolumeDetails], error) {
	logger.NewDefaultLogger().Debugf("[GetAWSRecommendationForEBSVolume] Running the function")
	var recommendation *aws_model.Recommendation[aws_model.EBSVolumeDetails]
	coResult, err := GetAWSCOResultForEBSVolume(awsAccessKeyId, awsSecretAccessKey, region, accountId, volumeId)
	logger.NewDefaultLogger().Debugf("Error: %v", err)

	// TODO: Account Id shouldn't be fetched from AWS everytime, we should use customer account object here
	awsClient, err := cloud_lib.GetAwsClient(awsAccessKeyId, awsSecretAccessKey, "")
	if err != nil {
		return nil, err
	}
	accounId, err := awsClient.GetAwsAccountID()
	if err != nil {
		return nil, err
	}

	for _, coItem := range coResult {
		var currentCost model.ResourceCost
		recommendationItems := []aws_model.RecommendationItem[aws_model.EBSVolumeDetails]{}
		for _, recom := range coItem.VolumeRecommendationOptions {
			var newCost model.ResourceCost

			if *recom.SavingsOpportunity.SavingsOpportunityPercentage > 0 && *recom.SavingsOpportunity.EstimatedMonthlySavings.Value > 0 {
				if currentCost == (model.ResourceCost{}) {
					currentCost = model.ResourceCost{
						MinPrice: (*recom.SavingsOpportunity.EstimatedMonthlySavings.Value * 100) / *recom.SavingsOpportunity.SavingsOpportunityPercentage,
						Currency: *recom.SavingsOpportunity.EstimatedMonthlySavings.Currency,
						// TODO: Check if you can avoid recomputing
						MaxPrice: (*recom.SavingsOpportunity.EstimatedMonthlySavings.Value * 100) / *recom.SavingsOpportunity.SavingsOpportunityPercentage,
						Unit:     "Per Month",
					}
				}
				newCost = model.ResourceCost{
					MinPrice: currentCost.MinPrice - *recom.SavingsOpportunity.EstimatedMonthlySavings.Value,
					Currency: *recom.SavingsOpportunity.EstimatedMonthlySavings.Currency,
					MaxPrice: currentCost.MinPrice - *recom.SavingsOpportunity.EstimatedMonthlySavings.Value,
					Unit:     "Per Month",
				}
			}
			// Assign new Cost estimate from general Cost Provider
			if newCost == (model.ResourceCost{}) {
				newCost, err = GetEbsCost(aws_model.ProductInfo[aws_model.ProductAttributesEBS]{
					Attributes: aws_model.ProductAttributesEBS{
						VolumeApiName: *recom.Configuration.VolumeType,
						RegionCode:    region,
					},
					ProductFamily: "Storage"})
				if err != nil {
					logger.NewDefaultLogger().Errorf(err.Error())
				}
			}

			recommendationItems = append(recommendationItems, aws_model.RecommendationItem[aws_model.EBSVolumeDetails]{
				Resource: aws_model.EBSVolumeDetails{
					VolumeId:                 GetResourceIdFromArn(*coItem.VolumeArn),
					VolumeType:               *recom.Configuration.VolumeType,
					VolumeSize:               *recom.Configuration.VolumeSize,
					VolumeBaselineIOPS:       *recom.Configuration.VolumeBaselineIOPS,
					VolumeBaselineThroughput: *recom.Configuration.VolumeBaselineThroughput,
					VolumeBurstIOPS:          *recom.Configuration.VolumeBurstIOPS,
					VolumeBurstThroughput:    *recom.Configuration.VolumeBurstThroughput,
				},
				Cost: newCost,
				//EstimatedCostSavings: fmt.Sprintf("%.2f", (currentCost.MinPrice-newCost.MinPrice)*100/currentCost.MinPrice) + "%",
				EstimatedCostSavings:    fmt.Sprintf("%.2f", *recom.SavingsOpportunity.SavingsOpportunityPercentage) + "%",
				EstimatedMonthlySavings: fmt.Sprintf("%f %s", *recom.SavingsOpportunity.EstimatedMonthlySavings.Value, *recom.SavingsOpportunity.EstimatedMonthlySavings.Currency),
			})
		}
		// Assign current Cost estimate from general Cost Provider
		if currentCost == (model.ResourceCost{}) {
			currentCost, err = GetEbsCost(aws_model.ProductInfo[aws_model.ProductAttributesEBS]{
				Attributes: aws_model.ProductAttributesEBS{
					VolumeApiName: *coItem.CurrentConfiguration.VolumeType,
					RegionCode:    region,
				},
				ProductFamily: "Storage"})
			if err != nil {
				logger.NewDefaultLogger().Errorf(err.Error())
			}
		}

		recommendation = &aws_model.Recommendation[aws_model.EBSVolumeDetails]{
			Source:        aws_model.RECOMMENDATION_AWSCO,
			CloudProvider: aws_model.CLOUD_PROVIDER_AWS,
			AccountId:     accounId,
			CurrentResourceDetails: aws_model.EBSVolumeDetails{
				VolumeType: *coItem.CurrentConfiguration.VolumeType,
				//VolumeName: coItem.CurrentConfiguration.RootVolume,
				VolumeSize:               *coItem.CurrentConfiguration.VolumeSize,
				VolumeBaselineIOPS:       *coItem.CurrentConfiguration.VolumeBaselineIOPS,
				VolumeBaselineThroughput: *coItem.CurrentConfiguration.VolumeBaselineThroughput,
				VolumeBurstIOPS:          *coItem.CurrentConfiguration.VolumeBurstIOPS,
				VolumeBurstThroughput:    *coItem.CurrentConfiguration.VolumeBurstThroughput,
				VolumeArn:                *coItem.VolumeArn,
				Region:                   region,
			},
			CurrentCost:         currentCost,
			RecommendationItems: recommendationItems,
			TimeStamp:           time.Now().Unix(),
		}
		opr := storage.GetDefaultDBOperators()
		opr.RecommendationOperator.AddRecommendation(recommendation)
		// For single resource, there will always be single result. We can break/return here.
		break
	}
	return recommendation, err
}
