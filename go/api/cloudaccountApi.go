package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"cloudsweep/cloud_lib"
	"cloudsweep/model"
	"cloudsweep/runner"
	"cloudsweep/scheduler"
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

	accounts, err := srv.opr.AccountOperator.GetAllAccounts(query)

	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}
	//TODO when length >1
	if len(accounts) == 0 {
		//srv.SendResponse404(writer, nil)
		json.NewEncoder(writer).Encode(make([]string, 0))
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

	accountid := vars["accountid"]
	fmt.Println(vars)
	query := fmt.Sprintf(`{"accountid": "%s"}`, accountid)
	fmt.Println(query)
	deleteCount, err := srv.opr.AccountOperator.DeleteAllCloudAccounts(query)

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
	var acc model.CloudAccountData
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

	//Validate Cloud credentials
	var regionList []string
	if strings.TrimSpace(acc.AccountType) == "aws" {
		awsClient, err := cloud_lib.GetAwsClient(acc.AwsCredentials.AccessKeyID, acc.AwsCredentials.SecretAccessKey, "af-south-1")

		if err != nil {
			srv.SendResponse500(writer, err)
			return
		}
		acc.AwsCredentials.AccountID, err = awsClient.GetAwsAccountID()

		if err != nil {
			errString := fmt.Sprintf("Failed to fetch AWS Account Id with given credentials. %s", err.Error())
			srv.SendResponse409(writer, errors.New(errString))
			return
		}

		regionList, err = awsClient.GetSubscribedRegionCodes()
		if err != nil {
			errString := fmt.Sprintf("Failed to fetch AWS Subscribed region with given credentials. %s", err.Error())
			srv.SendResponse409(writer, errors.New(errString))
			return
		}
		/* Old validation implementaion
		if !utils.ValidateAwsCredentials(acc.AwsCredentials.AccessKeyID, acc.AwsCredentials.SecretAccessKey) {
			errString := fmt.Sprintf("AWS Authentication Failed with given access key and secret")
			err := errors.New(errString)
			srv.logwriter.Warnf(errString)
			srv.SendResponse409(writer, err)
			return
		}*/

	} else {
		errString := fmt.Sprintf("Unknown Account type %s , supported account types are aws,gcp,azure and oci.", acc.AccountType)
		err := errors.New(errString)
		srv.logwriter.Warnf(errString)
		srv.SendResponse400(writer, err)
		return
	}

	//Writing cloundaccount data to MongoDB
	id, err := srv.opr.AccountOperator.AddCloudAccount(acc)
	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}

	//Sending response
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	objID, err := primitive.ObjectIDFromHex(id)
	acc.CloudAccountID = objID
	json.NewEncoder(writer).Encode(acc)

	//Getting default regions
	defaultPolicyList, _ := srv.opr.PolicyOperator.GetAllDefaultPolicyDetails()
	var policyIDList []string
	for _, defaultpolicy := range defaultPolicyList {
		var policy model.Policy
		policy.PolicyName = defaultpolicy.PolicyName
		policy.PolicyDefinition = defaultpolicy.PolicyDefinition
		policy.Recommendation = defaultpolicy.Recommendation
		policy.PolicyType = "Default"
		policy.AccountID = acc.AccountID

		query := fmt.Sprintf(`{"policyname": "%s", "policytype": "Default"}`, policy.PolicyName)
		result, _ := srv.opr.PolicyOperator.GetAllPolicyDetails(query)
		if len(result) == 0 {
			id, err := srv.opr.PolicyOperator.AddPolicy(policy)

			if err != nil {
				srv.logwriter.Errorf(fmt.Sprintf("Failed to add default policy %s, with error %s.", policy.PolicyName, err.Error()))
			} else {
				srv.logwriter.Infof(fmt.Sprintf("Added default policy for account %s, policy name %s", acc.AccountID, policy.PolicyName))
				policyIDList = append(policyIDList, id)
			}
		} else {
			policyIDList = append(policyIDList, result[0].PolicyID.Hex())
			srv.logwriter.Infof(fmt.Sprintf("Default policy  with name %s, already existing for account %s", policy.PolicyName, acc.AccountID))
		}
	}

	if len(policyIDList) != 0 {
		var pipeline model.PipeLine
		pipeline.AccountID = acc.AccountID
		pipeline.CloudAccountID = acc.CloudAccountID.Hex()
		pipeline.Enabled = true
		pipeline.PipeLineName = fmt.Sprintf("Default_%s", acc.Name)
		pipeline.RunStatus = model.UNKNOWN
		schedule := model.Schedule{Minute: "0", Hour: "12", DayOfMonth: "*", Month: "*", DayOfWeek: "*"}
		pipeline.Schedule = schedule
		pipeline.Policies = policyIDList
		pipeline.Default = true

		//create aws client and get subscription regions

		//pipeline.ExecutionRegions = []string{"ap-southeast-2"} //TODO make regions to all
		pipeline.ExecutionRegions = regionList
		//Add pipeline
		pipelineid, err := srv.opr.PipeLineOperator.AddPipeLine(pipeline)
		if err != nil {
			srv.logwriter.Errorf(fmt.Sprintf("Failed to add default pipeline for cloud account %s, with error %s", acc.CloudAccountID, err.Error()))
		} else {
			srv.logwriter.Infof(fmt.Sprintf("Added default pipeline for account %s, policy name %s", acc.AccountID, pipelineid))
			pipelines, _ := srv.opr.PipeLineOperator.GetPipeLineDetails(pipelineid)
			scheduler.GetDefaultPipelineScheduler().AddPipelineSchedule(pipelines[0])
			runner.ValidateAndRunPipeline(pipelineid)
		}

	}
}

