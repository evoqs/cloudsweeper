package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"testmod/model"
	"testmod/storage"
	"testmod/utils"

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
	router.HandleFunc("/accounts", srv.PostAccount).Methods("POST")
	http.ListenAndServe(":8000", router)
}

func (srv *Server) GetAllAccount(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	accounts := storage.GetAccounts(srv.dbM, utils.GetConfig().Database.Name, `{}`)
	json.NewEncoder(writer).Encode(accounts)

}

func (srv *Server) PostAccount(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	var acc model.AccountData
	_ = json.NewDecoder(request.Body).Decode(&acc)
	fmt.Println(acc)
	storage.PutAccounts(srv.dbM, utils.GetConfig().Database.Name, acc)
	json.NewEncoder(writer).Encode(`{"status" : "success"}`)

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
		account := accounts[0]
		json.NewEncoder(writer).Encode(account)

	}

}
