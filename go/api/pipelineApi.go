package api

import (
	"cloudsweep/model"
	"cloudsweep/policy_converter"
	"cloudsweep/utils"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (srv *Server) RunPipeLine(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(request)
	pipelineid := vars["pipelineid"]

	//Get pipline and policy details , Skipping pipeline implemetation and directly using policies
	//TODO

	pipeLineList, err := srv.opr.PipeLineOperator.GetPipeLineDetails(pipelineid)
	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}
	if len(pipeLineList) == 0 {
		srv.SendResponse404(writer, nil)
		return
	}

	pipeLine := pipeLineList[0]

	//Getting the policy

	policyids := pipeLine.PolicyID
	if len(policyids) == 0 {
		srv.SendResponse404(writer, errors.New("No policy Defined for Pipeline."))
	}

	for _, policyid := range policyids {

		policyList, err := srv.opr.PolicyOperator.GetPolicyDetails(policyid)
		if err != nil {
			srv.SendResponse500(writer, err)
			return
		}

		if len(policyList) == 0 {
			srv.SendResponse404(writer, nil)
			return
		}
		fmt.Println("Length", len(policyList))
		policy := policyList[0]
		policyJson := policy.PolicyDefinition

		/*Steps involved in running
		1. Validating a run is not already in progress
		2. create a run folder for policy in /tmp
		3. fecthing the policy json and converting into yaml
		4. Fectch the cloud credentails for the policy run
		5. trigger the run and update the pipleline status as running
		6. Send Accepted response 200

		*/

		//Step 1 TODO
		//Step 2 create runfolder
		RunFolder := fmt.Sprintf("/tmp/%s.%s", policyid, strconv.Itoa(rand.Intn(100000)))
		os.Mkdir(RunFolder, os.ModePerm)

		//Step 3
		policyFile := fmt.Sprintf("%s/%s", RunFolder, "policy.yml")
		err = policy_converter.ConvertJsonToYamlAndWriteToFile(policyJson, policyFile)
		if err != nil {
			srv.SendResponse500(writer, err)
			return
		}

		//Get creds
		cloudAccList, err := srv.opr.AccountOperator.GetCloudAccount(policy.CloudAccountID)
		if err != nil {
			srv.SendResponse500(writer, err)
			return
		}

		cloudAcc := cloudAccList[0]
		if cloudAcc.AccountType == model.AWS {
			srv.SendResponse200(writer, fmt.Sprintf("Recived run request for %s, created runfolder %s", pipelineid, policyJson))
			var envvars []string
			envvars = append(envvars, fmt.Sprintf("AWS_DEFAULT_REGION=%s", cloudAcc.AwsCredentials.Region))
			envvars = append(envvars, fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", cloudAcc.AwsCredentials.AccessKeyID))
			envvars = append(envvars, fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", cloudAcc.AwsCredentials.SecretAccessKey))

			activatePath := utils.GetConfig().Custodian.C7nAwsInstall
			utils.RunCustodianPolicy(envvars, RunFolder, policyFile, activatePath)
			return
		}

		srv.SendResponse500(writer, errors.New("Internal Server Error"))
	}
}

func (srv *Server) AddPipeLine(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")

	var pipeline model.PipeLine
	err := json.NewDecoder(request.Body).Decode(&pipeline)

	if err != nil {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid json payload for POST request, %s", err.Error())))
		return
	}

	id, err := srv.opr.PipeLineOperator.AddPipeLine(pipeline)

	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}

	//srv.SendResponse200(writer, fmt.Sprintf("Successfully Added Policy with ID %s", id))

	writer.WriteHeader(http.StatusOK)
	pipeline.PipeLineID, err = primitive.ObjectIDFromHex(id)
	json.NewEncoder(writer).Encode(pipeline)

}

func (srv *Server) UpdatePipeLine(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	var pipeline model.PipeLine
	err := json.NewDecoder(request.Body).Decode(&pipeline)

	if err != nil {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid json payload for POST request, %s", err.Error())))
		return
	}

	count, err := srv.opr.PipeLineOperator.UpdatePipeLine(pipeline)

	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}

	if count == 0 {
		srv.SendResponse404(writer, nil)
		return
	}

	srv.SendResponse200(writer, fmt.Sprintf("Updated %d Policy with ID %s", count, pipeline.PipeLineID))
}

func (srv *Server) GetPipeLine(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(request)
	pipelineid := vars["pipelineid"]

	if !primitive.IsValidObjectID(pipelineid) {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid ObjectID: %s", pipelineid)))
		return
	}

	pipline, err := srv.opr.PipeLineOperator.GetPipeLineDetails(pipelineid)

	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}

	//TODO when length >1

	if len(pipline) == 0 {

		srv.SendResponse404(writer, nil)
		return
	} else if len(pipline) > 1 {
		err := errors.New("Internal Server Error, DB data consistency issue , duplicate pipline with same ID")
		srv.SendResponse500(writer, err)
		return
	}

	writer.WriteHeader(http.StatusOK)
	pipeln := pipline[0]
	json.NewEncoder(writer).Encode(pipeln)
}

func (srv *Server) GetAllPipeLine(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(request)
	accountid := vars["accountid"]
	//query := `{"accountid": ` + accountid + `}`
	query := fmt.Sprintf(`{"accountid": "%s"}`, accountid)
	fmt.Println(query)

	pipelines, err := srv.opr.PipeLineOperator.GetAllPipeLines(query)

	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}
	//TODO when length >1
	if len(pipelines) == 0 {
		srv.SendResponse404(writer, nil)
		return

	} else {
		writer.WriteHeader(http.StatusOK)
		json.NewEncoder(writer).Encode(pipelines)
	}

}

func (srv *Server) DeletePipeLine(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(request)

	pipelineid := vars["pipelineid"]
	if !primitive.IsValidObjectID(pipelineid) {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid ObjectID: %s", pipelineid)))
		return
	}

	deleteCount, err := srv.opr.PipeLineOperator.DeletePipeLine(pipelineid)

	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}

	if deleteCount == 0 {
		srv.SendResponse404(writer, nil)
		return

	} else {
		srv.SendResponse200(writer, fmt.Sprintf("Successfully deleted pipeline, %s", pipelineid))
	}

}
