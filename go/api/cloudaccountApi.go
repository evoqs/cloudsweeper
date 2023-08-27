package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"cloudsweep/model"
	"cloudsweep/storage"
	"cloudsweep/utils"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Fecthes all cloud accounts associated with a sweepAccount
func (srv *Server) GetAllCloudAccount(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(request)
	accountid := vars["accountid"]
	//query := `{"accountid": ` + accountid + `}`
	query := fmt.Sprintf(`{"accountid": "%s"}`, accountid)
	fmt.Println(query)
	accounts, err := storage.GetAllAccounts(srv.dbM, utils.GetConfig().Database.Name, query)

	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}
	//TODO when length >1
	if len(accounts) == 0 {
		srv.SendResponse404(writer, nil)
		return

	} else {
		writer.WriteHeader(http.StatusOK)
		json.NewEncoder(writer).Encode(accounts)
	}

}

// Delete all cloud accounts associated with a sweepAccount
func (srv *Server) DeleteAllCloudAccount(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(request)
	//accountid := vars["accountid"]
	accountid := vars["accountid"]
	fmt.Println(vars)
	query := fmt.Sprintf(`{"accountid": "%s"}`, accountid)
	fmt.Println(query)
	deleteCount, err := storage.DeleteAllCloudAccounts(srv.dbM, utils.GetConfig().Database.Name, query)

	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}

	if deleteCount == 0 {
		srv.SendResponse404(writer, nil)
		return
	} else {
		srv.SendResponse200(writer, fmt.Sprintf("Deleted %d cloud accounts successfully.", deleteCount))
		return
	}

}

// ************************************* Cloud Account related functions ********************************************
func (srv *Server) AddCloudAccount(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	//decoding post json to Accountdata Model
	var acc model.AccountData
	err := json.NewDecoder(request.Body).Decode(&acc)

	if err != nil {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid json payload for POST request, %s", err.Error())))
		return
	}

	//validate json input
	if acc.AccountID == "" {
		err := errors.New("Account ID cannot be null")
		srv.SendResponse400(writer, err)
		return
	}

	if acc.AwsCredentials.AccessKeyID == "" {
		err := errors.New("AWS AccessKeyID cannot be null")
		srv.SendResponse400(writer, err)
		return
	}

	if acc.AwsCredentials.SecretAccessKey == "" {
		err := errors.New("AWS Access Secret cannot be null")
		srv.SendResponse400(writer, err)
		return
	}

	//Writing cloundaccount data to MongoDB
	id, err := storage.AddCloudAccount(srv.dbM, utils.GetConfig().Database.Name, acc)
	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	objID, err := primitive.ObjectIDFromHex(id)
	acc.CloudAccountID = objID
	json.NewEncoder(writer).Encode(acc)

}

func (srv *Server) UpdateCloudAccount(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	//decoding post json to Accountdata Model
	var acc model.AccountData
	err := json.NewDecoder(request.Body).Decode(&acc)

	if err != nil {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid json payload for POST request, %s", err.Error())))
		return
	}

	//validate json input
	if acc.AccountID == "" {
		err := errors.New("Account ID cannot be null")
		srv.SendResponse400(writer, err)
		return
	}

	if acc.AwsCredentials.AccessKeyID == "" {
		err := errors.New("AWS AccessKeyID cannot be null")
		srv.SendResponse400(writer, err)
		return
	}

	if acc.AwsCredentials.SecretAccessKey == "" {
		err := errors.New("AWS Access Secret cannot be null")
		srv.SendResponse400(writer, err)
		return
	}

	updateCount, err := storage.UpdateCloudAccount(srv.dbM, utils.GetConfig().Database.Name, acc)
	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}

	fmt.Printf("Update Count: %d", updateCount)

	if updateCount == 0 {
		srv.SendResponse404(writer, nil)
		return
	} else {
		srv.SendResponse200(writer, fmt.Sprintf("Updated cloud account successfully."))
		return
	}

}

func (srv *Server) GetCloudAccount(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(request)
	accountid := vars["cloudaccountid"]
	accounts, err := storage.GetCloudAccount(srv.dbM, utils.GetConfig().Database.Name, accountid)

	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}

	//TODO when length >1

	if len(accounts) == 0 {

		srv.SendResponse404(writer, nil)
		return
	} else if len(accounts) > 1 {
		err := errors.New("Internal Server Error, DB data consistency issue")
		srv.SendResponse500(writer, err)
		return
	}

	writer.WriteHeader(http.StatusOK)
	account := accounts[0]
	json.NewEncoder(writer).Encode(account)

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
	deleteCount, err := storage.DeleteCloudAccount(srv.dbM, utils.GetConfig().Database.Name, query)

	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}

	if deleteCount == 0 {
		srv.SendResponse404(writer, nil)
		return

	} else {
		srv.SendResponse200(writer, fmt.Sprintf("Successfully deleted cloudaccount, %s", accountid))

	}

}
