package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

type countingAWSCredentialsProvider struct{ calls atomic.Int32 }

func (p *countingAWSCredentialsProvider) Credentials(context.Context) (AWSCredentials, error) {
	p.calls.Add(1)
	return AWSCredentials{AccessKeyID: "AKID", SecretAccessKey: "SECRET"}, nil
}

func TestDynamoDBLoadDeviceConnectionGrantsReusesCredentialsAcrossPolls(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()
	provider := &countingAWSCredentialsProvider{}
	store, err := NewDynamoDBAIAgentClientSnapshot(DynamoDBAIAgentClientSnapshotConfig{
		Region: "ap-northeast-2", TableName: "riido-ai-agent-development", Endpoint: server.URL,
		CredentialsProvider: provider, Now: func() time.Time { return time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatalf("NewDynamoDBAIAgentClientSnapshot: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	for range 3 {
		if _, err := store.loadDeviceConnectionGrants(context.Background()); err != nil {
			t.Fatalf("load grants: %v", err)
		}
	}
	if got := provider.calls.Load(); got != 1 {
		t.Fatalf("credentials fetched %d times for 3 polls, want 1", got)
	}
}
