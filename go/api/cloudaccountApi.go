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

	//reading from url
	//vars := mux.Vars(request)
	//accountid := vars["accountid"]

	sweepaccountid := request.Header.Get(AccountIDHeader)
	if !primitive.IsValidObjectID(sweepaccountid) {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid Customer(Sweeper) Account ID: %s", sweepaccountid)))
		return
	}

	query := fmt.Sprintf(`{"sweepaccountid": "%s"}`, sweepaccountid)
	fmt.Println(query)

	accounts, err := srv.opr.AccountOperator.QueryAccountTable(query)

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

	sweepaccountid := request.Header.Get(AccountIDHeader)
	if !primitive.IsValidObjectID(sweepaccountid) {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid Customer(Sweeper) Account ID: %s", sweepaccountid)))
		return
	}

	query := fmt.Sprintf(`{"sweepaccountid": "%s"}`, sweepaccountid)
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
	sweepaccountid := request.Header.Get(AccountIDHeader)
	if !primitive.IsValidObjectID(sweepaccountid) {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid Customer(Sweeper) Account ID: %s", sweepaccountid)))
		return
	}

	var acc model.CloudAccountData
	err := json.NewDecoder(request.Body).Decode(&acc)

	if err != nil {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid json payload for POST request, %s", err.Error())))
		return
	}

	//validate json input
	acc.SweepAccountID = sweepaccountid

	if acc.Name == "" {
		errString := fmt.Sprintf("Cloud Account name cannot be empty.")
		srv.SendResponse400(writer, errors.New(errString))
	}

	//Validate Cloud credentials
	var regionList []string
	if strings.TrimSpace(acc.AccountType) == "aws" {

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

		regionList, err = awsClient.GetSubscribedRegionCodes()
		if err != nil {
			errString := fmt.Sprintf("Failed to fetch AWS Subscribed region with given credentials. %s", err.Error())
			srv.SendResponse409(writer, errors.New(errString))
			return
		}

	} else {
		errString := fmt.Sprintf("Unsupported Account type %s , supported account types are aws, Future expansion: gcp,azure and oci.", acc.AccountType)
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
	defaultPolicyList, err := srv.opr.PolicyOperator.GetAllDefaultPolicyDetails()
	if err != nil {
		srv.SendResponse500(writer, fmt.Errorf("Failed to fetch default policies, %s", err))
		return
	}

	var policyIDList []string
	for _, defaultpolicy := range defaultPolicyList {
		policyIDList = append(policyIDList, defaultpolicy.PolicyID.Hex())
		srv.logwriter.Infof(fmt.Sprintf("Default policy with name %s, added", defaultpolicy.PolicyName))
	}

	if len(policyIDList) != 0 {
		var pipeline model.PipeLine
		pipeline.SweepAccountID = acc.SweepAccountID
		pipeline.CloudAccountID = acc.CloudAccountID.Hex()
		pipeline.Enabled = true
		pipeline.PipeLineName = fmt.Sprintf("Default_%s", acc.Name)
		pipeline.RunStatus = model.UNKNOWN
		schedule := model.Schedule{Minute: "0", Hour: "12", DayOfMonth: "*", Month: "*", DayOfWeek: "*"}
		pipeline.Schedule = schedule
		pipeline.Policies = policyIDList
		pipeline.Default = true
		pipeline.Notification.EmailAddresses = acc.EmailList
		//create aws client and get subscription regions

		//pipeline.ExecutionRegions = []string{"ap-southeast-2"} //TODO make regions to all
		pipeline.ExecutionRegions = regionList
		//Add pipeline
		pipelineid, err := srv.opr.PipeLineOperator.AddPipeLine(pipeline)
		if err != nil {
			srv.logwriter.Errorf(fmt.Sprintf("Failed to add default pipeline for cloud account %s, with error %s", acc.CloudAccountID, err.Error()))
		} else {
			srv.logwriter.Infof(fmt.Sprintf("Added default pipeline for customer sweep account %s, policy name %s", acc.SweepAccountID, pipelineid))
			pipelines, _ := srv.opr.PipeLineOperator.GetPipeLineDetails(pipelineid)
			scheduler.GetDefaultPipelineScheduler().AddPipelineSchedule(pipelines[0])
			runner.ValidateAndRunPipeline(pipelineid)
		}

	} else {
		srv.SendResponse500(writer, fmt.Errorf("No default policies found, failed to add default pipeline"))
		return
	}
}

func (srv *Server) UpdateCloudAccount(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	sweepaccountid := request.Header.Get(AccountIDHeader)
	if !primitive.IsValidObjectID(sweepaccountid) {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid Customer(Sweeper) Account ID: %s", sweepaccountid)))
		return
	}
	//decoding post json to Accountdata Model
	var acc model.CloudAccountData
	err := json.NewDecoder(request.Body).Decode(&acc)
	if err != nil {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid json payload for POST request, %s", err.Error())))
		return
	}

	//original,err := srv.opr.AccountOperator.GetCloudAccountWithObjectID(acc.CloudAccountID.Hex())
	query := fmt.Sprintf(`{"sweepaccountid": "%s", "cloudaccountid": "%s"}`, sweepaccountid, acc.CloudAccountID.Hex())
	original, err := srv.opr.AccountOperator.QueryAccountTable(query)

	if len(original) == 0 {
		srv.SendResponse404(writer, nil)
		return
	}

	if err != nil {
		srv.SendResponse500(writer, fmt.Errorf("Failed to fetch existing account details, %s", err))
		return
	}

	if strings.TrimSpace(acc.AccountType) == strings.TrimSpace(original[0].AccountType) {
		srv.SendResponse400(writer, fmt.Errorf("Account Type should match for update operation, Expected %s, Received %s", original[0].AccountType, acc.AccountType))
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
		errString := fmt.Sprintf("Unsupported Account type %s , supported account types are aws, Future expansion: gcp,azure and oci.", acc.AccountType)
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

	accounts, err := srv.opr.AccountOperator.GetCloudAccountWithObjectID(accountid)
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

	accounts, err := srv.opr.AccountOperator.GetCloudAccountWithObjectID(accountid)
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
