package controlplane

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestControlPlaneOwnerReceiverRejectsUnregisteredRequestsWithoutCalls(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"alias", `{"query":"query ControlPlaneOwnerHealthCheck { alias: healthCheck }","operationName":"ControlPlaneOwnerHealthCheck"}`},
		{"repeated", `{"query":"query ControlPlaneOwnerHealthCheck { healthCheck ` + `health` + `Check }","operationName":"ControlPlaneOwnerHealthCheck"}`},
		{"fragment", `{"query":"query ControlPlaneOwnerHealthCheck { ...Health } fragment Health on Query { healthCheck }","operationName":"ControlPlaneOwnerHealthCheck"}`},
		{"unknown", `{"query":"query Unknown { healthCheck }","operationName":"Unknown"}`},
		{"variables", `{"query":"query ControlPlaneOwnerHealthCheck { healthCheck }","operationName":"ControlPlaneOwnerHealthCheck","variables":{"unused":1}}`},
		{"duplicate-json", `{"query":"query ControlPlaneOwnerHealthCheck { healthCheck }","query":"query ControlPlaneOwnerHealthCheck { healthCheck }","operationName":"ControlPlaneOwnerHealthCheck"}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			spy := healthyControlPlaneOwnerSpy()
			request := httptest.NewRequest(http.MethodPost, "/owner/graphql", strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			NewGraphQLHandler(spy).ServeHTTP(response, request)
			if response.Code != http.StatusUnprocessableEntity || spy.healthCalls != 0 || spy.fireCalls != 0 {
				t.Fatalf("status=%d calls=%d/%d", response.Code, spy.healthCalls, spy.fireCalls)
			}
		})
	}
}

func TestControlPlaneOwnerReceiverFailsClosedBeforeUseCase(t *testing.T) {
	valid := `{"query":"query ControlPlaneOwnerHealthCheck { healthCheck }","operationName":"ControlPlaneOwnerHealthCheck"}`
	tests := []struct {
		name        string
		handler     http.Handler
		method      string
		contentType string
		wantStatus  int
	}{
		{"provider-absent", NewGraphQLHandler(nil), http.MethodPost, "application/json", http.StatusServiceUnavailable},
		{"wrong-method", NewGraphQLHandler(healthyControlPlaneOwnerSpy()), http.MethodGet, "application/json", http.StatusMethodNotAllowed},
		{"wrong-content-type", NewGraphQLHandler(healthyControlPlaneOwnerSpy()), http.MethodPost, "text/plain", http.StatusUnsupportedMediaType},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, "/owner/graphql", strings.NewReader(valid))
			request.Header.Set("Content-Type", test.contentType)
			response := httptest.NewRecorder()
			test.handler.ServeHTTP(response, request)
			if response.Code != test.wantStatus {
				t.Fatalf("status=%d want=%d", response.Code, test.wantStatus)
			}
		})
	}
}
