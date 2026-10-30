package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestAssignmentContentBatchReturnsConsecutivePagesWithinBudget(t *testing.T) {
	f := newAssignmentContentBatchFixture(t)
	status, body := readAssignmentContentBatch(t, f, "body-0", 8<<10)
	if status != http.StatusOK {
		t.Fatalf("status=%d body=%s", status, body)
	}
	var got assignmentContentBatchResponse
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Data.Pages) != 2 || got.Data.Pages[0].Cursor != "body-0" || got.Data.Pages[1].Cursor != "body-1" {
		t.Fatalf("pages=%+v", got.Data.Pages)
	}
	if got.Data.NextCursor != "body-2" || got.Data.Complete || got.Data.SourceID != "task-private" || got.Data.Part != "description" || got.Data.Version != "v1" {
		t.Fatalf("batch=%+v", got.Data)
	}
	if len(got.Sources) != 1 || got.Sources[0].Part != "body" {
		t.Fatalf("sources=%+v", got.Sources)
	}

	status, body = readAssignmentContentBatch(t, f, "body-2", 256<<10)
	if status != http.StatusOK {
		t.Fatalf("status=%d body=%s", status, body)
	}
	got = assignmentContentBatchResponse{}
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatal(err)
	}
	// The batch ends at the part boundary; the next part keeps its own cursor.
	if len(got.Data.Pages) != 4 || got.Data.Pages[3].Cursor != "body-5" || !got.Data.Complete || got.Data.NextCursor != "" {
		t.Fatalf("batch=%+v", got.Data)
	}

	reads, err := f.contexts.ListAssignmentContextReads(t.Context(), f.leased.ID)
	if err != nil {
		t.Fatal(err)
	}
	served := 0
	for _, read := range reads {
		if read.Via == AssignmentContextViaCLI && read.Status == AssignmentContextReadServed && read.SourceID == "task-private" {
			served++
		}
	}
	if served != 2 {
		t.Fatalf("one audit record per batched read, got %d", served)
	}
}

func TestAssignmentContentBatchRejectsInvalidBudget(t *testing.T) {
	f := newAssignmentContentBatchFixture(t)
	for _, maxBytes := range []int{-1, 1024, (256 << 10) + 1} {
		if status, body := readAssignmentContentBatch(t, f, "body-0", maxBytes); status != http.StatusBadRequest {
			t.Fatalf("max_bytes=%d status=%d body=%s", maxBytes, status, body)
		}
	}
	request, _ := json.Marshal(AgentAssignmentContextRequest{AssignmentID: f.leased.ID, LeaseToken: f.leased.LeaseToken, View: "base", MaxBytes: 8 << 10})
	if status, body := f.do("/v1/agents/agent-owned-codex/assignment-context", string(request), f.device); status != http.StatusBadRequest {
		t.Fatalf("non-content max_bytes status=%d body=%s", status, body)
	}
}
