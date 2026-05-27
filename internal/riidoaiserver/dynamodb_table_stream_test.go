package riidoaiserver

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLoadDynamoDBTableStreamARN(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 8, 9, 10, 0, time.UTC)
	requests := make(chan capturedDynamoDBRequest, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		requests <- capturedDynamoDBRequest{method: r.Method, header: r.Header.Clone(), body: body}
		_, _ = w.Write([]byte(`{"Table":{"LatestStreamArn":"arn:aws:dynamodb:ap-northeast-2:123456789012:table/riido-ai-server-event-outbox/stream/2026-05-26T08:09:10.000"}}`))
	}))
	defer server.Close()

	provider, err := NewStaticAWSCredentialsProvider("AKIDEXAMPLE", "SECRET", "SESSION")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	arn, err := LoadDynamoDBTableStreamARN(context.Background(), DynamoDBTableStreamARNConfig{
		Region:              "ap-northeast-2",
		TableName:           "riido-ai-server-event-outbox",
		Endpoint:            server.URL,
		CredentialsProvider: provider,
		Now:                 func() time.Time { return fixedNow },
	})
	if err != nil {
		t.Fatalf("LoadDynamoDBTableStreamARN: %v", err)
	}
	if !strings.Contains(arn, "riido-ai-server-event-outbox/stream") {
		t.Fatalf("arn = %q", arn)
	}

	got := <-requests
	assertDynamoDBTarget(t, got, dynamoDBDescribeTableTarget)
	var payload struct {
		TableName string `json:"TableName"`
	}
	if err := json.Unmarshal(got.body, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload.TableName != "riido-ai-server-event-outbox" {
		t.Fatalf("table name = %q", payload.TableName)
	}
}

func TestDescribeDynamoDBTableCapturesBillingAndKeySchema(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"Table":{"TableName":"riido-ai-server-event-outbox","TableStatus":"ACTIVE","LatestStreamArn":"arn:aws:dynamodb:ap-northeast-2:123456789012:table/riido-ai-server-event-outbox/stream/2026-05-26T08:09:10.000","BillingModeSummary":{"BillingMode":"PAY_PER_REQUEST"},"KeySchema":[{"AttributeName":"task_id","KeyType":"HASH"},{"AttributeName":"event_seq","KeyType":"RANGE"}]}}`))
	}))
	defer server.Close()

	provider, err := NewStaticAWSCredentialsProvider("AKID", "SECRET", "")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	table, err := DescribeDynamoDBTable(context.Background(), DynamoDBTableStreamARNConfig{
		Region:              "ap-northeast-2",
		TableName:           "riido-ai-server-event-outbox",
		Endpoint:            server.URL,
		CredentialsProvider: provider,
	})
	if err != nil {
		t.Fatalf("DescribeDynamoDBTable: %v", err)
	}
	if table.BillingMode != "PAY_PER_REQUEST" || table.TableStatus != "ACTIVE" || len(table.KeySchema) != 2 {
		t.Fatalf("table = %+v", table)
	}
	if table.KeySchema[0].AttributeName != "task_id" || table.KeySchema[0].KeyType != "HASH" {
		t.Fatalf("key schema = %+v", table.KeySchema)
	}
}

func TestLoadDynamoDBTableStreamARNRejectsTableWithoutStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"Table":{}}`))
	}))
	defer server.Close()

	provider, err := NewStaticAWSCredentialsProvider("AKID", "SECRET", "")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	_, err = LoadDynamoDBTableStreamARN(context.Background(), DynamoDBTableStreamARNConfig{
		Region:              "ap-northeast-2",
		TableName:           "events",
		Endpoint:            server.URL,
		CredentialsProvider: provider,
	})
	if err == nil || !strings.Contains(err.Error(), "LatestStreamArn") {
		t.Fatalf("expected missing stream arn error, got %v", err)
	}
}
