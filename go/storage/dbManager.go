package storage

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type DBManger struct {
	url         string
	mongoClinet *mongo.Client
}

func GetDBManager() *DBManger {
	return new(DBManger)
}
func (dbm *DBManger) SetDbUrl(url string) {
	dbm.url = url
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

func (dbm DBManger) InsertRecord(dbname string, collection string, rec interface{}) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := dbm.mongoClinet.Database(dbname).Collection(collection).InsertOne(ctx, rec)

	return err
}

func (dbm DBManger) Disconnect() error {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := dbm.mongoClinet.Disconnect(ctx)
	return err
}

func (dbm DBManger) QueryRecord(dbname string, collection string, query string) (*mongo.Cursor, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var bquery interface{}
	err := bson.UnmarshalExtJSON([]byte(query), true, &bquery)
	if err != nil {
		// handle error
	}
	cursor, err := dbm.mongoClinet.Database(dbname).Collection(collection).Find(ctx, &bquery)

	return cursor, err

}

func (dbm DBManger) QueryOneRecord(dbname string, collection string, query string) (*mongo.SingleResult, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var bquery interface{}
	err := bson.UnmarshalExtJSON([]byte(query), true, &bquery)

	if err != nil {
		// handle error
	}
	cursor := dbm.mongoClinet.Database(dbname).Collection(collection).FindOne(ctx, &bquery, options.FindOne())

	return cursor, err

}

func (dbm DBManger) UpdateRecord(dbname string, collection string, query string, rec interface{}) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var bquery interface{}
	err := bson.UnmarshalExtJSON([]byte(query), true, &bquery)
	if err != nil {
		// handle error
	}
	_, err = dbm.mongoClinet.Database(dbname).Collection(collection).ReplaceOne(ctx, &bquery, &rec)

	return err

}

func (dbm DBManger) DeleteOneRecord(dbname string, collection string, query string) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var bquery interface{}
	err := bson.UnmarshalExtJSON([]byte(query), true, &bquery)
	if err != nil {
		// handle error
	}
	_, err = dbm.mongoClinet.Database(dbname).Collection(collection).DeleteOne(ctx, &bquery)

	return err

}

func (dbm DBManger) DeleteMultipleRecord(dbname string, collection string, query string) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var bquery interface{}
	err := bson.UnmarshalExtJSON([]byte(query), true, &bquery)
	if err != nil {
		// handle error
	}
	_, err = dbm.mongoClinet.Database(dbname).Collection(collection).DeleteMany(ctx, &bquery)

	return err

}
