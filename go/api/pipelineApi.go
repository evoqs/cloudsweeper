package api

import (
	"cloudsweep/model"
	"cloudsweep/runner"
	"cloudsweep/scheduler"
	"cloudsweep/utils"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (srv *Server) RunPipeLine(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(request)
	pipelineid := vars["pipelineid"]

	sweepaccountid := request.Header.Get(AccountIDHeader)
	if !primitive.IsValidObjectID(sweepaccountid) {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid Customer(Sweeper) Account ID: %s", sweepaccountid)))
		return
	}

	pipeline, err := srv.opr.PipeLineOperator.GetPipeLineDetails(pipelineid)
	if err != nil {
		srv.SendResponse500(writer, fmt.Errorf("Failed to get pipline details, %s", err))
		return
	}

	if len(pipeline) == 0 {
		srv.SendResponse404(writer, err)
		return
	}

	if pipeline[0].SweepAccountID != sweepaccountid {
		srv.SendResponse404(writer, err)
		return
	}

	if !pipeline[0].Enabled {
		srv.SendResponse400(writer, fmt.Errorf("Cannot run a disabled policy"))
		return
	}
	rc, err := runner.ValidateAndRunPipeline(pipelineid)
	if rc == 200 {
		srv.SendResponse200(writer, "Accepted pipeline request for run.")
	} else if rc == 500 {
		srv.SendResponse500(writer, err)
	} else if rc == 404 {
		srv.SendResponse404(writer, err)
	} else if rc == 409 {
		srv.SendResponse409(writer, err)
	}
}

func (srv *Server) AddPipeLine(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")
	sweepaccountid := request.Header.Get(AccountIDHeader)
	if !primitive.IsValidObjectID(sweepaccountid) {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid Customer(Sweeper) Account ID: %s", sweepaccountid)))
		return
	}

	var pipeline model.PipeLine
	err := json.NewDecoder(request.Body).Decode(&pipeline)

	if err != nil {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid json payload for POST request, %s", err.Error())))
		return
	}

	if len(pipeline.ExecutionRegions) == 0 {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid request, Atleast once execution region is needed to create pipline.")))
		return
	}

	if len(pipeline.Policies) == 0 {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid request, Atleast once policy is needed to create pipline.")))
		return
	}

	if utils.CheckDuplicates[string](pipeline.Policies) {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Duplicate polices in pileline create request")))
		return
	}

	if utils.CheckDuplicates[string](pipeline.ExecutionRegions) {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Duplicate execution regions in pileline create request")))
		return
	}

	if !primitive.IsValidObjectID(pipeline.CloudAccountID) {
		srv.SendResponse400(writer, fmt.Errorf("Invalid cloud Account ID: %s", pipeline.CloudAccountID))
		return
	}

	//Validate the cloud account belongs to same sweepid
	cloudaccounts, err := srv.opr.AccountOperator.GetCloudAccountWithObjectID(pipeline.CloudAccountID)
	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}

	if len(cloudaccounts) == 0 {

		srv.SendResponse404(writer, fmt.Errorf("Cannot find cloud account %s", pipeline.CloudAccountID))
		return
	} else if len(cloudaccounts) > 1 {
		err := errors.New("Internal Server Error, DB data consistency issue")
		srv.SendResponse500(writer, err)
		return
	}

	if cloudaccounts[0].SweepAccountID != sweepaccountid {
		srv.SendResponse404(writer, fmt.Errorf("Cannot find cloud account %s", pipeline.CloudAccountID))
		return
	}

	//Validate all policies are belonging to the same sweep account
	for count := range pipeline.Policies {
		policyid := pipeline.Policies[count]
		if !primitive.IsValidObjectID(policyid) {
			srv.SendResponse400(writer, fmt.Errorf("Invalid Policy ID: %s", policyid))
			return
		}
		policies, err := srv.opr.PolicyOperator.GetPolicyDetails(pipeline.Policies[count])

		if err != nil {
			srv.SendResponse500(writer, err)
			return
		}

		if len(policies) == 0 {
			srv.SendResponse404(writer, fmt.Errorf("Cannot find policy %s ", policyid))
			return
		}

		if policies[0].SweepAccountID != sweepaccountid {
			srv.SendResponse404(writer, fmt.Errorf("Cannot find policy %s", policyid))
			return
		}
	}

	pipeline.SweepAccountID = sweepaccountid
	pipeline.Default = false
	pipeline.RunStatus = model.UNKNOWN
	pipeline.LastRunTime = 0

	id, err := srv.opr.PipeLineOperator.AddPipeLine(pipeline)
	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}
	pipelines, err := srv.opr.PipeLineOperator.GetPipeLineDetails(id)
	if err != nil || len(pipelines) < 1 {
		srv.SendResponse500(writer, err)
		return
	}
	// Schedule the newly added pipeline
	if pipeline.Enabled {
		err = scheduler.GetDefaultPipelineScheduler().AddPipelineSchedule(pipelines[0])
		if err != nil {
			srv.SendResponse207(writer, fmt.Errorf("Pipeline added Successfully, Failed to add schedule for the pipeline"))
			return
		}
	}
	//srv.SendResponse200(writer, fmt.Sprintf("Successfully Added Policy with ID %s", id))

	writer.WriteHeader(http.StatusOK)
	pipeline.PipeLineID, err = primitive.ObjectIDFromHex(id)
	json.NewEncoder(writer).Encode(pipeline)

}

