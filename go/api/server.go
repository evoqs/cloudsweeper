package api

import (
	logger "cloudsweep/logging"
	"cloudsweep/storage"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
	logwriter logger.Logger
	opr       storage.DbOperators
	socket    string
}

const AccountIDHeader = "AccountId"

func (srv *Server) StartApiServer(socket string, dbO storage.DbOperators) {
	srv.socket = socket
	srv.logwriter = logger.NewDefaultLogger()
	srv.opr = dbO
	router := mux.NewRouter()

	//SweepAccount Level operations
	router.HandleFunc("/accounts/cloudaccounts", srv.GetAllCloudAccount).Methods("GET")
	router.HandleFunc("/accounts/cloudaccounts", srv.DeleteAllCloudAccount).Methods("DELETE")
	router.HandleFunc("/accounts/pipelines", srv.GetAllPipeLine).Methods("GET")
	router.HandleFunc("/accounts/policies", srv.GetAllPolicies).Methods("GET")

	//Cloud Account operations
	router.HandleFunc("/cloudaccount", srv.AddCloudAccount).Methods("POST")
	router.HandleFunc("/cloudaccount", srv.UpdateCloudAccount).Methods("PUT")
	router.HandleFunc("/cloudaccount/{cloudaccountid}", srv.GetCloudAccount).Methods("GET")
	router.HandleFunc("/cloudaccount/{cloudaccountid}", srv.DeleteCloudAccount).Methods("DELETE")
	router.HandleFunc("/cloudaccount/{cloudaccountid}/authtest", srv.AuthCheckCloudAccount).Methods("POST")

	//Policy related operations
	router.HandleFunc("/policy", srv.AddCustodianPolicy).Methods("POST")
	router.HandleFunc("/policy", srv.UpdateCustodianPolicy).Methods("PUT")
	router.HandleFunc("/policy/{policyid}", srv.GetCustodianPolicy).Methods("GET")
	router.HandleFunc("/policy/{policyid}", srv.DeleteCustodianPolicy).Methods("DELETE")

	//policyResults
	router.HandleFunc("/policyresults", srv.GetPolicyRunResult).Methods("GET")
	//pipelineResults
	router.HandleFunc("/pipelineresults", srv.GetPipelineRunResult).Methods("GET")

	//Default Policy related operations
	router.HandleFunc("/defaultpolicy", srv.GetDefaultCustodianPolicies).Methods("GET")
	router.HandleFunc("/defaultpolicy", srv.AddDefaultCustodianPolicy).Methods("POST")
	router.HandleFunc("/defaultpolicy", srv.UpdateDefaultCustodianPolicy).Methods("PUT")
	router.HandleFunc("/defaultpolicy/{defaultpolicyid}", srv.DeleteDefaultCustodianPolicy).Methods("DELETE") //completed

	//Run a pipeline
	router.HandleFunc("/pipeline/{pipelineid}/run", srv.RunPipeLine).Methods("POST")
	router.HandleFunc("/pipeline", srv.AddPipeLine).Methods("POST")
	router.HandleFunc("/pipeline/{pipelineid}", srv.GetPipeLine).Methods("GET")
	router.HandleFunc("/pipeline/{pipelineid}", srv.DeletePipeLine).Methods("DELETE")
	router.HandleFunc("/pipeline", srv.UpdatePipeLine).Methods("PUT")

	//AWS
	router.HandleFunc("/aws/regions", srv.GetAllRegions).Methods("GET")

	http.ListenAndServe(socket, router)
}

func (srv *Server) SendResponse500(writer http.ResponseWriter, err error) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusInternalServerError)
	resp := getResponse500()
	resp.Error = err.Error()

	json.NewEncoder(writer).Encode(resp)
}

func (srv *Server) SendResponse400(writer http.ResponseWriter, err error) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusBadRequest)
	resp := getResponse400()
	resp.Error = err.Error()

	json.NewEncoder(writer).Encode(resp)
}

func (srv *Server) SendResponse409(writer http.ResponseWriter, err error) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusConflict)
	resp := getResponse409()
	resp.Error = err.Error()

	json.NewEncoder(writer).Encode(resp)
}

func (srv *Server) SendResponse404(writer http.ResponseWriter, err error) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusNotFound)
	resp := getResponse404()
	if err != nil {
		resp.Error = err.Error()
	}

	json.NewEncoder(writer).Encode(resp)
}

func (srv *Server) SendResponse200(writer http.ResponseWriter, msg string) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	resp := getResponse200()
	if msg != "" {
		resp.Status = msg
	}

	json.NewEncoder(writer).Encode(resp)
}

func (srv *Server) SendResponse207(writer http.ResponseWriter, err error) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusMultiStatus)
	resp := getResponse207()
	if err != nil {
		resp.Error = err.Error()
	}

	json.NewEncoder(writer).Encode(resp)
}
