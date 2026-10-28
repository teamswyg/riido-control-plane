package riidoaiserver

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestExecutionClaimAuthenticatedHTTPToDurableObservation(t *testing.T) {
	f := newContextRetentionFixture(t)
	record := sampleQueuedAssignmentOperationRecord(f.now)
	record.Assignment.CreatedBy = "requester"
	item := sampleAssignmentProjectionDynamoDBItem(t, record.Assignment, 1)
	f.fake.put(item)
	f.queryPages = []contextRetentionQueryPage{{Items: []map[string]map[string]string{item}}}
	actor := NewStoreWithConfig(StoreConfig{OperationStore: f.store, Now: func() time.Time { return f.now }})
	defer actor.Close()
	devices := NewDevelopmentAIAgentClientStore()
	devices.deviceCredentials["device"] = deviceCredentialRecord{deviceID: "device", secretHash: sha256.Sum256([]byte("fixture-secret")), ownerPrincipalID: "owner"}
	server := NewServer(ServerConfig{AIAgentClient: devices, Assignment: actor}).Handler()
	req := httptest.NewRequest(http.MethodPost, "/v1/agents/"+record.AgentID+"/poll", strings.NewReader(`{"device_id":"device","wait_ms":1}`))
	req.Header.Set(deviceIDHeader, "device")
	req.Header.Set(deviceSecretHeader, "fixture-secret")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	var poll PollResponse
	if resp.Code != http.StatusOK || json.Unmarshal(resp.Body.Bytes(), &poll) != nil || poll.Action != PollStart {
		t.Fatalf("claim status=%d", resp.Code)
	}
	claims, err := f.store.loadDeviceExecutionClaims(context.Background(), "device")
	if err != nil || len(claims) != 1 || claims[0].RequesterPrincipalID != "requester" || claims[0].OwnerPrincipalID != "owner" {
		t.Fatalf("durable provenance missing: count=%d err=%v", len(claims), err)
	}
	observed, err := actor.observeDeviceExecutions(context.Background(), "device")
	if err != nil || observed != deviceExecutionBlocked {
		t.Fatalf("observation=%s err=%v", observed, err)
	}
	// Later device ownership changes must not rewrite the claim-time owner.
	devices.deviceCredentials["device"] = deviceCredentialRecord{deviceID: "device", secretHash: sha256.Sum256([]byte("replacement")), ownerPrincipalID: "requester"}
	observed, err = actor.observeDeviceExecutions(context.Background(), "device")
	if err != nil || observed != deviceExecutionBlocked {
		t.Fatalf("owner change bypassed history: %s %v", observed, err)
	}
	f.now = f.now.Add(time.Minute)
	observed, err = actor.observeDeviceExecutions(context.Background(), "device")
	if err != nil || observed != deviceExecutionUnknown {
		t.Fatalf("expired lease must stay unknown: %s %v", observed, err)
	}
	// The scheduler may mark a stale assignment terminal. This is no execution-end proof.
	ctx := context.WithValue(context.Background(), authenticatedClaimDeviceKey{}, authenticatedClaimDevice{deviceID: "device", ownerPrincipalID: "owner"})
	if _, err := actor.PollAgent(ctx, record.AgentID, PollRequest{}); err != nil {
		t.Fatal(err)
	}
	observed, err = actor.observeDeviceExecutions(context.Background(), "device")
	claims, loadErr := f.store.loadDeviceExecutionClaims(context.Background(), "device")
	if err != nil || loadErr != nil || observed != deviceExecutionUnknown || len(claims) != 1 {
		t.Fatal("stale repair discarded execution history")
	}
}
