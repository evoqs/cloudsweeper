package model

import "time"

type AccountData struct {
	AccountID      int            `json:"accountid" bson:"accountid"`
	CloudAccountID int            `json:"cloudaccountid" bson:"cloudaccountid"`
	AccountType    string         `json:"accounttype" bson:"accounttype"`
	Description    string         `json:"description" bson:"description"`
	AwsCredentials AwsCredentials `json:"awscredentials" bson:"awscredentials"`
}

type AwsCredentials struct {
	AccessKeyID     string `json:"aws_access_key_id" bson:"aws_access_key_id"`
	SecretAccessKey string `json:"aws_secret_access_key" bson:"aws_secret_access_key"`
}

func (w *AccountData) SetData(id int, typ string, des string) {
	w.AccountID = id
	w.AccountType = typ
	w.Description = des
}

type CSPolicy struct {
	PolicyName       string `json:"policyname" bson:"policyname"`
	PolicyID         int    `json:"policyid" bson:"policyid"`
	AccountID        int    `json:"accountid" bson:"accountid"`
	CloudAccountID   int    `json:"cloudaccountid" bson:"cloudaccountid"`
	PolicyDefinition string `json:"policydefinition" bson:"policydefinition"`
}

type PipeLine struct {
	AccountID      int       `json:"accountid" bson:"accountid"`
	CloudAccountID int       `json:"cloudaccountid" bson:"cloudaccountid"`
	PipeLineID     int       `json:"piplineid" bson:"piplineid"`
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

type IdCounter struct {
	CounterID          int `json:"counterid" bson:"counterid"`
	NextAccountID      int `json:"nextaccountid" bson:"nextaccountid"`
	NextPolicyID       int `json:"nextpolicyid" bson:"nextpolicyid"`
	NextCloudAccountID int `json:"nextcloudaccountid" bson:"nextcloudaccountid"`
}
