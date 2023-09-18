package aws_model

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
	InstanceFamily  string `json:"instanceFamily"`
	InstanceType    string `json:"instanceType"`
	Memory          string `json:"memory"`
	RegionCode      string `json:"regionCode"`
	ServiceCode     string `json:"serviceCode"`
	ServiceName     string `json:"serviceName"`
	Tenancy         string `json:"tenancy"`
	UsageType       string `json:"usageType"`
	VCPU            string `json:"vcpu"`
	OperatingSystem string `json:"operatingSystem"`
	ClockSpeed      string `json:"clockSpeed"`
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
	Version           string                    `json:"version" bson:"version"`
	PublicationDate   string                    `json:"publicationdate" bson:"publicationDate"`
	ProductFamily     string                    `json:"productfamily" bson:"productFamily"`
	PricePerUnit      map[string]float64        `json:"priceperunit" bson:"pricePerUnit"`
	Unit              string                    `json:"unit" bson:"unit"`
	ProductAttributes ProductAttributesInstance `json:"productattributes" bson:"productAttributes"`
}
