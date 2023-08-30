package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AccountData struct {
	AccountID      string             `json:"accountid" bson:"accountid"`
	CloudAccountID primitive.ObjectID `json:"cloudaccountid,omitempty" bson:"_id,omitempty"`
	AccountType    string             `json:"accounttype" bson:"accounttype"`
	Description    string             `json:"description" bson:"description"`
	AwsCredentials AwsCredentials     `json:"awscredentials" bson:"awscredentials"`
}

type AwsCredentials struct {
	AccessKeyID     string `json:"aws_access_key_id" bson:"aws_access_key_id"`
	SecretAccessKey string `json:"aws_secret_access_key" bson:"aws_secret_access_key"`
}

type CSPolicy struct {
	PolicyName       string             `json:"policyname" bson:"policyname"`
	PolicyID         primitive.ObjectID `json:"policyid" bson:"_id,omitempty"`
	AccountID        string             `json:"accountid" bson:"accountid"`
	CloudAccountID   string             `json:"cloudaccountid" bson:"cloudaccountid"`
	PolicyDefinition string             `json:"policydefinition" bson:"policydefinition"`
}

type PipeLine struct {
	AccountID      string    `json:"accountid" bson:"accountid"`
	CloudAccountID string    `json:"cloudaccountid" bson:"cloudaccountid"`
	PipeLineID     string    `json:"piplineid" bson:"piplineid"`
	PipeLineName   string    `json:"piplinename" bson:"piplinename"`
	PolicyID       []int     `json:"policyid" bson:"policyid"`
	Schedule       Schedule  `json:"schedule" bson:"schedule"`
	Enabled        bool      `json:"enabled" bson:"enabled"`
	RunStatus      RunStatus `json:"status" bson:"status"`
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
