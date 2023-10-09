package api

import (
	"encoding/json"
	"net/http"

	awsutil "cloudsweep/cloud_lib"
)

func (srv *Server) GetAllRegions(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	writer.Header().Set("Content-Type", "application/json")

	regions, err := awsutil.GetAllRegions()

	if err != nil {
		srv.SendResponse500(writer, err)
		return
	}
	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(regions)
}
