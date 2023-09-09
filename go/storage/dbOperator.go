package storage

import (
	log "cloudsweep/logging"
	"cloudsweep/model"
	"context"
	"fmt"
	"strconv"
)

const policyTable = "policies"
const pipelineTable = "pipelines"
const accountTable = "account"
const CounterTable = "idcounter"
const counterTableId = "100"

// *********************************************************Db Operations Cloud Account ***********************************************************************
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

func DeleteCloudAccount(dbM DBManger, dbName string, objectid string) (int64, error) {

	result, err := dbM.DeleteOneRecordWithObjectID(dbName, accountTable, objectid)

	return result.DeletedCount, err
}

//*********************************************************Db Operations policies ***********************************************************************

func AddPolicy(dbM DBManger, dbName string, policy model.CSPolicy) (string, error) {

	id, err := dbM.InsertRecord(dbName, policyTable, policy)
	return id, err
}

func UpdatePolicy(dbM DBManger, dbName string, policy model.CSPolicy) (int64, error) {

	objectId := policy.PolicyID
	result, err := dbM.UpdateRecordWithObjectId(dbName, policyTable, objectId.Hex(), policy)
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, err
}

func GetPolicyDetails(dbM DBManger, dbName string, policyid string) ([]model.CSPolicy, error) {

	var results []model.CSPolicy

	cursor, err := dbM.QueryRecordWithObjectID(dbName, policyTable, policyid)

	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	if results != nil {
		fmt.Println("Length " + strconv.Itoa(len(results)))
	}
	return results, err
}

func DeleteCustodianPolicy(dbM DBManger, dbName string, objectid string) (int64, error) {

	result, err := dbM.DeleteOneRecordWithObjectID(dbName, policyTable, objectid)
	return result.DeletedCount, err
}

// *********************************************************Db Operations pipelines ***********************************************************************
func AddPipeline(dbM DBManger, dbName string, pipeline model.PipeLine) (string, error) {
	id, err := dbM.InsertRecord(dbName, pipelineTable, pipeline)
	return id, err
}

func UpdatePipeline(dbM DBManger, dbName string, pipeline model.PipeLine) (int64, error) {
	objectId := pipeline.PipeLineID
	result, err := dbM.UpdateRecordWithObjectId(dbName, pipelineTable, objectId.Hex(), pipeline)
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, err
}

func GetPipeline(dbM DBManger, dbName string, pipelineId string) ([]model.PipeLine, error) {
	var results []model.PipeLine
	cursor, err := dbM.QueryRecordWithObjectID(dbName, pipelineTable, pipelineId)

	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	if results != nil {
		fmt.Println("Length " + strconv.Itoa(len(results)))
	}
	return results, err
}

func GetAllPipelines(dbM DBManger, dbName string) ([]model.PipeLine, error) {
	var results []model.PipeLine
	cursor, err := dbM.QueryAllRecords(dbName, pipelineTable)
	if err = cursor.All(context.TODO(), &results); err != nil {
		log.NewDefaultLogger().Infof("Problem in getting all the pipelines. Reason: " + err.Error())
	}
	return results, err
}

func DeletePipeline(dbM DBManger, dbName string, objectid string) (int64, error) {
	result, err := dbM.DeleteOneRecordWithObjectID(dbName, pipelineTable, objectid)
	return result.DeletedCount, err
}
