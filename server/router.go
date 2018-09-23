package server

import (
	"net/http"

	"github.com/alexanderbez/titan/core"
	"github.com/alexanderbez/titan/manager"
	"github.com/gorilla/mux"
)

func (srvr *Server) createRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/executions/latest", srvr.GetLatestExecution()).Methods("GET")

	return router
}

// GetLatestExecution returns the latest monitor execution.
func (srvr *Server) GetLatestExecution() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		value, err := srvr.db.Get(core.BadgerMonitorsNamespace, manager.MonitorExecKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(value)
	}
}
