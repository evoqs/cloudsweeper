package storage

import (
	"cloudsweep/model"
	"context"
	"fmt"
	"strconv"
)

type PolicyOperator struct {
	dbM DBManger
}

func (opr *PolicyOperator) AddPolicy(policy model.Policy) (string, error) {

	id, err := opr.dbM.InsertRecord(policyTable, policy)
	return id, err
}

func (opr *PolicyOperator) UpdatePolicy(policy model.Policy) (int64, error) {

	objectId := policy.PolicyID
	result, err := opr.dbM.UpdateRecordWithObjectId(policyTable, objectId.Hex(), policy)
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, err
}

func (opr *PolicyOperator) GetPolicyDetails(policyid string) ([]model.Policy, error) {

	var results []model.Policy

	cursor, err := opr.dbM.QueryRecordWithObjectID(policyTable, policyid)

	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	if results != nil {
		fmt.Println("Length " + strconv.Itoa(len(results)))
	}
	return results, err
}

func (opr *PolicyOperator) QueryPolicyDetails(query string) ([]model.Policy, error) {

	var results []model.Policy

	cursor, err := opr.dbM.QueryRecord(policyTable, query)

	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	if results != nil {
		fmt.Println("Length " + strconv.Itoa(len(results)))
	}
	return results, err
}

func (opr *PolicyOperator) DeleteCustodianPolicy(objectid string) (int64, error) {

	result, err := opr.dbM.DeleteOneRecordWithObjectID(policyTable, objectid)
	return result.DeletedCount, err
}

// Policy result Operations
func (opr *PolicyOperator) AddPolicyResult(policyresult model.PolicyResult) (string, error) {

	id, err := opr.dbM.InsertRecord(policyResultTable, policyresult)
	return id, err
}

func (opr *PolicyOperator) UpdatePolicyResult(policyresult model.PolicyResult) (int64, error) {

	objectId := policyresult.PolicyResultID
	result, err := opr.dbM.UpdateRecordWithObjectId(policyResultTable, objectId.Hex(), policyresult)
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, err
}

func (opr *PolicyOperator) GetPolicyResultDetails(query string) ([]model.PolicyResult, error) {

	var results []model.PolicyResult

	fmt.Println("Query: ", query)
	cursor, err := opr.dbM.QueryRecord(policyResultTable, query)

	/*if cursor == nil {
		fmt.Println("Empty result Set")
		return nil, nil
	}*/
	if err = cursor.All(context.TODO(), &results); err != nil {
		fmt.Println(err)
		return nil, err
	}
	if results != nil {
		fmt.Println("Length " + strconv.Itoa(len(results)))
	}
	return results, err
}

// Default Policy Operations
func (opr *PolicyOperator) AddDefaultPolicy(policy model.Policy) (string, error) {

	id, err := opr.dbM.InsertRecord(policyTable, policy)
	return id, err
}

func (opr *PolicyOperator) UpdateDefaultPolicy(policy model.Policy) (int64, error) {

	objectId := policy.PolicyID
	result, err := opr.dbM.UpdateRecordWithObjectId(policyTable, objectId.Hex(), policy)
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, err
}

func (opr *PolicyOperator) GetAllDefaultPolicyDetails() ([]model.Policy, error) {

	var results []model.Policy

	query := fmt.Sprintf(`{"isDefault": %t}`, true)
	cursor, err := opr.dbM.QueryRecord(policyTable, query)

	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	if results != nil {
		fmt.Println("Length " + strconv.Itoa(len(results)))
	}
	return results, err
}

func (opr *PolicyOperator) DeleteDefaultCustodianPolicy(objectid string) (int64, error) {

	result, err := opr.dbM.DeleteOneRecordWithObjectID(policyTable, objectid)
	return result.DeletedCount, err
}
