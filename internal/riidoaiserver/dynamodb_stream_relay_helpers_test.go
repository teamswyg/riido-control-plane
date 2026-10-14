package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

const streamRelayTestARN = "arn:aws:dynamodb:ap-northeast-2:123456789012:table/riido-ai-server-event-outbox/stream/2026-05-26T03:04:05.000"

func marshalOutboxRecordForStreamTest(t *testing.T, event TaskEvent) string {
	t.Helper()
	record, err := json.Marshal(OutboxRecord{
		SchemaVersion: OutboxRecordSchemaVersion,
		Event:         event,
	})
	if err != nil {
		t.Fatalf("marshal outbox record: %v", err)
	}
	return string(record)
}

func assertStreamRelayRequest(t *testing.T, request capturedDynamoDBRequest, wantTarget string) {
	t.Helper()
	if request.method != http.MethodPost {
		t.Fatalf("method = %s", request.method)
	}
	if request.header.Get("Content-Type") != dynamoDBJSONContentType {
		t.Fatalf("content-type = %q", request.header.Get("Content-Type"))
	}
	if request.header.Get("X-Amz-Target") != wantTarget {
		t.Fatalf("target = %q", request.header.Get("X-Amz-Target"))
	}
	authorization := request.header.Get("Authorization")
	wantPrefix := "AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/20260526/ap-northeast-2/dynamodb/aws4_request"
	if !strings.HasPrefix(authorization, wantPrefix) {
		t.Fatalf("authorization = %q", authorization)
	}
	wantSignedHeaders := "SignedHeaders=content-type;host;x-amz-date;x-amz-security-token;x-amz-target"
	if !strings.Contains(authorization, wantSignedHeaders) {
		t.Fatalf("authorization signed headers = %q", authorization)
	}
}
