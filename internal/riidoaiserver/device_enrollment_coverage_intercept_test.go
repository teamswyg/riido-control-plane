package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func coverageEnrollmentIntercept(t *testing.T, fixture *membershipSharedDynamoFixture, mode string) func(http.ResponseWriter, *http.Request, []byte) bool {
	t.Helper()
	var consumed atomic.Bool
	return func(w http.ResponseWriter, r *http.Request, body []byte) bool {
		if !coverageBirthRequest(r, body) {
			return false
		}
		if mode == "reject" {
			w.WriteHeader(http.StatusServiceUnavailable)
			return true
		}
		if !consumed.CompareAndSwap(false, true) {
			return false
		}
		recorder := httptest.NewRecorder()
		fixture.handle(t, recorder, r)
		if recorder.Code != http.StatusOK {
			t.Error("fixture initial transaction failed")
			w.WriteHeader(recorder.Code)
			return true
		}
		if mode == "lost_ack_missing_policy" {
			fixture.mu.Lock()
			for key, item := range fixture.items {
				if dynamoDBStringValue(item, "policy_schema") == executionCoverageSchema {
					delete(fixture.items, key)
				}
			}
			fixture.mu.Unlock()
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		return true
	}
}
