package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const AWS = "aws"

type CloudAccountData struct {
	Name           string             `json:"name" bson:"name"`
	SweepAccountID string             `bson:"sweepaccountid"`
	CloudAccountID primitive.ObjectID `json:"cloudaccountid,omitempty" bson:"_id,omitempty"`
	AccountType    string             `json:"accounttype" bson:"accounttype"`
	Description    string             `json:"description" bson:"description"`
	AwsCredentials AwsCredentials     `json:"awscredentials" bson:"awscredentials"`
	EmailList      []string           `json:"emaillist"`
}

type AwsCredentials struct {
	//Region          string `json:"aws_region" bson:"aws_region"`
	AccountID       string `json:"aws_account_id" bson:"aws_account_id"`
	AccessKeyID     string `json:"aws_access_key_id" bson:"aws_access_key_id"`
	SecretAccessKey string `json:"aws_secret_access_key" bson:"aws_secret_access_key"`
}

type Policy struct {
	PolicyName        string             `json:"policyname" bson:"policyname"`
	PolicyDescription string             `json:"description" bson:"description"`
	PolicyID          primitive.ObjectID `json:"policyid" bson:"_id,omitempty"`
	SweepAccountID    string             `bson:"sweepaccountid"`
	IsDefault         bool               `json:"isDefault" bson:"isDefault"`
	PolicyDefinition  string             `json:"policydefinition" bson:"policydefinition"`
	Recommendation    string             `json:"recommendation" bson:"recommendation"`
}

type PolicyResult struct {
	PolicyResultID    primitive.ObjectID `json:"policyresultid" bson:"_id,omitempty"`
	PolicyID          string             `json:"policyid" bson:"policyid"`
	PipelIneID        string             `json:"pipelineid" bson:"pipelineid"`
	Resource          string             `json:"resource" bson:"resource"`
	Resultlist        []RegionResult     `json:"resultlist" bson:"resultlist"`
	LastRunStatus     string             `json:"lastrunstatus" bson:"lastrunstatus"`
	LastRunTime       int64              `json:"lastruntime" bson:"lastruntime"`
	DisplayDefinition interface{}        `json:"displayDefinition" bson:"displayDefinition"`
}

type RegionResult struct {
	Result interface{} `json:"result" bson:"result"`
	Region string      `json:"region" bson:"region"`
}

type ResultMetaData struct {
	Cost            string                 `json:"currentcost" bson:"currentcost"`
	Recommendations []ResultRecommendation `json:"recommendations" bson:"recommendations"`
}

type ResultRecommendation struct {
	Recommendation          string `json:"recommendation" bson:"recommendation"`
	Price                   string `json:"price" bson:"price"`
	EstimatedCostSavings    string `json:"estimatedCostSaving" bson:"estimatedCostSaving"`
	EstimatedMonthlySavings string `json:"estimatedMonthlySaving" bson:"estimatedMonthlySaving"`
}

type ResourceCost struct {
	MinPrice float64
	MaxPrice float64
	Unit     string
	Currency string
}

// ******************* TO BE used in case needed to parse the policy json
type CSPolicyDef struct {
	Policies []CSPolicy `json:"policies" bson:"policies"`
}

type CSPolicy struct {
	Name     string `json:"name" bson:"name"`
	Resource string `json:"resource" bson:"resource"`
}

//******************* TO BE used in case needed to parse the policy json

type PipeLineNotification struct {
	EmailAddresses []string `json:"emailAddresses"`
	SlackUrls      []string `json:"slackUrls"`
	WebhookUrls    []string `json:"webhookUrls"`
}

type PipeLine struct {
	SweepAccountID   string               `json:"sweepaccountid" bson:"sweepaccountid"`
	CloudAccountID   string               `json:"cloudaccountid" bson:"cloudaccountid"`
	PipeLineID       primitive.ObjectID   `json:"piplineid" bson:"_id,omitempty"`
	PipeLineName     string               `json:"piplinename" bson:"piplinename"`
	Description      string               `json:"description" bson:"description"`
	Policies         []string             `json:"policies" bson:"policyid"`
	Schedule         Schedule             `json:"schedule" bson:"schedule"`
	Enabled          bool                 `json:"enabled" bson:"enabled"`
	Default          bool                 `json:"default" bson:"default"`
	RunStatus        RunStatus            `json:"status" bson:"status"`
	LastRunTime      int64                `json:"lastruntime" bson:"lastruntime"`
	ExecutionRegions []string             `json:"execregions" bson:"execregions"`
	Notification     PipeLineNotification `json:"notifications" bson:"notifications"`
}

type RunStatus int

const (
	RUNNING   RunStatus = 0
	COMPLETED RunStatus = 1
	FAILED    RunStatus = 2
	UNKNOWN   RunStatus = 3
)

type Schedule struct {
	Minute     string `json:"minute" bson:"minute"`
	Hour       string `json:"hour" bson:"hour"`
	DayOfMonth string `json:"dayofmonth" bson:"dayofmonth"`
	Month      string `json:"month" bson:"month"`
	DayOfWeek  string `json:"dayofweek" bson:"dayofweek"`
}

type Response200 struct {
	Status     string `json:"status"`
	StatusCode int    `json:"responsecode"`
}

type Response201 struct {
	Status     string `json:"status"`
	StatusCode int    `json:"responsecode"`
}

type Response400 struct {
	Error      string `json:"error"`
	StatusCode int    `json:"responsecode"`
}

type Response409 struct {
	Error      string `json:"error"`
	StatusCode int    `json:"responsecode"`
}

type Response404 struct {
	Error      string `json:"error"`
	StatusCode int    `json:"responsecode"`
}

type Response500 struct {
	Error      string `json:"internalerror"`
	StatusCode int    `json:"responsecode"`
}

type MongoIDQuery struct {
	ObjectID primitive.ObjectID `bson:"_id"`
}

// Result Display Definitions
func GetAWSInstanceDisplayDefinition() string {
	return `{"displayOrder": ["instanceId","instanceType","platformDetails","region","availabilityZone","stateName"],"displayName": {"instanceId": "Instance Id", "instanceType": "Instance Type", "platformDetails":"Platform", "region":"Region","availabilityZone":"Availability Zone", "stateName":"Instance State"}}`
}

func GetAWSVolumeDisplayDefinition() string {
	return `{"displayOrder": ["volumeId","volumeType","size","snapshotId","region","availabilityZone","state","attachments"],"displayName": {"volumeId": "Volume Id", "volumeType": "Volume Type", "size":"Size","snapshotId":"Snapshot Id", "region":"Region","availabilityZone":"Availability Zone", "state":"State","attachments":"Attached"}}`
}

func GetAWSEIPDisplayDefinition() string {
	return `{"displayOrder": ["allocationId","publicIp","publicIpv4pool","domain","networkBorderGroup"],"displayName": {"allocationId": "Allocation ID","publicIp": "Public IP", "publicIpv4pool": "IPv4 Pool", "domain":"Domain", "region":"Region","networkBorderGroup":"Network Border Group"}}`
}

func GetAWSSnapshotDisplayDefinition() string {
	return `{"displayOrder": ["snapshotId","volumeId","volumeSize","description","state","storageTier"],"displayName": {"snapshotId": "Snapshot Id", "volumeId": "Volume Id", "volumeSize":"Size", "description":"Description","state":"State","storageTier":"Storage Tier"}}`
}
