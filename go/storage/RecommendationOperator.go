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

type RecommendationOperator struct {
	dbM       DBManger
	tableName string
}

func GetDefaultRecommendationOperator() *RecommendationOperator {
	// TODO: DBManager also should be set here, once DBManager.go is capable to create the object independently
	return &RecommendationOperator{
		tableName: "recommendations",
	}
}

func (opr *RecommendationOperator) AddRecommendation(resource interface{}) (string, error) {
	id, err := opr.dbM.InsertRecord(opr.tableName, resource)
	return id, err
}

func (opr *RecommendationOperator) UpdateRecommendation(resource interface{}) (int64, error) {
	var Id primitive.ObjectID
	switch r := resource.(type) {
	case *aws_model.Recommendation[aws_model.InstanceDetails]:
		Id = r.Id
	case *aws_model.Recommendation[aws_model.EBSVolumeDetails]:
		Id = r.Id
	default:
	}

	result, err := opr.dbM.UpdateRecordWithObjectId(opr.tableName, Id.Hex(), resource)
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, err
}

func (opr *RecommendationOperator) GetRecommendation(resourceId string, targetType interface{}) ([]interface{}, error) {
	var results []interface{}
	cursor, err := opr.dbM.QueryRecordWithObjectID(opr.tableName, resourceId)
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	if results != nil {
		logger.NewDefaultLogger().Debugf("Length of Recommendations" + strconv.Itoa(len(results)))
	}
	return results, err
}

func (opr *RecommendationOperator) RunQuery(query string) ([]interface{}, error) {
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

func (opr *RecommendationOperator) GetQueryResult(query string, results interface{}) error {
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

func (opr *RecommendationOperator) GetQueryResultCount(query string) (int, error) {
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

func (opr *RecommendationOperator) DeleteRecommendation(resourdId string) (int64, error) {

	result, err := opr.dbM.DeleteOneRecordWithObjectID(opr.tableName, resourdId)
	return result.DeletedCount, err
}

func (opr *RecommendationOperator) DeleteOldRecommendations(thresholdTime time.Time) (*mongo.DeleteResult, error) {
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
