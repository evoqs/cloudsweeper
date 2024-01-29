package storage

import "cloudsweep/config"

const policyTable = "policies"

// const defaultpolicyTable = "defaultpolicies"
const policyResultTable = "policyrunresult"
const cloudaccountTable = "cloudaccount"
const pipelineTable = "pipeline"

var operatorRepo = make(map[string]*DbOperators)

type DbOperators struct {
	PipeLineOperator       PipeLineOperator
	AccountOperator        AccountOperator
	PolicyOperator         PolicyOperator
	CostOperator           CostOperator
	RecommendationOperator RecommendationOperator
}

func MakeDBOperators(dbm *DBManger) *DbOperators {
	var Db_Operators DbOperators

	var PipeLine_Operator PipeLineOperator
	var Policy_Operator PolicyOperator
	var Account_Operator AccountOperator
	costOperator := GetDefaultCostOperator()
	recommendationOperator := GetDefaultRecommendationOperator()

	PipeLine_Operator.dbM = *dbm
	Policy_Operator.dbM = *dbm
	Account_Operator.dbM = *dbm
	costOperator.dbM = *dbm
	recommendationOperator.dbM = *dbm

	Db_Operators.AccountOperator = Account_Operator
	Db_Operators.PipeLineOperator = PipeLine_Operator
	Db_Operators.PolicyOperator = Policy_Operator
	Db_Operators.CostOperator = *costOperator
	Db_Operators.RecommendationOperator = *recommendationOperator
	operatorRepo[dbm.dbName] = &Db_Operators
	return (&Db_Operators)
}

func GetDBOperators(dbName string) *DbOperators {
	return operatorRepo[dbName]
}

func GetDefaultDBOperators() *DbOperators {
	return operatorRepo[config.GetConfig().Database.Name]
}
