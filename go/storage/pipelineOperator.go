package storage

import (
	"cloudsweep/model"
	"context"
	"fmt"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
)

type PipeLineOperator struct {
	dbM DBManger
}

func (opr *PipeLineOperator) AddPipeLine(pipeline model.PipeLine) (string, error) {

	id, err := opr.dbM.InsertRecord(pipelineTable, pipeline)
	return id, err
}

func (opr *PipeLineOperator) UpdatePipeLine(pipeline model.PipeLine) (int64, error) {

	objectId := pipeline.PipeLineID
	result, err := opr.dbM.UpdateRecordWithObjectId(pipelineTable, objectId.Hex(), pipeline)
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, err
}

func (opr *PipeLineOperator) GetPipeLineDetails(pipelineid string) ([]model.PipeLine, error) {

	var results []model.PipeLine

	cursor, err := opr.dbM.QueryRecordWithObjectID(pipelineTable, pipelineid)

	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	if results != nil {
		fmt.Println("Length " + strconv.Itoa(len(results)))
	}
	return results, err
}

// Fetches all pipelines belonging to an account
func (opr *PipeLineOperator) GetAccountPipeLines(query string) ([]model.PipeLine, error) {

	var results []model.PipeLine
	cursor, err := opr.dbM.QueryRecord(pipelineTable, query)

	fmt.Println(err)
	if err = cursor.All(context.TODO(), &results); err != nil {
		fmt.Println(err)
	}
	if results != nil {
		fmt.Println("Length " + strconv.Itoa(len(results)))
	}
	return results, err
}

func (opr *PipeLineOperator) GetAllPipeLines() ([]model.PipeLine, error) {
	var results []model.PipeLine
	cursor, err := opr.dbM.QueryAllRecords(pipelineTable)

	fmt.Println(err)
	if err = cursor.All(context.TODO(), &results); err != nil {
		fmt.Println(err)
	}
	if results != nil {
		fmt.Println("Length " + strconv.Itoa(len(results)))
	}
	return results, err
}

func (opr *PipeLineOperator) DeletePipeLine(objectid string) (int64, error) {

	result, err := opr.dbM.DeleteOneRecordWithObjectID(pipelineTable, objectid)
	return result.DeletedCount, err
}

func (opr *PipeLineOperator) GetPipelineResultDetails(query string) ([]bson.M, error) {

	//var results []model.PolicyResult

	fmt.Println("Query: ", query)
	cursor, err := opr.dbM.QueryRecord(policyResultTable, query)

	/*if cursor == nil {
		fmt.Println("Empty result Set")
		return nil, nil
	}*/

	var docs []bson.M
	if err = cursor.All(context.TODO(), &docs); err != nil {
		fmt.Println(err)
		return nil, err
	}

	/*for result.Next(ctx) {
	      var document bson.M
	      err = result.Decode(&document)
	      if err != nil {
	          log.Println(err)
	      }
	      docs = append(docs, document)
	  }
	  return docs*/

	fmt.Printf("%+v", docs)
	return docs, err
}
