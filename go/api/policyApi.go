package api

import (
	"cloudsweep/model"
	"cloudsweep/storage"
	"cloudsweep/utils"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (srv *Server) AddCustodianPolicy(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")

	var policy model.CSPolicy
	err := json.NewDecoder(request.Body).Decode(&policy)

	if err != nil {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid json payload for POST request, %s", err.Error())))
		return
	}

	id, err := storage.AddPolicy(srv.dbM, utils.GetConfig().Database.Name, policy)

	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}

	srv.SendResponse200(writer, fmt.Sprintf("Successfully Added Policy with ID %s", id))

	writer.WriteHeader(http.StatusOK)
	policy.PolicyID, err = primitive.ObjectIDFromHex(id)
	json.NewEncoder(writer).Encode(policy)
}

func (srv *Server) UpdateCustodianPolicy(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	var policy model.CSPolicy
	err := json.NewDecoder(request.Body).Decode(&policy)

	if err != nil {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid json payload for POST request, %s", err.Error())))
		return
	}

	count, err := storage.UpdatePolicy(srv.dbM, utils.GetConfig().Database.Name, policy)

	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}

	if count == 0 {
		srv.SendResponse404(writer, nil)
		return
	}

	srv.SendResponse200(writer, fmt.Sprintf("Updated %d Policy with ID %s", count, policy.PolicyID))
}

func (srv *Server) GetCustodianPolicy(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(request)
	policyid := vars["policyid"]

	if !primitive.IsValidObjectID(policyid) {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid ObjectID: %s", policyid)))
		return
	}

	policies, err := storage.GetPolicyDetails(srv.dbM, utils.GetConfig().Database.Name, policyid)

	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}

	//TODO when length >1

	if len(policies) == 0 {

		srv.SendResponse404(writer, nil)
		return
	} else if len(policies) > 1 {
		err := errors.New("Internal Server Error, DB data consistency issue , duplicate policies with same ID")
		srv.SendResponse500(writer, err)
		return
	}

	writer.WriteHeader(http.StatusOK)
	policy := policies[0]
	json.NewEncoder(writer).Encode(policy)
}

func (srv *Server) DeleteCustodianPolicy(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(request)

	policyid := vars["policyid"]
	if !primitive.IsValidObjectID(policyid) {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid ObjectID: %s", policyid)))
		return
	}

	deleteCount, err := storage.DeleteCustodianPolicy(srv.dbM, utils.GetConfig().Database.Name, policyid)

	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}

	if deleteCount == 0 {
		srv.SendResponse404(writer, nil)
		return

	} else {
		srv.SendResponse200(writer, fmt.Sprintf("Successfully deleted policy, %s", policyid))
	}

}
