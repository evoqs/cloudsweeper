package api

import (
	"cloudsweep/storage"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
	//dbM  storage.DBManger
	opr    storage.DbOperators
	socket string
}

func (srv *Server) StartApiServer(socket string, dbO storage.DbOperators) {
	srv.socket = socket
	//srv.dbM = dbM
	srv.opr = dbO
	router := mux.NewRouter()

	//SweepAccount Level operations
	router.HandleFunc("/accounts/{accountid}", srv.GetAllCloudAccount).Methods("GET")
	router.HandleFunc("/accounts/{accountid}", srv.DeleteAllCloudAccount).Methods("DELETE")
	router.HandleFunc("/accounts/{accountid}/pipeline", srv.GetAllPipeLine).Methods("GET")

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
	router.HandleFunc("/policy/{policyid}/results", srv.GetPolicyRunResult).Methods("GET")

	//Run a pipeline
	router.HandleFunc("/pipeline/{pipelineid}/run", srv.RunPipeLine).Methods("POST")
	router.HandleFunc("/pipeline", srv.AddPipeLine).Methods("POST")
	router.HandleFunc("/pipeline/{pipelineid}", srv.GetPipeLine).Methods("GET")
	router.HandleFunc("/pipeline/{pipelineid}", srv.DeletePipeLine).Methods("DELETE")

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
