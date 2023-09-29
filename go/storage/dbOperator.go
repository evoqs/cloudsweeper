package storage

import "cloudsweep/utils"

const policyTable = "policies"
const policyResultTable = "policyrunresult"
const accountTable = "account"
const pipelineTable = "pipeline"

var operatorRepo = make(map[string]*DbOperators)

type DbOperators struct {
	PipeLineOperator PipeLineOperator
	AccountOperator  AccountOperator
	PolicyOperator   PolicyOperator
	CostOperator     CostOperator
}

func MakeDBOperators(dbm *DBManger) *DbOperators {
	var Db_Operators DbOperators

	var PipeLine_Operator PipeLineOperator
	var Policy_Operator PolicyOperator
	var Account_Operator AccountOperator
	costOperator := GetDefaultCostOperator()

	PipeLine_Operator.dbM = *dbm
	Policy_Operator.dbM = *dbm
	Account_Operator.dbM = *dbm
	costOperator.dbM = *dbm

	Db_Operators.AccountOperator = Account_Operator
	Db_Operators.PipeLineOperator = PipeLine_Operator
	Db_Operators.PolicyOperator = Policy_Operator
	Db_Operators.CostOperator = *costOperator
	operatorRepo[dbm.dbName] = &Db_Operators
	return (&Db_Operators)
}

func GetDBOperators(dbName string) *DbOperators {
	return operatorRepo[dbName]
}

func GetDefaultDBOperators() *DbOperators {
	return operatorRepo[utils.GetConfig().Database.Name]
}

// *********************************************************Db Operations Cloud Account ***********************************************************************
/*
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
*/
//*********************************************************Db Operations policies ***********************************************************************
/*
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
*/