func (srv *Server) UpdatePipeLine(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	var requestPipeline model.PipeLine
	err := json.NewDecoder(request.Body).Decode(&requestPipeline)

	if !primitive.IsValidObjectID(requestPipeline.PipeLineID.Hex()) {
		srv.SendResponse400(writer, fmt.Errorf("Invalid Pipeline ID  %s", requestPipeline.PipeLineID.Hex()))
		return
	}

	sweepaccountid := request.Header.Get(AccountIDHeader)
	if !primitive.IsValidObjectID(sweepaccountid) {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid Customer(Sweeper) Account ID: %s", sweepaccountid)))
		return
	}

	if err != nil {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid json payload for POST request, %s", err.Error())))
		return
	}

	if len(requestPipeline.ExecutionRegions) == 0 {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid request, Atleast once execution region is needed to create pipline.")))
		return
	}

	if len(requestPipeline.Policies) == 0 {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid request, Atleast once policy is needed to create pipline.")))
		return
	}

	if utils.CheckDuplicates[string](requestPipeline.Policies) {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Duplicate polices in pileline update request")))
		return
	}

	if utils.CheckDuplicates[string](requestPipeline.ExecutionRegions) {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Duplicate execution regions in pipeline update request")))
		return
	}

	pipelines, err := srv.opr.PipeLineOperator.GetPipeLineDetails(requestPipeline.PipeLineID.Hex())
	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}

	if len(pipelines) == 0 {
		srv.SendResponse404(writer, nil)
		return
	}

	original := pipelines[0]
	if sweepaccountid != original.SweepAccountID {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Pipeline Account ID not matching with existing pipeline Account ID")))
		return
	}

	if requestPipeline.CloudAccountID != original.CloudAccountID {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Pipeline Cloud Account ID not matching with existing ID")))
		return
	}

	fmt.Printf("Run status %d %d %s", requestPipeline.RunStatus, model.RUNNING, original.PipeLineID)
	if original.RunStatus == model.RUNNING {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Cannot update pipeline while it is running")))
		return
	}

	//Validate Policies
	if original.Default {
		requestPipeline.Policies = original.Policies
	} else {

		for _, policy := range requestPipeline.Policies {

			if !primitive.IsValidObjectID(policy) {
				srv.SendResponse400(writer, fmt.Errorf("Invalid Policy ID: %s", policy))
				return
			}
			policyDetail, err := srv.opr.PolicyOperator.GetPolicyDetails(policy)

			if err != nil {
				srv.SendResponse500(writer, errors.New(fmt.Sprintf("Failed to fetch pipeline policy details, %s", err.Error())))
				return
			}
			if len(policyDetail) == 0 {
				srv.SendResponse404(writer, fmt.Errorf("Cannot find policy %s ", policy))
				return
			}
			if policyDetail[0].SweepAccountID != original.SweepAccountID {
				srv.SendResponse400(writer, errors.New(fmt.Sprintf("Policy %s not belonging to Sweep Account %s", policy, original.SweepAccountID)))
				return
			}
		}
	}

	//TODO Get default pipeline IDS and update.

	requestPipeline.RunStatus = original.RunStatus
	requestPipeline.LastRunTime = original.LastRunTime
	requestPipeline.Default = original.Default
	requestPipeline.SweepAccountID = sweepaccountid

	count, err := srv.opr.PipeLineOperator.UpdatePipeLine(requestPipeline)

	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}

	//if requestPipeline.Schedule != nil {
	if requestPipeline.Enabled {
		err = scheduler.GetDefaultPipelineScheduler().UpdatePipelineSchedule(requestPipeline)
		if err != nil {
			srv.SendResponse207(writer, fmt.Errorf("Pipeline Updated Successfully, Failed to update schedule for the pipeline"))
			return
		}
	} else {
		scheduler.GetDefaultPipelineScheduler().DeletePipelineSchedule(requestPipeline.PipeLineID.Hex())
	}
	//}

	srv.SendResponse200(writer, fmt.Sprintf("Updated %d Pipeline with ID %s", count, requestPipeline.PipeLineID))
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

	sweepaccountid := request.Header.Get(AccountIDHeader)
	if !primitive.IsValidObjectID(sweepaccountid) {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid Customer(Sweeper) Account ID: %s", sweepaccountid)))
		return
	}

	pipline, err := srv.opr.PipeLineOperator.GetPipeLineDetails(pipelineid)

	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}

	if len(pipline) == 0 {

		srv.SendResponse404(writer, nil)
		return
	}
	pipeln := pipline[0]

	if pipeln.SweepAccountID != sweepaccountid {
		srv.SendResponse404(writer, nil)
		return
	}

	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(pipeln)
}

