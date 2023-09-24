package aws_model

import "go.mongodb.org/mongo-driver/bson/primitive"

type PricingDataInstance struct {
	FormatVersion string              `json:"FormatVersion"`
	NextToken     string              `json:"NextToken"`
	PriceList     []PriceItemInstance `json:"PriceList"`
}

type PriceItemInstance struct {
	Product         ProductInfoInstance `json:"product"`
	PublicationDate string              `json:"publicationDate"`
	ServiceCode     string              `json:"serviceCode"`
	Terms           TermsInstance       `json:"terms"`
	Version         string              `json:"version"`
}

type ProductInfoInstance struct {
	Attributes    ProductAttributesInstance `json:"attributes"`
	ProductFamily string                    `json:"productFamily"`
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

type TermsInstance struct {
	OnDemand map[string]OnDemandTermInstance `json:"OnDemand"`
}

type OnDemandTermInstance struct {
	EffectiveDate   string                            `json:"effectiveDate"`
	PriceDimensions map[string]PriceDimensionInstance `json:"priceDimensions"`
}

type PriceDimensionInstance struct {
	PricePerUnit PricePerUnitInstance `json:"pricePerUnit"`
	Unit         string               `json:"unit"`
	Description  string               `json:"description"`
}

// Supported list of currency codes: https://repost.aws/knowledge-center/supported-aws-currencies
type PricePerUnitInstance struct {
	USD string `json:"USD"`
	AUD string `json:"AUD"`
	BRL string `json:"BRL"`
	CAD string `json:"CAD"`
	CHF string `json:"CHF"`
}

type ResourceCostInstance struct {
	Id                primitive.ObjectID        `json:"id" bson:"_id,omitempty"`
	CloudProvider     string                    `json:"cloudProvider" bson:"cloudProvider"`
	Version           string                    `json:"version" bson:"version"`
	PublicationDate   string                    `json:"publicationDate" bson:"publicationDate"`
	ProductFamily     string                    `json:"productFamily" bson:"productFamily"`
	PricePerUnit      map[string]float64        `json:"pricePerUnit" bson:"pricePerUnit"`
	Unit              string                    `json:"unit" bson:"unit"`
	ProductAttributes ProductAttributesInstance `json:"productAttributes" bson:"productAttributes"`
}
