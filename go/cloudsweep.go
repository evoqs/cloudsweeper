package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"

	"testmod/api"
	"testmod/model"
	"testmod/storage"
	"testmod/utils"
)

func main() {
	cfg := utils.LoadConfig()
	dbUrl, err := utils.GetDBUrl(&cfg)

	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(dbUrl)
	}

	dbM := storage.GetDBManager()
	dbM.SetDbUrl(dbUrl)

	_, err = dbM.Connect()
	if err != nil {
		fmt.Println("Failed to connect to DB " + err.Error())
	}
	err = dbM.CheckConnection()
	if err != nil {
		fmt.Println("Connection Check failed")
	} else {
		fmt.Println("Successfully Connected")
		defer dbM.Disconnect()
	}

	result := storage.GetIdCounter(*dbM, cfg.Database.Name)
	if result.CounterID == 0 {
		fmt.Println("Counter not set, Initializing counter")
		storage.InitializeCounter(*dbM, cfg.Database.Name)

	}

	//fmt.Println(result)
	var server api.Server
	server.StartApiServer("8000", *dbM)
	//InsertRandomRecord(*dbM, cfg.Database.Name)
	//InsertRandomRecord(*dbM, cfg.Database.Name)
	//InsertRandomRecord(*dbM, cfg.Database.Name)
	/*
		accId := 64300
		query := `{"accountId": ` + strconv.Itoa(accId) + `}`
		QueryRandomRecord(*dbM, cfg.Database.Name, query)

		UpdateRandomRecord(*dbM, cfg.Database.Name, query)
		fmt.Println("Query All")
		query = `{}`
		QueryRandomRecord(*dbM, cfg.Database.Name, query)

		query = `{"accountId": ` + strconv.Itoa(98407) + `}`
		DeleteRandomRecord(*dbM, cfg.Database.Name, query)
	*/
}

func InsertRandomRecord(dbM storage.DBManger, dbName string) {
	var acc model.AccountData
	acc = model.AccountData{
		AccountID:   0,
		AccountType: "aws",
		Description: "AWS account ",
	}

	acc.AccountID = rand.Intn(100000)
	acc.Description = acc.Description + strconv.Itoa(rand.Intn(100000))
	fmt.Println(acc)

	err := dbM.InsertRecord(dbName, "account", &acc)
	if err != nil {
		fmt.Println("Insert record failed with " + err.Error())

	} else {
		fmt.Println("Successfuly Inserted the record")
	}
}

func QueryRandomRecord(dbM storage.DBManger, dbName string, query string) {

	var results []model.AccountData

	cursor, err := dbM.QueryRecord(dbName, "account", query)

	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	for _, result := range results {
		cursor.Decode(&result)
		output, err := json.Marshal(result) //, "", "    ")
		//output, err := json.MarshalIndent(result, "", "    ")
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\n", output)
	}
}

func UpdateRandomRecord(dbM storage.DBManger, dbName string, query string) {
	var acc model.AccountData
	acc = model.AccountData{
		AccountID:   0,
		AccountType: "aws",
		Description: "Updated AWS account ",
	}

	acc.AccountID = rand.Intn(100000)
	acc.Description = acc.Description + strconv.Itoa(rand.Intn(100000))
	fmt.Println(acc)

	err := dbM.UpdateRecord(dbName, "account", query, &acc)
	if err != nil {
		fmt.Println("Update record failed with " + err.Error())

	} else {
		fmt.Println("Successfuly Updated the record")
	}
}

func DeleteRandomRecord(dbM storage.DBManger, dbName string, query string) {

	err := dbM.DeleteRecord(dbName, "account", query)
	if err != nil {
		fmt.Println("Delete record failed with " + err.Error())

	} else {
		fmt.Println("Successfuly deleted the record")
	}
}
