package storage

import (
	"cloudsweep/model"
	"context"
	"fmt"
	"strconv"
)

type AccountOperator struct {
	dbM DBManger
}

func (opr *AccountOperator) GetAllAccounts(accountquery string) ([]model.AccountData, error) {

	var results []model.AccountData

	cursor, err := opr.dbM.QueryRecord(accountTable, accountquery)

	fmt.Println(err)
	if err = cursor.All(context.TODO(), &results); err != nil {
		fmt.Println(err)
	}
	if results != nil {
		fmt.Println("Length " + strconv.Itoa(len(results)))
	}
	return results, err
}

func (opr *AccountOperator) GetCloudAccount(cloudaccountid string) ([]model.AccountData, error) {

	var results []model.AccountData

	cursor, err := opr.dbM.QueryRecordWithObjectID(accountTable, cloudaccountid)

	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	if results != nil {
		fmt.Println("Length " + strconv.Itoa(len(results)))
	}
	return results, err
}

func (opr *AccountOperator) AddCloudAccount(acc model.AccountData) (string, error) {

	id, err := opr.dbM.InsertRecord(accountTable, acc)
	return id, err
}

func (opr *AccountOperator) UpdateCloudAccount(acc model.AccountData) (int64, error) {

	objectId := acc.CloudAccountID
	result, err := opr.dbM.UpdateRecordWithObjectId(accountTable, objectId.Hex(), acc)
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, err

}

func (opr *AccountOperator) DeleteAllCloudAccounts(query string) (int64, error) {

	result, err := opr.dbM.DeleteMultipleRecord(accountTable, query)
	fmt.Printf("Delete Count %d,", result.DeletedCount)
	fmt.Println(err)
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, err
}

func (opr *AccountOperator) DeleteCloudAccount(objectid string) (int64, error) {

	result, err := opr.dbM.DeleteOneRecordWithObjectID(accountTable, objectid)

	return result.DeletedCount, err
}
