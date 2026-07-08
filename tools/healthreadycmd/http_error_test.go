package main

import "testing"

func TestVerifyEndpointsRejectsNonJSONRoutes(t *testing.T) {
	endpoints := []endpointContract{
		{Name: "missing", Method: "GET", Path: "/not-json", Status: "ok", HTTPStatus: 200},
	}
	if _, err := verifyEndpoints(endpoints); err == nil {
		t.Fatal("expected decode error for unmatched route")
	}
}

func TestVerifyEndpointsRecordsWrongMethodResponse(t *testing.T) {
	endpoints := []endpointContract{
		{Name: "health", Method: "POST", Path: "/healthz", Status: "", HTTPStatus: 405},
	}
	results, err := verifyEndpoints(endpoints)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].HTTPStatus != 405 {
		t.Fatalf("results = %+v", results)
	}
}
