package controlplane

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestControlPlaneOwnerGraphQLBehaviorGolden(t *testing.T) {
	useCase := NewSourceReadyUseCase()
	status, err := useCase.HealthCheck(context.Background())
	if err != nil || status != http.StatusOK {
		t.Fatalf("health status=%d err=%v", status, err)
	}
	if err := useCase.FireError(context.Background()); !errors.Is(err, errControlPlaneOwnerFire) {
		t.Fatalf("fireError returned %v", err)
	}
}

func TestControlPlaneOwnerHealthCheckReturnsExactPublicResponse(t *testing.T) {
	spy := healthyControlPlaneOwnerSpy()
	request := httptest.NewRequest(http.MethodPost, "/owner/graphql", strings.NewReader(
		`{"query":"query ControlPlaneOwnerHealthCheck { healthCheck }","operationName":"ControlPlaneOwnerHealthCheck","variables":{}}`,
	))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	NewGraphQLHandler(spy).ServeHTTP(response, request)

	if response.Code != http.StatusOK || response.Body.String() != "{\"data\":{\"healthCheck\":200}}\n" {
		t.Fatalf("unexpected response status=%d body=%q", response.Code, response.Body.String())
	}
	if spy.healthCalls != 1 || spy.fireCalls != 0 {
		t.Fatalf("unexpected calls health=%d fire=%d", spy.healthCalls, spy.fireCalls)
	}
}

func TestControlPlaneOwnerFireErrorNeverInventsSuccess(t *testing.T) {
	spy := healthyControlPlaneOwnerSpy()
	spy.fireErr = nil
	request := httptest.NewRequest(http.MethodPost, "/owner/graphql", strings.NewReader(
		`{"query":"query ControlPlaneOwnerFireError { fireError }","operationName":"ControlPlaneOwnerFireError","variables":null}`,
	))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	NewGraphQLHandler(spy).ServeHTTP(response, request)

	want := "{\"data\":{\"fireError\":null},\"errors\":[{\"message\":\"control-plane fireError always fails\"}]}\n"
	if response.Code != http.StatusOK || response.Body.String() != want {
		t.Fatalf("unexpected response status=%d body=%q", response.Code, response.Body.String())
	}
	if spy.healthCalls != 0 || spy.fireCalls != 1 {
		t.Fatalf("unexpected calls health=%d fire=%d", spy.healthCalls, spy.fireCalls)
	}
}
