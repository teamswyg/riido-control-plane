package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

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
	if credentials.AccessKeyID != "ASIAEXAMPLE" ||
		credentials.SecretAccessKey != "SECRET" ||
		credentials.SessionToken != "TOKEN" {
		t.Fatalf("credentials = %+v", credentials)
	}
	if credentials.ExpiresAt.Format(time.RFC3339) != "2026-05-26T01:02:03Z" {
		t.Fatalf("expires_at = %s", credentials.ExpiresAt.Format(time.RFC3339))
	}
	if got := (<-requests).Get("Authorization"); got != "Bearer metadata-token" {
		t.Fatalf("authorization = %q", got)
	}
}
