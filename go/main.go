package main

import (
	"cloudsweep/model"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Replace the placeholder with your Atlas connection string
const uri = "mongodb://127.0.0.1:27017"

func main2() {

	// Use the SetServerAPIOptions() method to set the Stable API version to 1
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server

	client, err := mongo.Connect(context.TODO(), opts)

	if err != nil {
		panic(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	var jsonD model.CloudAccountData

	fmt.Println(jsonD)

	var acc model.CloudAccountData
	acc = model.CloudAccountData{
		AccountID:   "123213",
		AccountType: "aws",
		Description: "AWS account ",
	}

	fmt.Println(acc)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	//abc, _ := json.Marshal(jsonD.JsonData)
	_, err = client.Database("test").Collection("movie").InsertOne(ctx, bson.M{})
	if err != nil {
		fmt.Println(err)
	}

	var bdoc interface{}

	//bdoc = `{}

	if err != nil {
		fmt.Println(err)
	}
	col := client.Database("test").Collection("movie1")
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = col.InsertOne(ctx, bdoc)
	if err != nil {
		fmt.Println(err)
	}

}

func main1() {

	// Use the SetServerAPIOptions() method to set the Stable API version to 1
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server

	client, err := mongo.Connect(context.TODO(), opts)

	if err != nil {
		panic(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	var acc model.CloudAccountData
	acc = model.CloudAccountData{
		AccountID:   "123213",
		AccountType: "aws",
		Description: "AWS account ",
	}

	fmt.Println(acc.AccountID)

	var dataSet map[int]*model.CloudAccountData
	dataSet = make(map[int]*model.CloudAccountData)

	// Send a ping to confirm a successful connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var result bson.M

	if err := client.Database("admin").RunCommand(ctx, bson.D{{Key: "ping", Value: 1}}).Decode(&result); err != nil {
		panic(err)
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
	//coll := client.Database("test").Collection("movies")

	start := time.Now()
	for i := 1; i < 200; i++ {
		_, err = client.Database("test").Collection("movie5").InsertOne(ctx, dataSet[i])
		if err != nil {
			fmt.Println(err)
		}
		//fmt.Println(res)
	}

	fmt.Println("Insertion time ", time.Since(start))
}
