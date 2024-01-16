package storage

import (
	logger "cloudsweep/logging"
	aws_model "cloudsweep/model/aws"
	"context"
	"reflect"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const defaultTableName = "cloudresourcecost"

type CostOperator struct {
	dbM       DBManger
	tableName string
}

func GetDefaultCostOperator() *CostOperator {
	// TODO: DBManager also should be set here, once DBManager.go is capable to create the object independently
	return &CostOperator{
		tableName: defaultTableName,
	}
}

func (opr *CostOperator) AddResourceCost(resource interface{}) (string, error) {
	id, err := opr.dbM.InsertRecord(opr.tableName, resource)
	return id, err
}

func (opr *CostOperator) UpdateResourceCost(resource interface{}) (int64, error) {
	var Id primitive.ObjectID
	switch r := resource.(type) {
	case *aws_model.AwsResourceCost[any]:
		Id = r.Id
	default:
	}
	result, err := opr.dbM.UpdateRecordWithObjectId(opr.tableName, Id.Hex(), resource)
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, err
}

func (opr *CostOperator) GetResourceCostDetails(resourceId string, targetType interface{}) ([]interface{}, error) {
	var results []interface{}
	cursor, err := opr.dbM.QueryRecordWithObjectID(opr.tableName, resourceId)
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	if results != nil {
		logger.NewDefaultLogger().Debugf("Length of ResourceCostDetails" + strconv.Itoa(len(results)))
	}
	return results, err
}

func (opr *CostOperator) RunQuery(query string) ([]interface{}, error) {
	var results []interface{}
	cursor, err := opr.dbM.QueryRecord(opr.tableName, query)
	if err = cursor.All(context.TODO(), &results); err != nil {
		logger.NewDefaultLogger().Errorf("Error: %v", err)
	}
	if results != nil {
		logger.NewDefaultLogger().Debugf("Length of Records" + strconv.Itoa(len(results)))
	}
	return results, err
}

func (opr *CostOperator) GetQueryResult(query string, results interface{}) error {
	logger.NewDefaultLogger().Debugf("Querying DB collection %v", opr.tableName)
	cursor, err := opr.dbM.QueryRecord(opr.tableName, query)
	if err != nil {
		return err
	}

	sliceValue := reflect.ValueOf(results).Elem()
	targetType := sliceValue.Type().Elem()
	for cursor.Next(context.TODO()) {
		// Create a new instance of the target type
		target := reflect.New(targetType).Interface()
		if err := cursor.Decode(target); err != nil {
			return err
		}
		// Append the decoded result to the slice
		sliceValue.Set(reflect.Append(sliceValue, reflect.ValueOf(target).Elem()))
	}
	return nil
}

func (opr *CostOperator) GetQueryResultCount(query string) (int, error) {
	logger.NewDefaultLogger().Debugf("Querying DB collection %v", opr.tableName)
	cursor, err := opr.dbM.QueryRecord(opr.tableName, query)
	if err != nil {
		return 0, err
	}
	var count int
	for cursor.Next(context.TODO()) {
		count++
	}
	return count, nil
}

func (opr *CostOperator) DeleteResourceCost(resourdId string) (int64, error) {

	result, err := opr.dbM.DeleteOneRecordWithObjectID(opr.tableName, resourdId)
	return result.DeletedCount, err
}

func (opr *CostOperator) DeleteOldResourceCosts(thresholdTime time.Time) (*mongo.DeleteResult, error) {
	filterJSON := `{
		"timeStamp": {
			"$lt": ` + strconv.FormatInt(thresholdTime.Unix(), 10) + `
		}
	}`
	result, err := opr.dbM.DeleteMultipleRecord(opr.tableName, filterJSON)
	if err != nil {
		return nil, err
	}

	return result, nil
}
