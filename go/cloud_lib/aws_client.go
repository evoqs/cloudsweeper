package cloud_lib

import (
	"cloudsweep/config"
	logger "cloudsweep/logging"
	aws_model "cloudsweep/model/aws"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/computeoptimizer"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/aws/aws-sdk-go/service/sts"
)

var awsClientsMap sync.Map

type AwsClient struct {
	session     *session.Session
	once        sync.Once
	credentials *credentials.Credentials
}

func (ac *AwsClient) GetEC2Client() *ec2.EC2 {
	return ec2.New(ac.session)
}

// TODO: where and how to use this function?
func (ac *AwsClient) CheckAuthFailure(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		return awsErr.Code() == "AuthFailure"
	}
	return false
}

func (ac *AwsClient) GetAwsAccountID() (string, error) {
	stsClient := sts.New(ac.session)
	input := &sts.GetCallerIdentityInput{}

	result, err := stsClient.GetCallerIdentity(input)
	if err != nil {
		return "", err
	}

	return *result.Account, nil
}

func (ac *AwsClient) ensureValidSession() error {
	var err error
	if ac.credentials == nil || ac.credentials.IsExpired() {
		logger.NewDefaultLogger().Infof("Token empty or expired")
		// Recreate the session with new credentials
		ac.once.Do(func() {
			ac.session, err = session.NewSession(&aws.Config{
				Region:      ac.session.Config.Region,
				Credentials: ac.credentials,
			})
		})
	}
	return err
}

func (ac *AwsClient) ValidateCredentials() bool {
	stsClient := sts.New(ac.session)
	input := &sts.GetCallerIdentityInput{}

	_, err := stsClient.GetCallerIdentity(input)
	if err != nil {
		// Check if the error is due to authentication failure
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "AuthFailure" {
				logger.NewDefaultLogger().Errorf("AWS credentials validation failed: %v", err)
				return false
			}
		}
		logger.NewDefaultLogger().Errorf("Error validating AWS credentials: %v", err)
		return false
	}

	return true
}

func (ac *AwsClient) GetEC2ClientWithRegion(region string) *ec2.EC2 {
	return ec2.New(ac.session, aws.NewConfig().WithRegion(region))
}

func (ac *AwsClient) GetPriceClient() *pricing.Pricing {
	return pricing.New(ac.session)
}

func (ac *AwsClient) GetComputeOptimizerClient() *computeoptimizer.ComputeOptimizer {
	return computeoptimizer.New(ac.session)
}

func (ac *AwsClient) GetComputeOptimizerClientForRegion(region string) *computeoptimizer.ComputeOptimizer {
	return computeoptimizer.New(ac.session, aws.NewConfig().WithRegion(region))
}

func (ac *AwsClient) GetAllRegions() ([]*ec2.Region, error) {
	input := &ec2.DescribeRegionsInput{AllRegions: aws.Bool(true)}
	regions, err := ac.GetEC2Client().DescribeRegions(input)
	if err != nil {
		return nil, err
	}
	return regions.Regions, err
}

func (ac *AwsClient) GetSubscribedRegions() ([]*ec2.Region, error) {
	input := &ec2.DescribeRegionsInput{AllRegions: aws.Bool(false)}
	regions, err := ac.GetEC2Client().DescribeRegions(input)
	if err != nil {
		return nil, err
	}
	return regions.Regions, err
}

func (ac *AwsClient) GetAllRegionCodes() ([]string, error) {
	allRegions, err := ac.GetAllRegions()
	if err != nil {
		return nil, err
	}

	var regionCodes []string
	for _, region := range allRegions {
		regionCodes = append(regionCodes, aws.StringValue(region.RegionName))
	}

	return regionCodes, nil
}

func (ac *AwsClient) GetSubscribedRegionCodes() ([]string, error) {
	allRegions, err := ac.GetSubscribedRegions()
	if err != nil {
		return nil, err
	}

	var regionCodes []string
	for _, region := range allRegions {
		regionCodes = append(regionCodes, aws.StringValue(region.RegionName))
	}

	return regionCodes, nil
}

func (ac *AwsClient) GetAllInstanceTypes(region string, filters []*ec2.Filter, instanceTypes []string) ([]*ec2.InstanceTypeInfo, error) {
	ec2Client := ac.GetEC2ClientWithRegion(region)
	var allInstanceTypes []*ec2.InstanceTypeInfo
	var nextToken *string
	var instanceTypeFilter []*string
	if instanceTypes == nil || len(instanceTypes) == 0 {
		instanceTypeFilter = nil
	} else {
		instanceTypeFilter = aws.StringSlice(instanceTypes)
	}
	if len(filters) == 0 {
		filters = nil
	}
	for {
		input := &ec2.DescribeInstanceTypesInput{
			Filters:       filters,
			InstanceTypes: instanceTypeFilter,
			NextToken:     nextToken,
		}
		result, err := ec2Client.DescribeInstanceTypes(input)
		if err != nil {
			return nil, err
		}
		allInstanceTypes = append(allInstanceTypes, result.InstanceTypes...)
		if result.NextToken == nil {
			break
		}
		nextToken = result.NextToken
	}
	return allInstanceTypes, nil
}

