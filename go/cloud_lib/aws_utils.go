package cloud_lib

import (
	logger "cloudsweep/logging"
	aws_model "cloudsweep/model/aws"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/pricing"
)

func GetAwsSession() (*session.Session, error) {
	// Synchronize the function call
	// Create AWS credentials with your access key and secret key.
	// TODO: Read it from config file
	creds := credentials.NewStaticCredentials("AKIA4T2VWH7A6GQYCS7Z", "YAf6nke9U5SgXN3zGWZ+nYISOPTsWt55d2xQBzmt", "")

	// Create an AWS session with your credentials and desired region.
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"), // TOD0: Read the CS home region from the config file
		Credentials: creds,
	})
	// TODO: Make it a singleton
	return sess, err
}

func GetEC2Client() (*ec2.EC2, error) {
	// Create an EC2 service client
	sess, err := GetAwsSession()
	if err != nil {
		logger.NewDefaultLogger().Errorf("Problem in Creating AWS Session %s", err)
	}
	return ec2.New(sess), err
}

func GetEC2ClientWithRegion(region string) (*ec2.EC2, error) {
	// Create an EC2 service client
	sess, err := GetAwsSession()
	if err != nil {
		logger.NewDefaultLogger().Errorf("Problem in Creating AWS Session %s", err)
	}
	return ec2.New(sess, aws.NewConfig().WithRegion(region)), err
}

func GetPriceClient() (*pricing.Pricing, error) {
	sess, err := GetAwsSession()
	if err != nil {
		logger.NewDefaultLogger().Errorf("Problem in Creating AWS Session %s", err)
	}
	return pricing.New(sess), err
}

func GetAllRegions() ([]*ec2.Region, error) {
	// TODO: Check for error
	ec2Client, err := GetEC2Client()
	input := &ec2.DescribeRegionsInput{AllRegions: aws.Bool(true)}
	regions, err := ec2Client.DescribeRegions(input)

	return regions.Regions, err
}

func GetSubscribedRegions() ([]*ec2.Region, error) {
	// TODO: Check for error
	ec2Client, err := GetEC2Client()
	input := &ec2.DescribeRegionsInput{AllRegions: aws.Bool(false)}
	regions, err := ec2Client.DescribeRegions(input)

	return regions.Regions, err
}

func GetAllInstanceTypes(region string) []*ec2.InstanceTypeInfo {
	// TODO: Check for error
	ec2Client, err := GetEC2ClientWithRegion(region)

	input := &ec2.DescribeInstanceTypesInput{}
	result, err := ec2Client.DescribeInstanceTypes(input)
	if err != nil {
		logger.NewDefaultLogger().Errorf("Error listing instance types: %v", err)
		return nil
	}
	return result.InstanceTypes
}

func GetAllServiceCodes() ([]string, error) {
	pClient, err := GetPriceClient()
	if err != nil {
		logger.NewDefaultLogger().Errorf("Error Creating PriceClient : %v", err)
		return nil, err
	}

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

func GetAllVolumeTypes() ([]string, error) {
	return getAttributeValues("AmazonEC2", "volumeApiName")
}

func GetAllStorageMedia() ([]string, error) {
	return getAttributeValues("AmazonEC2", "storageMedia")
}

func getAttributeValues(serviceCode string, attributeName string) ([]string, error) {
	var storageMediaValues []string
	pClient, err := GetPriceClient()
	if err != nil {
		logger.NewDefaultLogger().Errorf("Error Creating PriceClient : %v", err)
		return storageMediaValues, err
	}

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

func GetProductFamilyList(serviceCode string) ([]string, error) {
	var productFamilies []string
	var pricingData aws_model.PricingData[aws_model.ProductAttributesInstance]
	pClient, err := GetPriceClient()
	if err != nil {
		logger.NewDefaultLogger().Errorf("Error Creating PriceClient : %v", err)
		return productFamilies, err
	}
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
