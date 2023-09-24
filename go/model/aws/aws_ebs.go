package aws_model

import "go.mongodb.org/mongo-driver/bson/primitive"

type PricingDataEBS struct {
	FormatVersion string         `json:"FormatVersion"`
	NextToken     string         `json:"NextToken"`
	PriceList     []PriceItemEBS `json:"PriceList"`
}

type PriceItemEBS struct {
	Product         ProductInfoEBS `json:"product"`
	PublicationDate string         `json:"publicationDate"`
	ServiceCode     string         `json:"serviceCode"`
	Terms           TermsEBS       `json:"terms"`
	Version         string         `json:"version"`
}

type ProductInfoEBS struct {
	Attributes    ProductAttributesEBS `json:"attributes"`
	ProductFamily string               `json:"productFamily"`
}

type ProductAttributesEBS struct {
	VolumeType              string `json:"volumeType"`
	RegionCode              string `json:"regionCode"`
	ServiceCode             string `json:"servicecode"`
	ServiceName             string `json:"servicename"`
	LocationType            string `json:"locationtype"`
	UsageType               string `json:"usagetype"`
	StorageMedia            string `json:"storageMedia"`
	MaxIopsBurstPerformance string `json:"maxIopsBurstPerformance"`
	MaxIopsvolume           string `json:"maxIopsvolume"`
	MaxThroughputvolume     string `json:"maxThroughputvolume"`
	VolumeApiName           string `json:"volumeApiName"`
}

type TermsEBS struct {
	OnDemand map[string]OnDemandTermEBS `json:"OnDemand"`
}

type OnDemandTermEBS struct {
	EffectiveDate   string                       `json:"effectiveDate"`
	PriceDimensions map[string]PriceDimensionEBS `json:"priceDimensions"`
}

type PriceDimensionEBS struct {
	PricePerUnit PricePerUnitEBS `json:"pricePerUnit"`
	Unit         string          `json:"unit"`
	Description  string          `json:"description"`
}

type PricePerUnitEBS struct {
	USD string `json:"USD"`
}

type ResourceCostEBS struct {
	Id                primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Version           string               `json:"version" bson:"version"`
	PublicationDate   string               `json:"publicationdate" bson:"publicationdate"`
	ProductFamily     string               `json:"productfamily" bson:"productfamily"`
	PricePerUnit      float64              `json:"priceperunit" bson:"priceperunit"`
	PriceCurrency     string               `json:"pricecurrency" bson:"pricecurrency"`
	Unit              string               `json:"unit" bson:"unit"`
	ProductAttributes ProductAttributesEBS `json:"productattributes" bson:"productattributes"`
}
