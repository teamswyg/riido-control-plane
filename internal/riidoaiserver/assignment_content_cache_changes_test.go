package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestAssignmentContentCacheRepresentationChanges(t *testing.T) {
	for _, name := range []string{"body", "sources", "captured-at", "truncated", "cursor"} {
		t.Run(name, func(t *testing.T) {
			f := newAssignmentContentHTTPFixture(t)
			req := contentCacheRequest(f)
			code, cold := readContentCache(t, f, req, "1", "", nil)
			if code != http.StatusOK || cold.ContentDigest == "" {
				t.Fatal("initial read failed", code)
			}
			f.contexts.mu.Lock()
			snapshot := f.contexts.snapshots[f.leased.ID]
			switch name {
			case "body":
				f.contexts.content[f.leased.ID][0].Data = json.RawMessage(`{"content":"cd"}`)
			case "sources":
				f.contexts.content[f.leased.ID][0].Sources[0].Version = "current-evidence"
			case "captured-at":
				snapshot.CapturedAt = snapshot.CapturedAt.Add(time.Second)
			case "truncated":
				snapshot.Truncated = !snapshot.Truncated
			case "cursor":
				req.Cursor = "last"
			}
			f.contexts.snapshots[f.leased.ID] = snapshot
			f.contexts.mu.Unlock()
			code, changed := readContentCache(t, f, req, "1", cold.ContentDigest, nil)
			if code != http.StatusOK || changed.NotModified || changed.ContentDigest == "" || changed.ContentDigest == cold.ContentDigest || string(changed.Data) == "null" {
				t.Fatal("changed current representation reused candidate", code)
			}
			code, hit := readContentCache(t, f, req, "1", changed.ContentDigest, nil)
			if code != http.StatusOK || !hit.NotModified {
				t.Fatal("new representation could not be reused", code)
			}
			if name == "sources" {
				reads, err := f.contexts.ListAssignmentContextReads(t.Context(), f.leased.ID)
				if err != nil || reads[len(reads)-1].SourceVersion != "current-evidence" {
					t.Fatal("hit did not audit current page sources", err)
				}
			}
		})
	}
}

func TestAssignmentContentCacheBindsIdenticalBatchToByteBudget(t *testing.T) {
	f := newAssignmentContentHTTPFixture(t)
	req := contentCacheRequest(f)
	req.MaxBytes = 8 << 10
	code, first := readContentCache(t, f, req, "1", "", nil)
	if code != http.StatusOK || first.ContentDigest == "" {
		t.Fatal("initial batch failed", code)
	}
	req.MaxBytes = 16 << 10
	code, second := readContentCache(t, f, req, "1", first.ContentDigest, nil)
	if code != http.StatusOK || second.NotModified || second.ContentDigest == first.ContentDigest || !bytes.Equal(first.Data, second.Data) {
		t.Fatal("identical batch was not bound to requested byte budget", code)
	}
}
