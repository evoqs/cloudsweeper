package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const AWS = "aws"

type CloudAccountData struct {
	AccountID      string             `json:"accountid" bson:"accountid"`
	CloudAccountID primitive.ObjectID `json:"cloudaccountid,omitempty" bson:"_id,omitempty"`
	AccountType    string             `json:"accounttype" bson:"accounttype"`
	Description    string             `json:"description" bson:"description"`
	AwsCredentials AwsCredentials     `json:"awscredentials" bson:"awscredentials"`
}

type AwsCredentials struct {
	//Region          string `json:"aws_region" bson:"aws_region"`
	AccessKeyID     string `json:"aws_access_key_id" bson:"aws_access_key_id"`
	SecretAccessKey string `json:"aws_secret_access_key" bson:"aws_secret_access_key"`
}

type Policy struct {
	PolicyName       string             `json:"policyname" bson:"policyname"`
	PolicyID         primitive.ObjectID `json:"policyid" bson:"_id,omitempty"`
	AccountID        string             `json:"accountid" bson:"accountid"`
	PolicyType       string             `json:"policytype" bson:"policytype"`
	PolicyDefinition string             `json:"policydefinition" bson:"policydefinition"`
	ExecutionRegions []string           `json:"execregions" bson:"execregions"`
}

type DefaultPolicy struct {
	PolicyName       string             `json:"policyname" bson:"policyname"`
	PolicyID         primitive.ObjectID `json:"policyid" bson:"_id,omitempty"`
	PolicyDefinition string             `json:"policydefinition" bson:"policydefinition"`
}

type PolicyResult struct {
	PolicyResultID primitive.ObjectID `json:"policyresultid" bson:"_id,omitempty"`
	PolicyID       string             `json:"policyid" bson:"policyid"`
	AccountID      string             `json:"accountid" bson:"accountid"`
	Resource       string             `json:"resource" bson:"resource"`
	Resultlist     []RegionResult     `json:"resultlist" bson:"resultlist"`
	LastRunStatus  string             `json:"lastrunstatus" bson:"lastrunstatus"`
}

type RegionResult struct {
	Result string `json:"result" bson:"result"`
	Region string `json:"region" bson:"region"`
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

type PipeLine struct {
	AccountID      string             `json:"accountid" bson:"accountid"`
	CloudAccountID string             `json:"cloudaccountid" bson:"cloudaccountid"`
	PipeLineID     primitive.ObjectID `json:"piplineid" bson:"_id,omitempty"`
	PipeLineName   string             `json:"piplinename" bson:"piplinename"`
	Policies       []string           `json:"policies" bson:"policyid"`
	Schedule       Schedule           `json:"schedule" bson:"schedule"`
	Enabled        bool               `json:"enabled" bson:"enabled"`
	RunStatus      RunStatus          `json:"status" bson:"status"`
	LastRunTime    string             `json:"lastruntime" bson:"lastruntime"`
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
