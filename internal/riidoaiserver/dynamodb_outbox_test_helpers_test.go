package riidoaiserver

import (
	"net/http"
	"testing"
)

type capturedDynamoDBRequest struct {
	method string
	header http.Header
	body   []byte
}

func mustStaticAWSTestProvider(t *testing.T, accessKeyID, sessionToken string) StaticAWSCredentialsProvider {
	t.Helper()
	provider, err := NewStaticAWSCredentialsProvider(accessKeyID, "SECRET", sessionToken)
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	return provider
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
