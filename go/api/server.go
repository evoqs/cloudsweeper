package api

import (
	"cloudsweep/storage"
	"encoding/json"
	"net/http"

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

	router.HandleFunc("/accounts/{accountid}", srv.GetAllCloudAccount).Methods("GET")
	router.HandleFunc("/accounts/{accountid}", srv.DeleteAllCloudAccount).Methods("DELETE")

	router.HandleFunc("/cloudaccount", srv.AddCloudAccount).Methods("POST")
	router.HandleFunc("/cloudaccount", srv.UpdateCloudAccount).Methods("PUT")
	router.HandleFunc("/cloudaccount/{cloudaccountid}", srv.GetCloudAccount).Methods("GET")
	router.HandleFunc("/cloudaccount/{cloudaccountid}", srv.DeleteCloudAccount).Methods("DELETE")

	http.ListenAndServe(":8000", router)
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
