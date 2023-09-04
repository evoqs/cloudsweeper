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

func (opr *PolicyOperator) AddPolicy(policy model.CSPolicy) (string, error) {

	id, err := opr.dbM.InsertRecord(policyTable, policy)
	return id, err
}

func (opr *PolicyOperator) UpdatePolicy(policy model.CSPolicy) (int64, error) {

	objectId := policy.PolicyID
	result, err := opr.dbM.UpdateRecordWithObjectId(policyTable, objectId.Hex(), policy)
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, err
}

func (opr *PolicyOperator) GetPolicyDetails(policyid string) ([]model.CSPolicy, error) {

	var results []model.CSPolicy

	cursor, err := opr.dbM.QueryRecordWithObjectID(policyTable, policyid)

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
