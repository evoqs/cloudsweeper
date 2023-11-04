package api

import (
	"cloudsweep/model"
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

	var policy model.Policy
	err := json.NewDecoder(request.Body).Decode(&policy)

	if err != nil {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid json payload for POST request, %s", err.Error())))
		return
	}
	policy.PolicyType = "custom"
	id, err := srv.opr.PolicyOperator.AddPolicy(policy)

	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}

	//srv.SendResponse200(writer, fmt.Sprintf("Successfully Added Policy with ID %s", id))

	writer.WriteHeader(http.StatusOK)
	policy.PolicyID, err = primitive.ObjectIDFromHex(id)
	json.NewEncoder(writer).Encode(policy)
}

func (srv *Server) UpdateCustodianPolicy(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	var policy model.Policy
	err := json.NewDecoder(request.Body).Decode(&policy)

	if err != nil {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid json payload for POST request, %s", err.Error())))
		return
	}
	policy.PolicyType = "custom"
	count, err := srv.opr.PolicyOperator.UpdatePolicy(policy)

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

	policies, err := srv.opr.PolicyOperator.GetPolicyDetails(policyid)

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

	deleteCount, err := srv.opr.PolicyOperator.DeleteCustodianPolicy(policyid)

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

func (srv *Server) GetPolicyRunResult(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")
	policyid := request.URL.Query().Get("policyid")
	cloudaccountid := request.URL.Query().Get("cloudaccountid")

	if !primitive.IsValidObjectID(policyid) {
		srv.logwriter.Warnf(fmt.Sprintf("Invalid Policy ObjectID: %s, received in get result query", policyid))
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid Policy ObjectID: %s", policyid)))
		return
	}

	if !primitive.IsValidObjectID(cloudaccountid) {
		srv.logwriter.Warnf(fmt.Sprintf("Invalid Cloud AccountID: %s, received in get result query", policyid))
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid Cloud AccountID: %s", cloudaccountid)))
		return
	}
	query := fmt.Sprintf(`{"policyid": "%s", "cloudaccountid":"%s"}`, policyid, cloudaccountid)
	policieResults, err := srv.opr.PolicyOperator.GetPolicyResultDetails(query)

	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}

	//TODO when length >1

	if len(policieResults) == 0 {

		srv.SendResponse404(writer, nil)
		return
	} else if len(policieResults) > 1 {
		err := errors.New("Internal Server Error, DB data consistency issue , duplicate policies with same ID")
		srv.SendResponse500(writer, err)
		return
	}

	writer.WriteHeader(http.StatusOK)
	policieResult := policieResults[0]
	json.NewEncoder(writer).Encode(policieResult)
}

func (srv *Server) AddDefaultCustodianPolicy(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")

	var defaultpolicy model.DefaultPolicy
	err := json.NewDecoder(request.Body).Decode(&defaultpolicy)

	if err != nil {
		errmsg := fmt.Sprintf("Invalid json payload for POST request, %s", err.Error())
		srv.logwriter.Errorf(errmsg)
		srv.SendResponse400(writer, errors.New(errmsg))
		return
	}

	id, err := srv.opr.PolicyOperator.AddDefaultPolicy(defaultpolicy)

	if err != nil {
		srv.logwriter.Errorf(fmt.Sprintf("DefaultPolicy Add request failed, %s", err.Error()))
		srv.SendResponse500(writer, err)
		return
	}

	//srv.SendResponse200(writer, fmt.Sprintf("Successfully Added Policy with ID %s", id))
	srv.logwriter.Infof("Successfully added default policy with ID", id)
	writer.WriteHeader(http.StatusOK)
	defaultpolicy.PolicyID, err = primitive.ObjectIDFromHex(id)
	json.NewEncoder(writer).Encode(defaultpolicy)
}

// Update an existing default policy
func (srv *Server) UpdateDefaultCustodianPolicy(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	var defaultpolicy model.DefaultPolicy
	err := json.NewDecoder(request.Body).Decode(&defaultpolicy)

	if err != nil {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid json payload for POST request, %s", err.Error())))
		return
	}

	count, err := srv.opr.PolicyOperator.UpdateDefaultPolicy(defaultpolicy)

	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}

	if count == 0 {
		srv.SendResponse404(writer, nil)
		return
	}

	msg := fmt.Sprintf("Updated %d Policy with ID %s", count, defaultpolicy.PolicyID)
	srv.logwriter.Infof(msg)
	srv.SendResponse200(writer, msg)
}

// List all default policies defined
func (srv *Server) GetDefaultCustodianPolicies(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")

	defaultpolicies, err := srv.opr.PolicyOperator.GetAllDefaultPolicyDetails()

	if err != nil {
		srv.logwriter.Errorf(fmt.Sprintf("DefaultPolicy Getrequest failed, %s", err.Error()))
		srv.SendResponse500(writer, err)
		return
	}

	//TODO when length >1

	if len(defaultpolicies) == 0 {

		srv.SendResponse404(writer, nil)
		return
	}

	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(defaultpolicies)
}

func (srv *Server) DeleteDefaultCustodianPolicy(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(request)

	defaultpolicyid := vars["defaultpolicyid"]
	if !primitive.IsValidObjectID(defaultpolicyid) {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid ObjectID: %s", defaultpolicyid)))
		return
	}

	deleteCount, err := srv.opr.PolicyOperator.DeleteDefaultCustodianPolicy(defaultpolicyid)

	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}

	if deleteCount == 0 {
		srv.SendResponse404(writer, nil)
		return

	} else {
		srv.SendResponse200(writer, fmt.Sprintf("Successfully deleted defaultpolicy, %s", defaultpolicyid))
	}

}
