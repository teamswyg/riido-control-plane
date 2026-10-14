package riidoaiserver

import (
	"net/http"
	"strings"
	"testing"
)

func assertDynamoDBOutboxPutRequest(t *testing.T, got capturedDynamoDBRequest) {
	t.Helper()
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
	wantPrefix := "AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/20260526/ap-northeast-2/dynamodb/aws4_request"
	if !strings.HasPrefix(authorization, wantPrefix) {
		t.Fatalf("authorization = %q", authorization)
	}
	wantSignedHeaders := "SignedHeaders=content-type;host;x-amz-date;x-amz-security-token;x-amz-target"
	if !strings.Contains(authorization, wantSignedHeaders) {
		t.Fatalf("authorization signed headers = %q", authorization)
	}
}