func (srv *Server) UpdateCloudAccount(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	//decoding post json to Accountdata Model
	var acc model.CloudAccountData
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

	if strings.TrimSpace(acc.AccountType) == "aws" {
		awsClient, err := cloud_lib.GetAwsClient(acc.AwsCredentials.AccessKeyID, acc.AwsCredentials.SecretAccessKey, "")

		if err != nil {
			srv.SendResponse500(writer, err)
			return
		}
		acc.AwsCredentials.AccountID, err = awsClient.GetAwsAccountID()

		if err != nil {
			errString := fmt.Sprintf("Failed to fetch AWS Account Id with given credentials. %s", err.Error())
			srv.SendResponse409(writer, errors.New(errString))
			return
		}

	} else {
		errString := fmt.Sprintf("Unknown Account type %s , supported account types are aws,gcp,azure and oci.", acc.AccountType)
		err := errors.New(errString)
		srv.logwriter.Warnf(errString)
		srv.SendResponse400(writer, err)
		return
	}

	updateCount, err := srv.opr.AccountOperator.UpdateCloudAccount(acc)
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

	if !primitive.IsValidObjectID(accountid) {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid ObjectID: %s", accountid)))
		return
	}

	accounts, err := srv.opr.AccountOperator.GetCloudAccount(accountid)
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

func (srv *Server) AuthCheckCloudAccount(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(request)

	accountid := vars["cloudaccountid"]
	if !primitive.IsValidObjectID(accountid) {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid ObjectID: %s", accountid)))
		return
	}

	accounts, err := srv.opr.AccountOperator.GetCloudAccount(accountid)
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

	account := accounts[0]
	if utils.ValidateAwsCredentials(account.AwsCredentials.AccessKeyID, account.AwsCredentials.SecretAccessKey) {
		srv.SendResponse200(writer, "Authentication Succeeded")
	} else {
		srv.SendResponse409(writer, errors.New("Authentication Failed"))
	}
}

func (srv *Server) DeleteCloudAccount(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(request)

	accountid := vars["cloudaccountid"]
	if !primitive.IsValidObjectID(accountid) {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid ObjectID: %s", accountid)))
		return
	}

	deleteCount, err := srv.opr.AccountOperator.DeleteCloudAccount(accountid)

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
