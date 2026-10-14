package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDynamoDBAIAgentClientSnapshotNilAndClosedBoundaries(t *testing.T) {
	var nilStore *DynamoDBAIAgentClientSnapshot
	var nilContext context.Context
	if snapshot, ok, err := nilStore.LoadAIAgentClientSnapshot(nilContext); err != nil || ok || snapshot.WorkspaceID != "" {
		t.Fatalf("nil LoadAIAgentClientSnapshot() = (%+v, %v, %v), want zero false nil", snapshot, ok, err)
	}
	if err := nilStore.SaveAIAgentClientSnapshot(nilContext, snapshotTestRecord(fixedSnapshotTestNow())); err != nil {
		t.Fatalf("nil SaveAIAgentClientSnapshot() error = %v, want nil", err)
	}
	if err := nilStore.Close(); err != nil {
		t.Fatalf("nil Close() error = %v, want nil", err)
	}

	fixture := newSnapshotDynamoDBFixture(t, fixedSnapshotTestNow(), nil, nil)
	if err := fixture.store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := fixture.store.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
	_, _, loadErr := fixture.store.LoadAIAgentClientSnapshot(context.Background())
	if !isClosedSnapshotStoreError(loadErr) {
		t.Fatalf("LoadAIAgentClientSnapshot closed error = %v", loadErr)
	}
	saveErr := fixture.store.SaveAIAgentClientSnapshot(context.Background(), AIAgentClientSnapshot{})
	if !isClosedSnapshotStoreError(saveErr) {
		t.Fatalf("SaveAIAgentClientSnapshot closed error = %v", saveErr)
	}
	fixture.server.Close()
}

func TestDynamoDBAIAgentClientSnapshotSequentialPartWritePropagatesPutError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"message":"boom"}`, http.StatusInternalServerError)
	}))
	defer server.Close()
	store := newSnapshotDynamoDBStore(t, fixedSnapshotTestNow(), server.URL, nil)
	defer store.Close()
	err := store.putSnapshotPartItemsSequentially(
		context.Background(),
		[]map[string]map[string]string{{"sk": {"S": dynamoDBAIAgentClientSnapshotPartSK("agents")}}},
		AWSCredentials{AccessKeyID: "AKID", SecretAccessKey: "SECRET"},
	)
	if err == nil || !strings.Contains(err.Error(), "dynamodb save AI Agent client snapshot") {
		t.Fatalf("putSnapshotPartItemsSequentially error = %v", err)
	}
}

func isClosedSnapshotStoreError(err error) bool {
	return err != nil && err.Error() == errDynamoDBAIAgentClientSnapshotClosed().Error()
}
