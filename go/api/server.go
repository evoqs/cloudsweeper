package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"cloudsweep/model"
	"cloudsweep/storage"
	"cloudsweep/utils"

	"github.com/gorilla/mux"
)

type Server struct {
	dbM  storage.DBManger
	port string
}

func (srv *Server) StartApiServer(port string, dbM storage.DBManger) {
	srv.port = port
	srv.dbM = dbM
	router := mux.NewRouter()
	router.HandleFunc("/accounts", srv.GetAllAccount).Methods("GET")
	router.HandleFunc("/accounts/{accountid}", srv.GetAccount).Methods("GET")
	router.HandleFunc("/accounts/{accountid}", srv.DeleteAccount).Methods("DELETE")

	router.HandleFunc("/cloudaccount", srv.AddCloudAccount).Methods("POST")
	router.HandleFunc("/cloudaccount/{cloudaccountid}", srv.GetCloudAccount).Methods("GET")
	router.HandleFunc("/cloudaccount/{cloudaccountid}", srv.DeleteCloudAccount).Methods("DELETE")

	http.ListenAndServe(":8000", router)
}

func (srv *Server) GetAllAccount(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	accounts := storage.GetAccounts(srv.dbM, utils.GetConfig().Database.Name, `{}`)
	json.NewEncoder(writer).Encode(accounts)

}

func (srv *Server) GetAccount(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(request)
	accountid := vars["accountid"]
	query := `{"accountid":` + accountid + `}`
	fmt.Println(query)
	accounts := storage.GetAccounts(srv.dbM, utils.GetConfig().Database.Name, query)

	//TODO when length >1
	if len(accounts) == 0 {
		writer.WriteHeader(http.StatusNotFound)
		json.NewEncoder(writer).Encode(nil)

	} else {
		writer.WriteHeader(http.StatusOK)
		json.NewEncoder(writer).Encode(accounts)

	}

}

func (srv *Server) DeleteAccount(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(request)
	//accountid := vars["accountid"]
	accountid := vars["accountid"]
	fmt.Println(vars)
	query := `{"accountid":` + accountid + `}`
	fmt.Println(query)
	err := storage.DeleteAccounts(srv.dbM, utils.GetConfig().Database.Name, query)

	//TODO when length >1
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		json.NewEncoder(writer).Encode(nil)

	} else {
		writer.WriteHeader(http.StatusOK)
		json.NewEncoder(writer).Encode(`{"status" : "success"}`)

	}

}

// ************************************* Cloud Account related functions ********************************************
func (srv *Server) AddCloudAccount(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	//decoding post json to Accountdata Model
	var acc model.AccountData
	_ = json.NewDecoder(request.Body).Decode(&acc)

	//Getting and updating the cloudaccountID counter
	//TODO Get and update can be combined
	counter := storage.GetIdCounter(srv.dbM, utils.GetConfig().Database.Name)
	cloudAccountID := counter.NextCloudAccountID
	fmt.Println("Counter: ", counter)
	acc.CloudAccountID = cloudAccountID
	counter.NextCloudAccountID = cloudAccountID + 1
	storage.UpdateIdCounter(srv.dbM, utils.GetConfig().Database.Name, counter)

	storage.PutAccounts(srv.dbM, utils.GetConfig().Database.Name, acc)
	json.NewEncoder(writer).Encode(`{"status" : "success"}`)

}

func (srv *Server) GetCloudAccount(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(request)
	accountid := vars["cloudaccountid"]
	query := `{"cloudaccountid":` + accountid + `}`
	fmt.Println(query)
	accounts := storage.GetAccounts(srv.dbM, utils.GetConfig().Database.Name, query)

	//TODO when length >1
	if len(accounts) == 0 {
		writer.WriteHeader(http.StatusNotFound)
		json.NewEncoder(writer).Encode(nil)

	} else {
		writer.WriteHeader(http.StatusOK)
		account := accounts[0]
		json.NewEncoder(writer).Encode(account)

	}

}

func (srv *Server) DeleteCloudAccount(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(request)
	//accountid := vars["accountid"]
	accountid := vars["cloudaccountid"]
	fmt.Println(vars)
	query := `{"cloudaccountid":` + accountid + `}`
	fmt.Println(query)
	err := storage.DeleteCloudAccount(srv.dbM, utils.GetConfig().Database.Name, query)

	//TODO when length >1
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		json.NewEncoder(writer).Encode(nil)

	} else {
		writer.WriteHeader(http.StatusOK)
		json.NewEncoder(writer).Encode(`{"status" : "success"}`)

	}

}
