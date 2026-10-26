package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestAssignmentContextCapturedAtAssignAndServedToLeaseHolder(t *testing.T) {
	f := newAssignmentContextFixture(t, NewMemoryAssignmentContextStore())
	snapshot, found, err := f.contexts.LoadAssignmentContextSnapshot(context.Background(), f.leased.ID)
	if err != nil || !found || snapshot.Context.Component.KeyNumber != "RIID-5900" || snapshot.RequesterPrincipalID != "user-1" ||
		snapshot.TriggerKind != AssignmentContextTriggerAssign || snapshot.TriggerID != "req-assign-1" {
		t.Fatalf("snapshot=%+v found=%v err=%v", snapshot, found, err)
	}
	reads, _ := f.contexts.ListAssignmentContextReads(context.Background(), f.leased.ID)
	if got := assignmentContextReadSummary(reads); got != "prompt:task/body,prompt:link/metadata,prompt:project/body,prompt:milestone/title" {
		t.Fatalf("prompt reads = %s", got)
	}
	code, body := f.read(AssignmentContextViewTask, "riid-4800", f.leased.LeaseToken)
	var out AgentAssignmentContextResponse
	if code != http.StatusOK || json.Unmarshal([]byte(body), &out) != nil || !strings.Contains(body, "project body") || len(out.Sources) != 1 || out.Sources[0].Kind != "project" {
		t.Fatalf("task view status=%d body=%s", code, body)
	}
	reads, _ = f.contexts.ListAssignmentContextReads(context.Background(), f.leased.ID)
	last := reads[len(reads)-1]
	if last.Via != AssignmentContextViaCLI || last.Status != AssignmentContextReadServed || last.RequestID != "cli_1" || last.SourceVersion == "" || last.AgentID != "agent-owned-codex" ||
		last.SourceUpdatedAt != "2026-09-30T15:00:00Z" || reads[0].SourceUpdatedAt != "2026-10-25T09:00:00Z" {
		t.Fatalf("cli read = %+v", last)
	}
	if code, body := f.read(AssignmentContextViewTree, "", f.leased.LeaseToken); code != http.StatusOK || strings.Contains(body, "task body") || !strings.Contains(body, "milestone-a") {
		t.Fatalf("tree view status=%d body=%s", code, body)
	}
}

func TestAssignmentContextDisabledByDefault(t *testing.T) {
	f := newAssignmentContextFixture(t, nil)
	if code, _ := f.read(AssignmentContextViewBase, "", f.leased.LeaseToken); code != http.StatusServiceUnavailable {
		t.Fatalf("feature off status=%d", code)
	}
}

func assignmentContextReadSummary(reads []AssignmentContextRead) string {
	parts := make([]string, 0, len(reads))
	for _, read := range reads {
		parts = append(parts, read.Via+":"+read.SourceKind+"/"+read.SourcePart)
	}
	return strings.Join(parts, ",")
}
