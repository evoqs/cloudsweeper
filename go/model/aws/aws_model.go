package aws_model

import (
	"cloudsweep/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CloudProvider string

const (
	CLOUD_PROVIDER_AWS   CloudProvider = "AWS"
	CLOUD_PROVIDER_AZURE CloudProvider = "AZURE"
	CLOUD_PROVIDER_GCP   CloudProvider = "GCP"
)

type RecommendationSource string

const (
	RECOMMENDATION_AWSCO RecommendationSource = "AWS"
	RECOMMENDATION_AI    RecommendationSource = "AI"
)

type PricingData[T any] struct {
	FormatVersion string         `json:"FormatVersion"`
	NextToken     string         `json:"NextToken"`
	PriceList     []PriceItem[T] `json:"PriceList"`
}

type PriceItem[T any] struct {
	Product         ProductInfo[T] `json:"product"`
	PublicationDate string         `json:"publicationDate"`
	ServiceCode     string         `json:"serviceCode"`
	Terms           Terms          `json:"terms"`
	Version         string         `json:"version"`
}

type ProductInfo[T any] struct {
	Attributes    T      `json:"attributes"`
	ProductFamily string `json:"productFamily"`
}

type Terms struct {
	OnDemand map[string]OnDemandTerm `json:"OnDemand"`
}

type OnDemandTerm struct {
	EffectiveDate   string                    `json:"effectiveDate"`
	PriceDimensions map[string]PriceDimension `json:"priceDimensions"`
}

type PriceDimension struct {
	PricePerUnit PricePerUnit `json:"pricePerUnit"`
	Unit         string       `json:"unit"`
	Description  string       `json:"description"`
}

// Supported list of currency codes: https://repost.aws/knowledge-center/supported-aws-currencies
type PricePerUnit struct {
	USD string `json:"USD"`
	AUD string `json:"AUD"`
	BRL string `json:"BRL"`
	CAD string `json:"CAD"`
	CHF string `json:"CHF"`
}

type AwsResourceCost[T any] struct {
	Id                primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CloudProvider     string             `json:"cloudProvider" bson:"cloudProvider"`
	Version           string             `json:"version" bson:"version"`
	PublicationDate   string             `json:"publicationDate" bson:"publicationDate"`
	ProductFamily     string             `json:"productFamily" bson:"productFamily"`
	PricePerUnit      map[string]float64 `json:"pricePerUnit" bson:"pricePerUnit"`
	Unit              string             `json:"unit" bson:"unit"`
	TimeStamp         int64              `json:"timeStamp" bson:"timeStamp"`
	ProductAttributes T                  `json:"productAttributes" bson:"productAttributes"`
}

type ProductAttributesInstance struct {
	InstanceFamily  string `json:"instanceFamily" bson:"instanceFamily"`
	InstanceType    string `json:"instanceType" bson:"instanceType"`
	Memory          string `json:"memory" bson:"memory"`
	RegionCode      string `json:"regionCode" bson:"regionCode"`
	ServiceCode     string `json:"serviceCode" bson:"serviceCode"`
	ServiceName     string `json:"serviceName" bson:"serviceName"`
	Tenancy         string `json:"tenancy" bson:"tenancy"`
	UsageType       string `json:"usageType" bson:"usageType"`
	VCPU            string `json:"vcpu" bson:"vcpu"`
	OperatingSystem string `json:"operatingSystem" bson:"operatingSystem"`
	ClockSpeed      string `json:"clockSpeed" bson:"clockSpeed"`
}

type ProductAttributesEBS struct {
	VolumeType              string `json:"volumeType" bson:"volumeType"`
	RegionCode              string `json:"regionCode" bson:"regionCode"`
	ServiceCode             string `json:"servicecode" bson:"servicecode"`
	ServiceName             string `json:"servicename" bson:"servicename"`
	LocationType            string `json:"locationtype" bson:"locationtype"`
	UsageType               string `json:"usagetype" bson:"usagetype"`
	StorageMedia            string `json:"storageMedia" bson:"storageMedia"`
	MaxIopsBurstPerformance string `json:"maxIopsBurstPerformance" bson:"maxIopsBurstPerformance"`
	MaxIopsvolume           string `json:"maxIopsvolume" bson:"maxIopsvolume"`
	MaxThroughputvolume     string `json:"maxThroughputvolume" bson:"maxThroughputvolume"`
	VolumeApiName           string `json:"volumeApiName" bson:"volumeApiName"`
}

type ProductAttributesEBSSnapshot struct {
	Location     string `json:"location" bson:"location"`
	RegionCode   string `json:"regionCode" bson:"regionCode"`
	ServiceCode  string `json:"servicecode" bson:"servicecode"`
	ServiceName  string `json:"servicename" bson:"servicename"`
	LocationType string `json:"locationtype" bson:"locationtype"`
	UsageType    string `json:"usagetype" bson:"usagetype"`
	// Values - Amazon S3 etc
	StorageMedia string `json:"storageMedia" bson:"storageMedia"`
}

type ProductAttributesElasticIp struct {
	RegionCode string `json:"regionCode" bson:"regionCode"`
}

type Recommendation[T InstanceDetails | EBSVolumeDetails] struct {
	Id                     primitive.ObjectID      `json:"id" bson:"_id,omitempty"`
	Source                 RecommendationSource    `json:"source" bson:"source"`
	CloudProvider          CloudProvider           `json:"cloudProvider" bson:"cloudProvider"`
	AccountId              string                  `json:"accountId" bson:"accountId"`
	CurrentResourceDetails T                       `json:"currentResourceDetails" bson:"currentResourceDetails"`
	CurrentCost            model.ResourceCost      `json:"cost" bson:"cost"`
	RecommendationItems    []RecommendationItem[T] `json:"recommendationItems" bson:"recommendationItems"`
	TimeStamp              int64                   `json:"timeStamp" bson:"timeStamp"`
}

type RecommendationItem[T InstanceDetails | EBSVolumeDetails] struct {
	Resource                T                  `json:"resource" bson:"resource"`
	Cost                    model.ResourceCost `json:"cost" bson:"cost"`
	EstimatedCostSavings    string             `json:"estimatedCostSavings" bson:"estimatedCostSavings"`
	EstimatedMonthlySavings string             `json:"estimatedMonthlySavings" bson:"estimatedMonthlySavings"`
}

// Used for Recommendation
type InstanceDetails struct {
	InstaneId     string `json:"instanceId" bson:"instanceId"`
	InstanceType  string `json:"instanceType" bson:"instanceType"`
	InstanceName  string `json:"instanceName" bson:"instanceName"`
	InstanceState string `json:"instanceState" bson:"instanceState"`
	Region        string `json:"region" bson:"region"`
	InstanceArn   string `json:"instanceArn" bson:"instanceArn"`
	//Cost          model.ResourceCost `json:"cost" bson:"cost"`
}

// Used for Recommendation
type EBSVolumeDetails struct {
	VolumeId                 string `json:"volumeId" bson:"volumeId"`
	VolumeType               string `json:"volumeType" bson:"volumeType"`
	VolumeName               string `json:"volumeName" bson:"volumeName"`
	VolumeSize               int64  `json:"volumeSize" bson:"volumeSize"`
	VolumeBaselineIOPS       int64  `json:"volumeBaselineIOPS" bson:"volumeBaselineIOPS"`
	VolumeBaselineThroughput int64  `json:"volumeBaselineThroughput" bson:"volumeBaselineThroughput"`
	VolumeBurstIOPS          int64  `json:"volumeBurstIOPS" bson:"volumeBurstIOPS"`
	VolumeBurstThroughput    int64  `json:"volumeBurstThroughput" bson:"volumeBurstThroughput"`
	Region                   string `json:"region" bson:"region"`
	VolumeArn                string `json:"volumeArn" bson:"volumeArn"`
	//Cost                     model.ResourceCost `json:"cost" bson:"cost"`
}

// Policy Result Models
type AwsInstanceResult struct {
	ResultData AwsInstanceResultData `json:"resultData" bson:"resultData"`
	MetaData   *model.ResultMetaData `json:"resultMetaData" bson:"resultMetaData"`
}
type AwsInstanceResultData struct {
	InstanceId       string `json:"instanceId" bson:"instanceId"`
	InstanceType     string `json:"instanceType" bson:"instanceType"`
	Region           string `json:"region" bson:"region"`
	PlatformDetails  string `json:"platformDetails" bson:"platformDetails"`
	AvailabilityZone string `json:"availabilityZone" bson:"availabilityZone"`
	GroupName        string `json:"groupName" bson:"groupName"`
	Tenancy          string `json:"tenancy" bson:"tenancy"`
	Code             string `json:"stateCode" bson:"stateCode"`
	StateName        string `json:"stateName" bson:"stateName"`
}
type AwsInstancePolicyResultData struct {
	InstanceId      string `json:"instanceId" bson:"instanceId"`
	InstanceType    string `json:"instanceType" bson:"instanceType"`
	Region          string `json:"region" bson:"region"`
	PlatformDetails string `json:"platformDetails" bson:"platformDetails"`

	Placement struct {
		AvailabilityZone string `json:"availabilityZone" bson:"availabilityZone"`
		GroupName        string `json:"groupName" bson:"groupName"`
		Tenancy          string `json:"tenancy" bson:"tenancy"`
	} `json:"placement" bson:"placement"`

	State struct {
		Code string `json:"code" bson:"code"`
		Name string `json:"name" bson:"name"`
	} `json:"state" bson:"state"`

	Tags []struct {
		Key   string `json:"key" bson:"key"`
		Value string `json:"value" bson:"value"`
	} `json:"tags" bson:"tags"`
}

type AwsBlockVolumeResult struct {
	ResultData AwsBlockVolumeResultData `json:"resultData" bson:"resultData"`
	MetaData   *model.ResultMetaData    `json:"resultMetaData" bson:"resultMetaData"`
}
type AwsBlockVolumeResultData struct {
	VolumeId         string `json:"volumeId" bson:"volumeId"`
	VolumeType       string `json:"volumeType" bson:"volumeType"`
	Encrypted        bool   `json:"encrypted" bson:"encrypted"`
	SnapshotId       string `json:"snapshotId" bson:"snapshotId"`
	Region           string `json:"region" bson:"region"`
	State            string `json:"state" bson:"state"`
	AvailabilityZone string `json:"availabilityZone" bson:"availabilityZone"`
	Attachments      bool   `json:"attachments" bson:"attachments"`
	Size             int    `json:"size" bson:"size"`
}
type AwsBlockVolumePolicyResultData struct {
	VolumeId         string                     `json:"volumeId" bson:"volumeId"`
	VolumeType       string                     `json:"volumeType" bson:"volumeType"`
	Encrypted        bool                       `json:"encrypted" bson:"encrypted"`
	SnapshotId       string                     `json:"snapshotId" bson:"snapshotId"`
	Region           string                     `json:"region" bson:"region"`
	State            string                     `json:"state" bson:"state"`
	AvailabilityZone string                     `json:"availabilityZone" bson:"availabilityZone"`
	Attachments      []AwsBlockVolumeAttachment `json:"attachments" bson:"attachments"`
	Size             int                        `json:"size" bson:"size"`
}

type AwsBlockVolumeAttachment struct {
	InstanceId string `json:"instanceId" bson:"instanceId"`
	State      string `json:"state" bson:"state"`
	Device     string `json:"device" bson:"device"`
}

type AwsElasticIPResult struct {
	ResultData AwsElasticIPResultData `json:"resultData" bson:"resultData"`
	MetaData   *model.ResultMetaData  `json:"resultMetaData" bson:"resultMetaData"`
}
type AwsElasticIPResultData struct {
	PublicIp           string `json:"publicIp" bson:"publicIp"`
	PublicIpv4Pool     string `json:"publicIpv4pool" bson:"publicIpv4pool"`
	Domain             string `json:"domain" bson:"domain"`
	NetworkBorderGroup string `json:"networkBorderGroup" bson:"networkBorderGroup"`
	AllocationId       string `json:"allocationId" bson:"allocationId"`
}

type AwsSnapshotResult struct {
	ResultData AwsSnapshotResultData `json:"resultData" bson:"resultData"`
	MetaData   *model.ResultMetaData `json:"resultMetaData" bson:"resultMetaData"`
}
type AwsSnapshotResultData struct {
	Description string `json:"description" bson:"description"`
	Encrypted   string `json:"encrypted" bson:"encrypted"`
	SnapshotId  string `json:"snapshotId" bson:"snapshotId"`
	State       string `json:"state" bson:"state"`
	VolumeId    string `json:"volumeId" bson:"volumeId"`
	VolumeSize  int    `json:"volumeSize" bson:"volumeSize"`
	StorageTier string `json:"storageTier" bson:"storageTier"`
}
