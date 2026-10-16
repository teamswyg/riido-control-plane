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

func TestDynamoDBLoadDeviceConnectionGrantsReadsOnlyTheGrantsPart(t *testing.T) {
	now := time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)
	snapshot := snapshotTestRecord(now)
	snapshot.DeviceConnectionGrants = []DeviceConnectionGrant{{
		DeviceID: "device-web", PrincipalID: "web-user", WorkspaceID: "workspace-beta", ConnectedAt: now, LastSeenAt: now,
	}}
	items, _, _, err := encodeSplitDynamoDBAIAgentClientSnapshot(snapshot, nil)
	if err != nil {
		t.Fatalf("encode split snapshot: %v", err)
	}
	partSK := dynamoDBAIAgentClientSnapshotPartSK(dynamoDBAIAgentClientSnapshotPartDeviceConnections)
	var part map[string]map[string]string
	for _, item := range items {
		if dynamoDBStringValue(item, "sk") == partSK {
			part = item
		}
	}
	hasPart := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if r.Header.Get("X-Amz-Target") != dynamoDBGetItemTarget || !strings.Contains(string(body), partSK) {
			t.Errorf("request target=%q body=%s", r.Header.Get("X-Amz-Target"), body)
		}
		response := map[string]any{}
		if hasPart {
			response["Item"] = part
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	store := newSnapshotDynamoDBStore(t, now, server.URL, nil)

	grants, err := store.loadDeviceConnectionGrants(context.Background())
	if err != nil || len(grants) != 1 || grants[0].PrincipalID != "web-user" || !grants[0].LastSeenAt.Equal(now) {
		t.Fatalf("grants = %+v, err = %v", grants, err)
	}
	hasPart = false
	if grants, err := store.loadDeviceConnectionGrants(context.Background()); err != nil || len(grants) != 0 {
		t.Fatalf("grants without a saved part = %+v, err = %v", grants, err)
	}
}
