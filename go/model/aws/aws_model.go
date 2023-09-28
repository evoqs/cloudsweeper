package aws_model

import "go.mongodb.org/mongo-driver/bson/primitive"

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
