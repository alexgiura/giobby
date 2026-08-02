package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterSystemRoutes(router *mux.Router) {
	router.HandleFunc("/healthz", healthCheckHandler()).Methods("GET")
	router.HandleFunc("/health", healthCheckHandler()).Methods("GET")
}

func healthCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
