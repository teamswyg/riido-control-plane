package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDynamoDBStoreSnapshotNilAndClosedBoundaries(t *testing.T) {
	var nilStore *DynamoDBStoreSnapshot
	var nilContext context.Context
	if snapshot, ok, err := nilStore.LoadStoreSnapshot(nilContext); err != nil || ok || snapshot.SchemaVersion != "" {
		t.Fatalf("nil LoadStoreSnapshot() = (%+v, %v, %v), want zero false nil", snapshot, ok, err)
	}
	if err := nilStore.SaveStoreSnapshot(nilContext, StoreSnapshot{}); err != nil {
		t.Fatalf("nil SaveStoreSnapshot() error = %v, want nil", err)
	}
	if err := nilStore.Close(); err != nil {
		t.Fatalf("nil Close() error = %v, want nil", err)
	}

	store := newDynamoDBStoreSnapshotForBoundary(t, `{}`)
	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
	_, _, loadErr := store.LoadStoreSnapshot(context.Background())
	if !isClosedStoreSnapshotError(loadErr) {
		t.Fatalf("LoadStoreSnapshot closed error = %v", loadErr)
	}
	saveErr := store.SaveStoreSnapshot(context.Background(), StoreSnapshot{SchemaVersion: StoreSnapshotSchemaVersion})
	if !isClosedStoreSnapshotError(saveErr) {
		t.Fatalf("SaveStoreSnapshot closed error = %v", saveErr)
	}
}

func newDynamoDBStoreSnapshotForBoundary(t *testing.T, body string) *DynamoDBStoreSnapshot {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	store, err := NewDynamoDBStoreSnapshot(DynamoDBStoreSnapshotConfig{
		Region:              "ap-northeast-2",
		TableName:           "assignments",
		Endpoint:            server.URL,
		CredentialsProvider: mustStaticAWSTestProvider(t, "AKID", ""),
	})
	if err != nil {
		t.Fatalf("NewDynamoDBStoreSnapshot: %v", err)
	}
	return store
}

func isClosedStoreSnapshotError(err error) bool {
	return err != nil && err.Error() == "riidoaiserver: DynamoDB snapshot store closed"
}
