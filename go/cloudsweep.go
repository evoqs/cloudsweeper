package main

import (
	"cloudsweep/api"
	"cloudsweep/scheduler"
	"cloudsweep/storage"
	"cloudsweep/utils"
	"fmt"
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
	dbM.SetDatabase(utils.GetConfig().Database.Name)

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
	dbo := storage.MakeDBOperators(dbM)

	//InsertRandomRecord(*dbM, cfg.Database.Name)
	//InsertRandomRecord(*dbM, cfg.Database.Name)
	/*
		accId := 64300
		query := `{"accountId": ` + strconv.Itoa(accId) + `}`
		QueryRandomRecord(*dbM, cfg.Database.Name, query)

		UpdateRandomRecord(*dbM, cfg.Database.Name, query)
		fmt.Println("Query All")


		query = `{"accountId": ` + strconv.Itoa(98407) + `}`
		DeleteRandomRecord(*dbM, cfg.Database.Name, query)
	*/
	//query := `64e97c7a6ca9765964be555e`
	//QueryRandomRecordWithId(*dbM, cfg.Database.Name, query)

	// Start Scheduler
	pipelineScheduler := scheduler.StartDefaultPipelineScheduler()
	pipelineScheduler.ScheduleAllPipelines()

	// Start Server
	startServer(dbo, utils.GetConfig().Server.Host, utils.GetConfig().Server.Port)

	//cpumodel := utils.GetCPUmodel()
	//fmt.Println(cpumodel)
}

func startServer(dbO *storage.DbOperators, host string, port string) {
	fmt.Println("Starting server")
	var server api.Server
	server.StartApiServer(fmt.Sprintf("%s:%s", host, port), *dbO)
}

/*
func InsertRandomRecord(dbM storage.DBManger, dbName string) {
	var acc model.AccountData
	cred := model.AwsCredentials{
		AccessKeyID:     "myaccesskey1234",
		SecretAccessKey: "topsecret",
	}
	acc = model.AccountData{
		AccountID:      "0",
		AccountType:    "aws",
		Description:    "AWS account ",
		AwsCredentials: cred,
	}

	acc.AccountID = strconv.Itoa(rand.Intn(100000))
	acc.Description = acc.Description + strconv.Itoa(rand.Intn(100000))
	fmt.Println(acc)

	recordId, err := dbM.InsertRecord(dbName, "account", &acc)
	if err != nil {
		fmt.Println("Insert record failed with " + err.Error())

	} else {
		fmt.Println("Successfuly Inserted the record " + recordId)
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

func QueryRandomRecordWithId(dbM storage.DBManger, dbName string, query string) {

	var results []model.AccountData

	cursor, err := dbM.QueryRecordWithObjectID(dbName, "account", query)

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
		AccountID:   "00",
		AccountType: "aws",
		Description: "Updated AWS account ",
	}

	acc.AccountID = strconv.Itoa(rand.Intn(100000))
	acc.Description = acc.Description + strconv.Itoa(rand.Intn(100000))
	fmt.Println(acc)

	_, err := dbM.UpdateRecord(dbName, "account", query, &acc)
	if err != nil {
		fmt.Println("Update record failed with " + err.Error())

	} else {
		fmt.Println("Successfuly Updated the record")
	}
}

func DeleteRandomRecord(dbM storage.DBManger, dbName string, query string) {

	err := dbM.DeleteOneRecord(dbName, "account", query)
	if err != nil {
		fmt.Println("Delete record failed with " + err.Error())

	} else {
		fmt.Println("Successfuly deleted the record")
	}
}
*/
