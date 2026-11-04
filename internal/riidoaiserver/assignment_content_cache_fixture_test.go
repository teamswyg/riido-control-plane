package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

const (
	contentCacheVersionHeader = "X-Riido-Context-Cache-Version"
	contentCacheKnownHeader   = "X-Riido-Context-Known-Digest"
)

type assignmentContentCacheResponse struct {
	AgentAssignmentContextResponse
	ContentDigest      string          `json:"content_digest"`
	NotModified        bool            `json:"not_modified"`
	Data               json.RawMessage `json:"data"`
	cacheFieldsPresent bool
}

func contentCacheRequest(f *assignmentContextFixture) AgentAssignmentContextRequest {
	return AgentAssignmentContextRequest{AssignmentID: f.leased.ID, LeaseToken: f.leased.LeaseToken, View: "content", Target: "task-private", Cursor: "first", RequestID: "cache-read"}
}

func readContentCache(t *testing.T, f *assignmentContextFixture, req AgentAssignmentContextRequest, version, digest string, extra func(*http.Request)) (int, assignmentContentCacheResponse) {
	t.Helper()
	request, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	code, body := f.do("/v1/agents/agent-owned-codex/assignment-context", string(request), func(r *http.Request) {
		f.device(r)
		if version != "" {
			r.Header.Set(contentCacheVersionHeader, version)
		}
		if digest != "" {
			r.Header.Set(contentCacheKnownHeader, digest)
		}
		if extra != nil {
			extra(r)
		}
	})
	var out assignmentContentCacheResponse
	if err := json.Unmarshal([]byte(body), &out); err != nil {
		t.Fatal("invalid JSON response", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(body), &fields); err != nil {
		t.Fatal(err)
	}
	_, hasDigest := fields["content_digest"]
	_, hasNotModified := fields["not_modified"]
	out.cacheFieldsPresent = hasDigest || hasNotModified
	return code, out
}
