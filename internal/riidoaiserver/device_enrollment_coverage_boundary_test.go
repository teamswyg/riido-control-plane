package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func coverageEnrollmentBoundary(t *testing.T, fixture *membershipSharedDynamoFixture, intercept func(http.ResponseWriter, *http.Request, []byte) bool) string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
		if intercept(w, r, body) {
			return
		}
		fixture.handle(t, w, r)
	}))
	t.Cleanup(server.Close)
	return server.URL
}

func coverageBirthRequest(r *http.Request, body []byte) bool {
	if r.Header.Get("X-Amz-Target") != dynamoDBTransactWriteTarget {
		return false
	}
	var request struct {
		TransactItems []struct {
			Put *dynamoDBTransactWritePutAction
		}
	}
	if json.Unmarshal(body, &request) != nil {
		return false
	}
	for _, action := range request.TransactItems {
		if action.Put != nil && dynamoDBStringValue(action.Put.Item, "policy_schema") == executionCoverageSchema && action.Put.ConditionExpression == "attribute_not_exists(pk) AND attribute_not_exists(sk)" {
			return true
		}
	}
	return false
}
