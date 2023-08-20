package model

type AccountData struct {
	AccountID      int    `json:"accountid" bson:"accountid"`
	CloudAccountID int    `json:"cloudaccountid" bson:"cloudaccountid"`
	AccountType    string `json:"accounttype" bson:"accounttype"`
	Description    string `json:"description" bson:"description"`
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

type IdCounter struct {
	CounterID          int `json:"counterid" bson:"counterid"`
	NextAccountID      int `json:"nextaccountid" bson:"nextaccountid"`
	NextPolicyID       int `json:"nextpolicyid" bson:"nextpolicyid"`
	NextCloudAccountID int `json:"nextcloudaccountid" bson:"nextcloudaccountid"`
}
