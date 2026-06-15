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

func TestDynamoDBOutboxWritesPutItem(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC)
	requests := make(chan capturedDynamoDBRequest, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		requests <- capturedDynamoDBRequest{
			method: r.Method,
			header: r.Header.Clone(),
			body:   body,
		}
		w.Header().Set("Content-Type", "application/x-amz-json-1.0")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	provider, err := NewStaticAWSCredentialsProvider("AKIDEXAMPLE", "SECRET", "SESSION")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	outbox, err := NewDynamoDBOutbox(DynamoDBOutboxConfig{
		Region:              "ap-northeast-2",
		TableName:           "riido-ai-server-event-outbox",
		Endpoint:            server.URL,
		CredentialsProvider: provider,
		Now:                 func() time.Time { return fixedNow },
	})
	if err != nil {
		t.Fatalf("NewDynamoDBOutbox: %v", err)
	}
	defer outbox.Close()

	event := TaskEvent{
		Seq:          7,
		TaskID:       "task-a",
		AssignmentID: "assignment-1",
		AgentID:      "jykim1",
		Type:         EventAssignmentLeased,
		State:        AssignmentLeased,
		Message:      "leased",
		Metadata:     map[string]string{"lease_token": "lease-1"},
		At:           fixedNow,
	}
	if err := outbox.AppendTaskEvent(context.Background(), event); err != nil {
		t.Fatalf("AppendTaskEvent: %v", err)
	}

	got := <-requests
	if got.method != http.MethodPost {
		t.Fatalf("method = %s", got.method)
	}
	if got.header.Get("Content-Type") != dynamoDBJSONContentType {
		t.Fatalf("content-type = %q", got.header.Get("Content-Type"))
	}
	if got.header.Get("X-Amz-Target") != dynamoDBPutItemTarget {
		t.Fatalf("target = %q", got.header.Get("X-Amz-Target"))
	}
	if got.header.Get("X-Amz-Date") != "20260526T010203Z" {
		t.Fatalf("x-amz-date = %q", got.header.Get("X-Amz-Date"))
	}
	if got.header.Get("X-Amz-Security-Token") != "SESSION" {
		t.Fatalf("session token = %q", got.header.Get("X-Amz-Security-Token"))
	}
	authorization := got.header.Get("Authorization")
	if !strings.HasPrefix(authorization, "AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/20260526/ap-northeast-2/dynamodb/aws4_request") {
		t.Fatalf("authorization = %q", authorization)
	}
	if !strings.Contains(authorization, "SignedHeaders=content-type;host;x-amz-date;x-amz-security-token;x-amz-target") {
		t.Fatalf("authorization signed headers = %q", authorization)
	}

	var payload struct {
		TableName           string                       `json:"TableName"`
		ConditionExpression string                       `json:"ConditionExpression"`
		Item                map[string]map[string]string `json:"Item"`
	}
	if err := json.Unmarshal(got.body, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload.TableName != "riido-ai-server-event-outbox" {
		t.Fatalf("table = %q", payload.TableName)
	}
	if payload.ConditionExpression != "attribute_not_exists(task_id) AND attribute_not_exists(event_seq)" {
		t.Fatalf("condition = %q", payload.ConditionExpression)
	}
	assertDynamoDBString(t, payload.Item, "task_id", "task-a")
	assertDynamoDBNumber(t, payload.Item, "event_seq", "7")
	assertDynamoDBString(t, payload.Item, "assignment_id", "assignment-1")
	assertDynamoDBString(t, payload.Item, "agent_id", "jykim1")
	assertDynamoDBString(t, payload.Item, "event_type", EventAssignmentLeased)
	assertDynamoDBString(t, payload.Item, "assignment_state", string(AssignmentLeased))
	assertDynamoDBString(t, payload.Item, "message", "leased")
	assertDynamoDBString(t, payload.Item, "metadata_json", `{"lease_token":"lease-1"}`)
	assertDynamoDBString(t, payload.Item, "schema_version", OutboxRecordSchemaVersion)
	assertDynamoDBString(t, payload.Item, "at", "2026-05-26T01:02:03Z")

	var record OutboxRecord
	if err := json.Unmarshal([]byte(payload.Item["event_json"]["S"]), &record); err != nil {
		t.Fatalf("decode event_json: %v", err)
	}
	if record.SchemaVersion != OutboxRecordSchemaVersion || record.Event.TaskID != "task-a" || record.Event.Type != EventAssignmentLeased {
		t.Fatalf("record = %+v", record)
	}
}

func TestDynamoDBOutboxTreatsConditionalCheckFailedAsIdempotent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"__type":"com.amazonaws.dynamodb.v20120810#ConditionalCheckFailedException","message":"duplicate"}`))
	}))
	defer server.Close()

	provider, err := NewStaticAWSCredentialsProvider("AKID", "SECRET", "")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	outbox, err := NewDynamoDBOutbox(DynamoDBOutboxConfig{
		Region:              "ap-northeast-2",
		TableName:           "events",
		Endpoint:            server.URL,
		CredentialsProvider: provider,
	})
	if err != nil {
		t.Fatalf("NewDynamoDBOutbox: %v", err)
	}
	defer outbox.Close()

	err = outbox.AppendTaskEvent(context.Background(), TaskEvent{
		Seq:    1,
		TaskID: "task-a",
		Type:   EventAssignmentQueued,
		At:     time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("AppendTaskEvent duplicate: %v", err)
	}
}

func TestDoAWSJSONReadsLargeDynamoDBResponse(t *testing.T) {
	largeValue := strings.Repeat("a", 2<<20)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", dynamoDBJSONContentType)
		_, _ = w.Write([]byte(`{"Items":[{"payload":{"S":"` + largeValue + `"}}]}`))
	}))
	defer server.Close()

	body, err := doAWSJSON(context.Background(), awsJSONRequest{
		endpoint:     server.URL,
		endpointHost: strings.TrimPrefix(server.URL, "http://"),
		region:       "ap-northeast-2",
		service:      dynamoDBService,
		target:       dynamoDBQueryTarget,
		contentType:  dynamoDBJSONContentType,
		payload:      []byte(`{}`),
		credentials:  AWSCredentials{AccessKeyID: "AKID", SecretAccessKey: "SECRET"},
		httpClient:   server.Client(),
		now:          func() time.Time { return time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC) },
	})
	if err != nil {
		t.Fatalf("doAWSJSON: %v", err)
	}
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

	_, err := doAWSJSON(context.Background(), awsJSONRequest{
		endpoint:     server.URL,
		endpointHost: strings.TrimPrefix(server.URL, "http://"),
		region:       "ap-northeast-2",
		service:      dynamoDBService,
		target:       dynamoDBQueryTarget,
		contentType:  dynamoDBJSONContentType,
		payload:      []byte(`{}`),
		credentials:  AWSCredentials{AccessKeyID: "AKID", SecretAccessKey: "SECRET"},
		httpClient:   server.Client(),
		now:          func() time.Time { return time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC) },
	})
	if err == nil {
		t.Fatal("expected oversized response error")
	}
	if !strings.Contains(err.Error(), "response body exceeds") {
		t.Fatalf("error = %v", err)
	}
}

func TestECSContainerCredentialsProviderFetchesCredentials(t *testing.T) {
	requests := make(chan http.Header, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r.Header.Clone()
		_, _ = w.Write([]byte(`{
			"AccessKeyId": "ASIAEXAMPLE",
			"SecretAccessKey": "SECRET",
			"Token": "TOKEN",
			"Expiration": "2026-05-26T01:02:03Z"
		}`))
	}))
	defer server.Close()

	provider, err := NewECSContainerCredentialsProvider(ECSContainerCredentialsProviderConfig{
		Endpoint:           server.URL,
		AuthorizationToken: "Bearer metadata-token",
	})
	if err != nil {
		t.Fatalf("NewECSContainerCredentialsProvider: %v", err)
	}
	credentials, err := provider.Credentials(context.Background())
	if err != nil {
		t.Fatalf("Credentials: %v", err)
	}
	if credentials.AccessKeyID != "ASIAEXAMPLE" || credentials.SecretAccessKey != "SECRET" || credentials.SessionToken != "TOKEN" {
		t.Fatalf("credentials = %+v", credentials)
	}
	if credentials.ExpiresAt.Format(time.RFC3339) != "2026-05-26T01:02:03Z" {
		t.Fatalf("expires_at = %s", credentials.ExpiresAt.Format(time.RFC3339))
	}
	gotHeader := <-requests
	if gotHeader.Get("Authorization") != "Bearer metadata-token" {
		t.Fatalf("authorization = %q", gotHeader.Get("Authorization"))
	}
}

func TestDynamoDBOutboxRejectsInvalidConfig(t *testing.T) {
	provider, err := NewStaticAWSCredentialsProvider("AKID", "SECRET", "")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	if _, err := NewDynamoDBOutbox(DynamoDBOutboxConfig{Region: "ap-northeast-2", CredentialsProvider: provider}); err == nil {
		t.Fatal("expected missing table error")
	}
	if _, err := NewDynamoDBOutbox(DynamoDBOutboxConfig{Region: "ap-northeast-2", TableName: "events"}); err == nil {
		t.Fatal("expected missing credentials provider error")
	}
	if _, err := NewStaticAWSCredentialsProvider("AKID", "", ""); err == nil {
		t.Fatal("expected missing secret key error")
	}
}

type capturedDynamoDBRequest struct {
	method string
	header http.Header
	body   []byte
}

func assertDynamoDBString(t *testing.T, item map[string]map[string]string, key, want string) {
	t.Helper()
	if item[key]["S"] != want {
		t.Fatalf("%s.S = %q", key, item[key]["S"])
	}
}

func assertDynamoDBNumber(t *testing.T, item map[string]map[string]string, key, want string) {
	t.Helper()
	if item[key]["N"] != want {
		t.Fatalf("%s.N = %q", key, item[key]["N"])
	}
}
