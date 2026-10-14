package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDoAWSJSONReadsLargeDynamoDBResponse(t *testing.T) {
	largeValue := strings.Repeat("a", 2<<20)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", dynamoDBJSONContentType)
		_, _ = w.Write([]byte(`{"Items":[{"payload":{"S":"` + largeValue + `"}}]}`))
	}))
	defer server.Close()

	body := doAWSJSONForDynamoDBQuery(t, server, awsJSONResponseBodyLimit)
	var response struct {
		Items []map[string]map[string]string `json:"Items"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got := response.Items[0]["payload"]["S"]; len(got) != len(largeValue) {
		t.Fatalf("payload length = %d, want %d", len(got), len(largeValue))
	}
}

func TestDoAWSJSONRejectsOversizedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", dynamoDBJSONContentType)
		_, _ = w.Write([]byte(strings.Repeat("a", awsJSONResponseBodyLimit+1)))
	}))
	defer server.Close()

	_, err := doAWSJSONForDynamoDBQueryErr(t, server, awsJSONResponseBodyLimit)
	if err == nil {
		t.Fatal("expected oversized response error")
	}
	if !strings.Contains(err.Error(), "response body exceeds") {
		t.Fatalf("error = %v", err)
	}
}

func doAWSJSONForDynamoDBQuery(t *testing.T, server *httptest.Server, limit int64) []byte {
	t.Helper()
	body, err := doAWSJSONForDynamoDBQueryErr(t, server, limit)
	if err != nil {
		t.Fatalf("doAWSJSON: %v", err)
	}
	return body
}

func doAWSJSONForDynamoDBQueryErr(t *testing.T, server *httptest.Server, limit int64) ([]byte, error) {
	t.Helper()
	return doAWSJSON(context.Background(), awsJSONRequest{
		endpoint: server.URL, endpointHost: strings.TrimPrefix(server.URL, "http://"),
		region: "ap-northeast-2", service: dynamoDBService, target: dynamoDBQueryTarget,
		contentType: dynamoDBJSONContentType, payload: []byte(`{}`),
		credentials: AWSCredentials{AccessKeyID: "AKID", SecretAccessKey: "SECRET"},
		httpClient:  server.Client(), now: func() time.Time { return time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC) },
	})
}
