package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

const contextChangesTestPath = "/v1/agents/agent-owned-codex/assignment-context"

func TestContextChangesHTTPReturnsMetadataAndRecordsActualTitle(t *testing.T) {
	f, previous := contextChangesHTTPFixture(t)
	code, body := f.do(contextChangesTestPath, contextChangesRequest(f, previous.AssignmentID, f.leased.LeaseToken), f.device)
	if code != http.StatusOK {
		t.Fatalf("status=%d body=%s", code, body)
	}
	var response struct {
		Data    AssignmentContextChanges
		Sources []AssignmentContextSource
	}
	if json.Unmarshal([]byte(body), &response) != nil || len(response.Data.Changes) != 1 || !response.Data.Complete || response.Data.FullRefreshRequired {
		t.Fatal("invalid comparison", body)
	}
	if response.Data.SnapshotID != f.leased.ID || response.Data.Since != previous.AssignmentID || response.Data.Changes[0].Status != "updated" {
		t.Fatal("wrong comparison identity", body)
	}
	if strings.Contains(body, "PRIVATE-BODY") || strings.Contains(body, "previous-thread") {
		t.Fatal("old or new content leaked", body)
	}
	if len(response.Sources) != 1 || response.Sources[0].Kind != "task" || response.Sources[0].Part != "title" {
		t.Fatal("wrong actual title audit", response.Sources)
	}
	reads, err := f.contexts.ListAssignmentContextReads(t.Context(), f.leased.ID)
	if err != nil || reads[len(reads)-1].SourceKind != "task" || reads[len(reads)-1].SourcePart != "title" {
		t.Fatal("wrong persisted audit", err)
	}
}

func TestContextChangesHTTPPreservesExecutionBoundaries(t *testing.T) {
	f, previous := contextChangesHTTPFixture(t)
	request := contextChangesRequest(f, previous.AssignmentID, f.leased.LeaseToken)
	if code, _ := f.do(contextChangesTestPath, request, nil); code != http.StatusUnauthorized {
		t.Fatal("missing device accepted", code)
	}
	if code, _ := f.do(contextChangesTestPath, contextChangesRequest(f, previous.AssignmentID, "wrong"), f.device); code != http.StatusForbidden {
		t.Fatal("wrong lease accepted", code)
	}
	f.event(AssignmentReady, EventAssignmentReady)
	f.event(AssignmentRunning, EventAssignmentRunning)
	f.event(AssignmentCompleted, EventAssignmentCompleted)
	if code, _ := f.do(contextChangesTestPath, request, f.device); code != http.StatusConflict {
		t.Fatal("terminal run accepted", code)
	}
}

func TestContextChangesHTTPConsumesCurrentRunQuota(t *testing.T) {
	f, previous := contextChangesHTTPFixture(t)
	f.contexts.readAttempts = map[string]int{f.leased.ID: assignmentContextReadLimit - 1}
	request := contextChangesRequest(f, previous.AssignmentID, f.leased.LeaseToken)
	if code, body := f.do(contextChangesTestPath, request, f.device); code != http.StatusOK {
		t.Fatal(code, body)
	}
	if code, _ := f.do(contextChangesTestPath, request, f.device); code != http.StatusTooManyRequests {
		t.Fatal("quota bypass", code)
	}
}
