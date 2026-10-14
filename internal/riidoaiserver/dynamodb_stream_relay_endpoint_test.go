package riidoaiserver

import (
	"strings"
	"testing"
)

func TestNormalizeDynamoDBStreamEndpointDefault(t *testing.T) {
	t.Parallel()
	gotEndpoint, gotHost, err := normalizeDynamoDBStreamEndpoint("ap-northeast-2", "")
	if err != nil {
		t.Fatalf("normalize default endpoint: %v", err)
	}
	want := "https://streams.dynamodb.ap-northeast-2.amazonaws.com"
	if gotEndpoint != want || gotHost != "streams.dynamodb.ap-northeast-2.amazonaws.com" {
		t.Fatalf("endpoint = %q host = %q, want %q", gotEndpoint, gotHost, want)
	}
}

func TestNormalizeDynamoDBStreamEndpointCustom(t *testing.T) {
	t.Parallel()
	gotEndpoint, gotHost, err := normalizeDynamoDBStreamEndpoint("unused", "http://127.0.0.1:8000")
	if err != nil {
		t.Fatalf("normalize custom endpoint: %v", err)
	}
	if gotEndpoint != "http://127.0.0.1:8000" || gotHost != "127.0.0.1:8000" {
		t.Fatalf("endpoint = %q host = %q", gotEndpoint, gotHost)
	}
}

func TestNormalizeDynamoDBStreamEndpointRejectsUnsafeShapes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		endpoint string
		wantErr  string
	}{
		{"invalid-url", "http://[::1", "parse DynamoDB stream endpoint"},
		{"scheme", "ftp://streams.example.test", "must use http or https"},
		{"host", "https:///streams", "host is required"},
		{"query", "https://streams.example.test?x=1", "must not include query or fragment"},
		{"fragment", "https://streams.example.test#frag", "must not include query or fragment"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, _, err := normalizeDynamoDBStreamEndpoint("ap-northeast-2", tc.endpoint)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("err = %v, want containing %q", err, tc.wantErr)
			}
		})
	}
}
