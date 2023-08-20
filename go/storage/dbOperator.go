package storage

import (
	"context"
	"fmt"
	"strconv"
	"testmod/model"
)

const policyTable = "policies"
const accountTable = "account"
const CounterTable = "idcounter"
const counterTableId = "100"

func GetAccounts(dbM DBManger, dbName string, query string) []model.AccountData {

	var results []model.AccountData

	cursor, err := dbM.QueryRecord(dbName, accountTable, query)

	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	if results != nil {
		fmt.Println("Length " + strconv.Itoa(len(results)))
	}
	return results
}

func PutAccounts(dbM DBManger, dbName string, acc model.AccountData) error {

	err := dbM.InsertRecord(dbName, accountTable, acc)
	return err
}

// For internal Use only
func GetIdCounter(dbM DBManger, dbName string) model.IdCounter {

	var result model.IdCounter
	query := `{"counterid" : ` + counterTableId + `}`

	cursor, err := dbM.QueryOneRecord(dbName, CounterTable, query)

	if err = cursor.Decode(&result); err != nil {
		//panic(err)

	}
	return result
}

func InitializeCounter(dbM DBManger, dbName string) error {
	var counter model.IdCounter
	counter.CounterID = 100
	counter.NextAccountID = 1000
	counter.NextCloudAccountID = 2000
	counter.NextPolicyID = 30000

	err := dbM.InsertRecord(dbName, CounterTable, counter)
	return err
}

// For internal Use only
func UpdateIdCounter(dbM DBManger, dbName string, counter model.IdCounter) error {

	query := `{"counterid" : ` + counterTableId + `}`

	err := dbM.UpdateRecord(dbName, CounterTable, query, counter)

	return err
}
