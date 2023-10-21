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

func (opr *AccountOperator) GetAllAccounts(accountquery string) ([]model.CloudAccountData, error) {

	var results []model.CloudAccountData

	cursor, err := opr.dbM.QueryRecord(cloudaccountTable, accountquery)

	fmt.Println(err)
	if err = cursor.All(context.TODO(), &results); err != nil {
		fmt.Println(err)
	}
	if results != nil {
		fmt.Println("Length " + strconv.Itoa(len(results)))
	}
	return results, err
}

func (opr *AccountOperator) GetCloudAccount(cloudaccountid string) ([]model.CloudAccountData, error) {

	var results []model.CloudAccountData

	cursor, err := opr.dbM.QueryRecordWithObjectID(cloudaccountTable, cloudaccountid)

	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	if results != nil {
		fmt.Println("Length " + strconv.Itoa(len(results)))
	}
	return results, err
}

func (opr *AccountOperator) AddCloudAccount(acc model.CloudAccountData) (string, error) {

	id, err := opr.dbM.InsertRecord(cloudaccountTable, acc)
	return id, err
}

func (opr *AccountOperator) UpdateCloudAccount(acc model.CloudAccountData) (int64, error) {

	objectId := acc.CloudAccountID
	result, err := opr.dbM.UpdateRecordWithObjectId(cloudaccountTable, objectId.Hex(), acc)
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, err

}

func (opr *AccountOperator) DeleteAllCloudAccounts(query string) (int64, error) {

	result, err := opr.dbM.DeleteMultipleRecord(cloudaccountTable, query)
	fmt.Printf("Delete Count %d,", result.DeletedCount)
	fmt.Println(err)
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, err
}

func (opr *AccountOperator) DeleteCloudAccount(objectid string) (int64, error) {

	result, err := opr.dbM.DeleteOneRecordWithObjectID(cloudaccountTable, objectid)

	return result.DeletedCount, err
}
