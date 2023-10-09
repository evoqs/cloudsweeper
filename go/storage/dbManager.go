package storage

import (
	"cloudsweep/model"
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type DBManger struct {
	url         string
	dbName      string
	mongoClinet *mongo.Client
}

func GetDBManager() *DBManger {
	return new(DBManger)
}

func (dbm *DBManger) SetDbUrl(url string) {
	dbm.url = url
}

func (dbm *DBManger) SetDatabase(dbName string) {
	dbm.dbName = dbName
}

func (dbm *DBManger) Connect() (*mongo.Client, error) {
	if dbm.url == "" {
		return nil, errors.New("URL needs to be set before attempting to connect")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbm.url))
	dbm.mongoClinet = client
	return client, err
}

func (dbm DBManger) CheckConnection() error {

	if dbm.mongoClinet == nil {
		return errors.New("Empty Client, Open a connection using Connect() method")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := dbm.mongoClinet.Ping(ctx, readpref.Primary())
	return err
}

func (dbm DBManger) InsertRecord(collection string, rec interface{}) (string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	record, err := dbm.mongoClinet.Database(dbm.dbName).Collection(collection).InsertOne(ctx, rec)

	id := record.InsertedID.(primitive.ObjectID).Hex()
	return id, err
}

func (dbm DBManger) Disconnect() error {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := dbm.mongoClinet.Disconnect(ctx)
	return err
}

func (dbm DBManger) QueryRecord(collection string, query string) (*mongo.Cursor, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var bquery interface{}
	err := bson.UnmarshalExtJSON([]byte(query), true, &bquery)
	if err != nil {
		fmt.Println("Error unmarshelling", err.Error())
	}
	cursor, err := dbm.mongoClinet.Database(dbm.dbName).Collection(collection).Find(ctx, &bquery)

	return cursor, err

}

func (dbm DBManger) QueryAllRecords(collection string) (*mongo.Cursor, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var bquery interface{}
	query := `{}`
	err := bson.UnmarshalExtJSON([]byte(query), true, &bquery)
	if err != nil {
		fmt.Println("Error unmarshelling", err.Error())
	}
	cursor, err := dbm.mongoClinet.Database(dbm.dbName).Collection(collection).Find(ctx, &bquery)
	return cursor, err
}

func (dbm DBManger) QueryRecordsWithPagination(collection string, page, pageSize int, sortField string, query bson.M) (*mongo.Cursor, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Define options for sorting, limiting, and querying
	options := options.Find().SetSort(bson.D{{Key: sortField, Value: 1}}).
		SetSkip(int64((page - 1) * pageSize)).SetLimit(int64(pageSize))

	cursor, err := dbm.mongoClinet.Database(dbm.dbName).Collection(collection).Find(ctx, query, options)
	return cursor, err
}

func (dbm DBManger) QueryRecordWithObjectID(collection string, mongoObjectid string) (*mongo.Cursor, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectId, err := primitive.ObjectIDFromHex(mongoObjectid)
	if err != nil {
		fmt.Println("Invalid id")
	}

	bquery := bson.M{"_id": objectId}
	cursor, err := dbm.mongoClinet.Database(dbm.dbName).Collection(collection).Find(ctx, &bquery)

	return cursor, err

}

func (dbm DBManger) QueryOneRecord(collection string, query string) (*mongo.SingleResult, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var bquery interface{}
	err := bson.UnmarshalExtJSON([]byte(query), true, &bquery)

	if err != nil {
		// handle error
	}
	cursor := dbm.mongoClinet.Database(dbm.dbName).Collection(collection).FindOne(ctx, &bquery, options.FindOne())

	return cursor, err

}

func (dbm DBManger) UpdateRecord(collection string, query string, rec interface{}) (*mongo.UpdateResult, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var bquery interface{}
	err := bson.UnmarshalExtJSON([]byte(query), true, &bquery)
	if err != nil {
		// handle error
	}
	result, err := dbm.mongoClinet.Database(dbm.dbName).Collection(collection).ReplaceOne(ctx, &bquery, &rec)
	return result, err
}

func (dbm DBManger) UpdateRecordWithObjectId(collection string, mongoObjectid string, rec interface{}) (*mongo.UpdateResult, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//fmt.Println(objectId)
	objectId, err := primitive.ObjectIDFromHex(mongoObjectid)
	bquery := bson.M{"_id": objectId}
	result, err := dbm.mongoClinet.Database(dbm.dbName).Collection(collection).ReplaceOne(ctx, &bquery, &rec)
	return result, err
}

func (dbm DBManger) DeleteOneRecord(collection string, query string) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var bquery interface{}
	err := bson.UnmarshalExtJSON([]byte(query), true, &bquery)
	if err != nil {
		// handle error
	}
	_, err = dbm.mongoClinet.Database(dbm.dbName).Collection(collection).DeleteOne(ctx, &bquery)

	return err

}

func (dbm DBManger) DeleteOneRecordWithObjectID(collection string, mongoObjectid string) (*mongo.DeleteResult, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	objectId, err := primitive.ObjectIDFromHex(mongoObjectid)
	bquery := bson.M{"_id": objectId}
	result, err := dbm.mongoClinet.Database(dbm.dbName).Collection(collection).DeleteOne(ctx, &bquery)

	return result, err

}

func (dbm DBManger) DeleteMultipleRecord(collection string, query string) (*mongo.DeleteResult, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var bquery interface{}
	err := bson.UnmarshalExtJSON([]byte(query), true, &bquery)
	if err != nil {
		// handle error
	}
	result, err := dbm.mongoClinet.Database(dbm.dbName).Collection(collection).DeleteMany(ctx, &bquery)

	return result, err

}

func getMongoObjectQuery(hexobjId string) (model.MongoIDQuery, error) {
	objID, err := primitive.ObjectIDFromHex(hexobjId)
	return model.MongoIDQuery{ObjectID: objID}, err
}
