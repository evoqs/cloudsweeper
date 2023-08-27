package storage

import (
	"cloudsweep/model"
	"context"
	"fmt"
	"strconv"
)

const policyTable = "policies"
const accountTable = "account"
const CounterTable = "idcounter"
const counterTableId = "100"

func GetAllAccounts(dbM DBManger, dbName string, cloudaccountid string) ([]model.AccountData, error) {

	var results []model.AccountData

	cursor, err := dbM.QueryRecord(dbName, accountTable, cloudaccountid)

	fmt.Println(err)
	if err = cursor.All(context.TODO(), &results); err != nil {
		fmt.Println(err)
	}
	if results != nil {
		fmt.Println("Length " + strconv.Itoa(len(results)))
	}
	return results, err
}

func GetCloudAccount(dbM DBManger, dbName string, cloudaccountid string) ([]model.AccountData, error) {

	var results []model.AccountData

	cursor, err := dbM.QueryRecordWithObjectID(dbName, accountTable, cloudaccountid)

	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	if results != nil {
		fmt.Println("Length " + strconv.Itoa(len(results)))
	}
	return results, err
}

func AddCloudAccount(dbM DBManger, dbName string, acc model.AccountData) (string, error) {

	id, err := dbM.InsertRecord(dbName, accountTable, acc)
	return id, err
}

func UpdateCloudAccount(dbM DBManger, dbName string, acc model.AccountData) (int64, error) {

	objectId := acc.CloudAccountID
	result, err := dbM.UpdateRecordWithObjectId(dbName, accountTable, objectId.Hex(), acc)
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, err

}

func DeleteAllCloudAccounts(dbM DBManger, dbName string, query string) (int64, error) {

	result, err := dbM.DeleteMultipleRecord(dbName, accountTable, query)
	fmt.Printf("Delete Count %d,", result.DeletedCount)
	fmt.Println(err)
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, err
}

func DeleteCloudAccount(dbM DBManger, dbName string, query string) (int64, error) {

	result, err := dbM.DeleteOneRecordWithObjectID(dbName, accountTable, query)

	return result.DeletedCount, err
}