func (srv *Server) GetAllPipeLine(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")

	sweepaccountid := request.Header.Get(AccountIDHeader)
	if !primitive.IsValidObjectID(sweepaccountid) {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid Customer(Sweeper) Account ID: %s", sweepaccountid)))
		return
	}

	query := fmt.Sprintf(`{"sweepaccountid": "%s"}`, sweepaccountid)
	pipelines, err := srv.opr.PipeLineOperator.QueryPipeLineDetails(query)

	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}
	//TODO when length >1
	if len(pipelines) == 0 {
		//srv.SendResponse404(writer, nil)
		json.NewEncoder(writer).Encode(make([]string, 0))
		return

	} else {
		writer.WriteHeader(http.StatusOK)
		json.NewEncoder(writer).Encode(pipelines)
	}

}

func (srv *Server) GetAllPolicies(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")

	sweepaccountid := request.Header.Get(AccountIDHeader)
	if !primitive.IsValidObjectID(sweepaccountid) {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid Customer(Sweeper) Account ID: %s", sweepaccountid)))
		return
	}

	query := fmt.Sprintf(`{"sweepaccountid": "%s"}`, sweepaccountid)
	srv.logwriter.Infof("Get all policies for account: ", sweepaccountid)

	policies, err := srv.opr.PolicyOperator.QueryPolicyDetails(query)

	if err != nil {
		srv.logwriter.Errorf("Get all policies for account: ", sweepaccountid, ",failed with error:", err)
		srv.SendResponse500(writer, err)
		return
	}
	//TODO when length >1
	if len(policies) == 0 {
		srv.logwriter.Infof("Get all policies for account: ", sweepaccountid, ",returned empty")
		json.NewEncoder(writer).Encode(make([]string, 0))
		//srv.SendResponse404(writer, nil)
		return

	} else {
		srv.logwriter.Infof("Get all policies for account: ", sweepaccountid, "succcess")
		writer.WriteHeader(http.StatusOK)
		json.NewEncoder(writer).Encode(policies)
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

	sweepaccountid := request.Header.Get(AccountIDHeader)
	if !primitive.IsValidObjectID(sweepaccountid) {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid Customer(Sweeper) Account ID: %s", sweepaccountid)))
		return
	}

	pipeline, err := srv.opr.PipeLineOperator.GetPipeLineDetails(pipelineid)
	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}

	if len(pipeline) == 0 {
		srv.SendResponse404(writer, nil)
		return

	}

	if pipeline[0].SweepAccountID != sweepaccountid {
		srv.SendResponse404(writer, nil)
		return
	}

	if pipeline[0].Default {
		err := errors.New("Cannot delete default pipeline, Only disabled allowed")
		srv.logwriter.Warnf(err.Error())
		srv.SendResponse400(writer, err)
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
		//Delete Results
		query := fmt.Sprintf(`{"pipelineid": "%s"}`, pipelineid)
		_, err = srv.opr.PipeLineOperator.DeletePipelineRunResult(query)
		if err != nil {
			srv.SendResponse500(writer, fmt.Errorf("Failed to delete results for pipeline %s , with error %s", pipelineid, err))
			srv.logwriter.Errorf("Failed to delete results for pipeline %s , with error %s", pipelineid, err)
			return
		} else {
			srv.logwriter.Infof("Deleted results for pipeline %s", pipelineid)
		}

		err = scheduler.GetDefaultPipelineScheduler().DeletePipelineSchedule(pipelineid)
		srv.SendResponse200(writer, fmt.Sprintf("Successfully deleted pipeline, %s", pipelineid))
	}
}

func (srv *Server) GetPipelineRunResult(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")
	pipelineId := request.URL.Query().Get("pipelineid")

	if !primitive.IsValidObjectID(pipelineId) {
		srv.logwriter.Warnf(fmt.Sprintf("Invalid Pipeline ID: %s, received in get result query", pipelineId))
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid Pipeline ID: %s", pipelineId)))
		return
	}

	sweepaccountid := request.Header.Get(AccountIDHeader)
	if !primitive.IsValidObjectID(sweepaccountid) {
		srv.SendResponse400(writer, errors.New(fmt.Sprintf("Invalid Customer(Sweeper) Account ID: %s", sweepaccountid)))
		return
	}

	query := fmt.Sprintf(`{"pipelineid":"%s", "sweepaccountid": "%s"}`, pipelineId, sweepaccountid)

	pipelineResults, err := srv.opr.PipeLineOperator.GetPipelineResultDetails(query)

	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}

	//TODO when length >1

	if len(pipelineResults) == 0 {

		srv.SendResponse404(writer, nil)
		return
	}

	writer.WriteHeader(http.StatusOK)
	//policieResult := pipelineResults[0]
	json.NewEncoder(writer).Encode(pipelineResults)
}
