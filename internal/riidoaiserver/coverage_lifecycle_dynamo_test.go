package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// One table and one transaction lock cover enrollment, claims, links and proof.
// Only the AWS transport is synthetic; adapters/handlers and CAS payloads are real.
type coverageLifecycleDynamo struct {
	mu                sync.Mutex
	items             map[string]map[string]map[string]string
	requests          map[string]int
	table             string
	beforePolicyFence func()
}

type coverageLifecycleAction struct {
	TableName                 string
	Key, Item                 map[string]map[string]string
	ConditionExpression       string
	ExpressionAttributeValues map[string]map[string]string
	ExpressionAttributeNames  map[string]string
}

func newCoverageLifecycleDynamo(t *testing.T) (*coverageLifecycleDynamo, string) {
	t.Helper()
	f := &coverageLifecycleDynamo{items: map[string]map[string]map[string]string{}, requests: map[string]int{}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { f.handle(t, w, r) }))
	t.Cleanup(server.Close)
	return f, server.URL
}

func (f *coverageLifecycleDynamo) sameTable(t *testing.T, table string) bool {
	t.Helper()
	if table == "" {
		t.Error("missing Dynamo table")
		return false
	}
	if f.table == "" {
		f.table = table
	}
	if f.table != table {
		t.Error("lifecycle crossed Dynamo table authorities")
		return false
	}
	return true
}
