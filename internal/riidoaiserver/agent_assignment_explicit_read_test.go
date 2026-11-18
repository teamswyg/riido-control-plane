package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func explicitReadFixture(t *testing.T, change func(*assignmentContextFixture, *explicitAssignmentResult, http.ResponseWriter, *http.Request) bool) *assignmentContextFixture {
	t.Helper()
	f := newAssignmentContextFixture(t, NewMemoryAssignmentContextStore())
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get(CommentDeliveryKeyHeader) != strings.Repeat("k", 32) || r.URL.Path != "/internal/agent-assignment-read/"+f.leased.WorkspaceID+"/"+f.leased.TaskID+"/"+f.leased.CreatedBy {
			t.Error("untrusted source request")
		}
		team := "team-one"
		v := explicitAssignmentResult{SchemaVersion: "riido-agent-explicit-read.v1", WorkspaceID: f.leased.WorkspaceID, RootID: f.leased.TaskID, PrincipalID: f.leased.CreatedBy, ID: "target-one", TeamID: &team, Title: "Current title", Kind: "task", Data: json.RawMessage(`{"id":"target-one","body":{"content":"readable evidence"}}`)}
		if change != nil && !change(f, &v, w, r) {
			return
		}
		_ = json.NewEncoder(w).Encode(v)
	}))
	t.Cleanup(remote.Close)
	f.service.config.AssignmentReadScope = true
	f.service.config.CommentDelivery.Key = strings.Repeat("k", 32)
	f.service.config.CommentDelivery.EvidenceBaseURL = remote.URL
	f.server = f.service.Handler()
	return f
}

func TestExplicitAssignmentReadsCurrentSourceAndRecords(t *testing.T) {
	f := explicitReadFixture(t, nil)
	code, body := f.read("task", "RAX-43", f.leased.LeaseToken)
	var out AgentAssignmentContextResponse
	if code != 200 || json.Unmarshal([]byte(body), &out) != nil || !out.Live || out.TaskID != f.leased.TaskID || len(out.Sources) != 1 || out.Sources[0].ID != "target-one" || out.Sources[0].Part != "metadata" {
		t.Fatalf("status=%d body=%s", code, body)
	}
	reads, _ := f.contexts.ListAssignmentContextReads(t.Context(), f.leased.ID)
	last := reads[len(reads)-1]
	if last.RequesterPrincipalID != f.leased.CreatedBy || last.Status != AssignmentContextReadServed || last.SourceVersion == "" {
		t.Fatal("missing original requester audit")
	}
}