func (ac *AwsClient) GetAllServiceCodes() ([]string, error) {
	pClient := ac.GetPriceClient()
	var serviceCodes []string
	// Loop until there are no more results
	for {
		input := &pricing.DescribeServicesInput{
			MaxResults: aws.Int64(100),
			NextToken:  nil,
		}
		result, err1 := pClient.DescribeServices(input)
		if err1 != nil {
			logger.NewDefaultLogger().Errorf("Error listing AWS services: %v", err1)
			return serviceCodes, err1
		}

		// Loop through the AWS services and add their service codes to the array
		for _, service := range result.Services {
			serviceCodes = append(serviceCodes, *service.ServiceCode)
		}
		if result.NextToken != nil {
			input.NextToken = result.NextToken
		} else {
			// No more results, break the loop
			break
		}
	}
	return serviceCodes, nil
}

func (ac *AwsClient) GetAllVolumeTypes() ([]string, error) {
	return ac.getAttributeValues("AmazonEC2", "volumeApiName")
}

func (ac *AwsClient) GetAllStorageMedia() ([]string, error) {
	return ac.getAttributeValues("AmazonEC2", "storageMedia")
}

func (ac *AwsClient) getAttributeValues(serviceCode string, attributeName string) ([]string, error) {
	var storageMediaValues []string
	pClient := ac.GetPriceClient()
	// Use the GetAttributeValues operation to retrieve possible values
	input := &pricing.GetAttributeValuesInput{
		ServiceCode:   &serviceCode,
		AttributeName: &attributeName,
		MaxResults:    nil,
		NextToken:     nil,
	}

	// Retrieve and process the results
	for {
		result, err := pClient.GetAttributeValues(input)
		if err != nil {
			logger.NewDefaultLogger().Errorf("Error fetching attribute values: %s", err)
			return storageMediaValues, err
		}

		// Loop through the values and print them
		for _, value := range result.AttributeValues {
			storageMediaValues = append(storageMediaValues, *value.Value)
			//fmt.Println("StorageMedia:", *value.Value)
		}

		if result.NextToken == nil {
			break
		}
		input.NextToken = result.NextToken
	}
	return storageMediaValues, nil
}

func (ac *AwsClient) GetProductFamilyList(serviceCode string) ([]string, error) {
	var productFamilies []string
	var pricingData aws_model.PricingData[aws_model.ProductAttributesInstance]
	pClient := ac.GetPriceClient()
	input := &pricing.GetProductsInput{
		ServiceCode: &serviceCode,
	}
	for {
		pricingResult, err := pClient.GetProducts(input)
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
		for _, priceItem := range pricingData.PriceList {
			//fmt.Println(priceItem.Product.ProductFamily)
			productFamilies = append(productFamilies, priceItem.Product.ProductFamily)
		}
		if pricingResult.NextToken != nil {
			fmt.Printf("Next Page Token: %s\n", *pricingResult.NextToken)
			input.NextToken = pricingResult.NextToken
		} else {
			// No more pages, break the loop
			break
		}
	}

	return productFamilies, nil
}

// =============== AWS Wrapper functions =======================

func GetAwsClient(awsAccessKeyId string, awsSecretAccessKey string, region string) (*AwsClient, error) {
	awsClient := &AwsClient{}
	awsClient.credentials = credentials.NewStaticCredentials(awsAccessKeyId, awsSecretAccessKey, "")
	var err error
	creds := credentials.NewStaticCredentials(awsAccessKeyId, awsSecretAccessKey, "")
	awsClient.session, err = session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: creds,
	})
	if err != nil {
		return awsClient, err
	}
	return awsClient, nil
}

func GetCSAdminAwsClient() (*AwsClient, error) {
	key := fmt.Sprintf("%s-%s-%s", config.GetConfig().Aws.Creds.Aws_access_key_id, config.GetConfig().Aws.Creds.Aws_secret_access_key, config.GetConfig().Aws.Creds.Aws_default_region)
	// Check if an instance with the given key exists
	if existingClient, ok := awsClientsMap.Load(key); ok {
		awsClient := existingClient.(*AwsClient)
		awsClient.ensureValidSession()
		return existingClient.(*AwsClient), nil
	}
	awsClient, err := GetAwsClient(config.GetConfig().Aws.Creds.Aws_access_key_id, config.GetConfig().Aws.Creds.Aws_secret_access_key, config.GetConfig().Aws.Creds.Aws_default_region)
	// TODO: cache needs to expire after specific amount of time
	awsClientsMap.Store(key, awsClient)
	return awsClient, err
}
