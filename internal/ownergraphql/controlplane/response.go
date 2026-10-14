package controlplane

import (
	"encoding/json"
	"net/http"
)

type graphQLError struct {
	Message string `json:"message"`
}

type failureResponse struct {
	Errors []graphQLError `json:"errors"`
}

type healthData struct {
	HealthCheck int `json:"healthCheck"`
}

type healthResponse struct {
	Data healthData `json:"data"`
}

type fireData struct {
	FireError *string `json:"fireError"`
}

type fireResponse struct {
	Data   fireData       `json:"data"`
	Errors []graphQLError `json:"errors"`
}

func writeControlPlaneOwnerFailure(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/graphql-response+json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(failureResponse{Errors: []graphQLError{{Message: message}}}); err != nil {
		return
	}
}

func writeResponse(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/graphql-response+json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		return
	}
}
