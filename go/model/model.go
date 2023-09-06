package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const AWS = "aws"

type AccountData struct {
	AccountID      string             `json:"accountid" bson:"accountid"`
	CloudAccountID primitive.ObjectID `json:"cloudaccountid,omitempty" bson:"_id,omitempty"`
	AccountType    string             `json:"accounttype" bson:"accounttype"`
	Description    string             `json:"description" bson:"description"`
	AwsCredentials AwsCredentials     `json:"awscredentials" bson:"awscredentials"`
}

type AwsCredentials struct {
	Region          string `json:"aws_region" bson:"aws_region"`
	AccessKeyID     string `json:"aws_access_key_id" bson:"aws_access_key_id"`
	SecretAccessKey string `json:"aws_secret_access_key" bson:"aws_secret_access_key"`
}

type Policy struct {
	PolicyName       string             `json:"policyname" bson:"policyname"`
	PolicyID         primitive.ObjectID `json:"policyid" bson:"_id,omitempty"`
	AccountID        string             `json:"accountid" bson:"accountid"`
	CloudAccountID   string             `json:"cloudaccountid" bson:"cloudaccountid"`
	PolicyDefinition string             `json:"policydefinition" bson:"policydefinition"`
}

type PolicyResult struct {
	PolicyResultID primitive.ObjectID `json:"policyresultid" bson:"_id,omitempty"`
	PolicyID       string             `json:"policyid" bson:"policyid"`
	Resource       string             `json:"resource" bson:"resource"`
	Result         string             `json:"result" bson:"result"`
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
	PolicyID       []string           `json:"policyid" bson:"policyid"`
	Schedule       Schedule           `json:"schedule" bson:"schedule"`
	Enabled        bool               `json:"enabled" bson:"enabled"`
	RunStatus      RunStatus          `json:"status" bson:"status"`
}

type RunStatus int

const (
	RUNNING   RunStatus = 0
	COMPLETED RunStatus = 1
	FAILED    RunStatus = 2
	UNKNOWN   RunStatus = 3
)

type Schedule struct {
	Hourly int            `json:"hourly" bson:"hourly"`
	Daily  ScheduleDaily  `json:"daily" bson:"daily"`
	Weekly ScheduleWeekly `json:"weekly" bson:"weekly"`
}

type ScheduleDaily struct {
	Intreval int       `json:"intreval" bson:"intreval"`
	Time     time.Time `json:"time" bson:"time"`
}

type ScheduleWeekly struct {
	DaysOfWeek string    `json:"daysofweek" bson:"daysofweek"`
	Time       time.Time `json:"time" bson:"time"`
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
