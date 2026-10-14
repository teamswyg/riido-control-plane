package main

import (
	"net/http"
	"testing"
)

func TestLoadOpenAPIGeneratedCountsGeneratedHTTPMethods(t *testing.T) {
	repo, m := writeSmokeFixtureRepo(t)
	ops, counts, err := loadOpenAPIGenerated(repo + "/" + m.OpenAPI)
	if err != nil {
		t.Fatal(err)
	}
	if counts != m.OperationCounts || ops["v2.bar"].Method != http.MethodPost {
		t.Fatalf("unexpected generated operation result: %+v %+v", counts, ops)
	}
	if isHTTPMethod("parameters") || !isHTTPMethod("trace") {
		t.Fatalf("HTTP method classifier drifted")
	}
}
