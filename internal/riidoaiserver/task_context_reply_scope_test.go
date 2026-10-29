package riidoaiserver

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"
)

func TestCurrentReplyMetadataDoesNotExpandAssignmentContext(t *testing.T) {
	var calls atomic.Int32
	upstream := viewerReplyContextServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		http.Error(w, "unexpected reply read", http.StatusBadRequest)
	})
	client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: upstream.URL})
	got, err := client.GetAIAgentTaskContextForRequest(t.Context(), AIAgentTaskContextRequest{ComponentID: "task", WorkspaceID: defaultAIAgentClientWorkspaceID, BearerToken: "viewer-token", Extended: true})
	if err != nil || got.Extended == nil || len(got.Extended.Comments) != 1 || calls.Load() != 0 {
		t.Fatalf("assignment context expanded: err=%v calls=%d", err, calls.Load())
	}
}

func TestCurrentReplyMetadataDoesNotFetchUnverifiedParents(t *testing.T) {
	var calls atomic.Int32
	upstream := viewerReplyContextServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		http.Error(w, "unexpected", http.StatusBadRequest)
	})
	client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: upstream.URL})
	for _, roots := range [][]AIAgentTaskContextComment{nil, {{ID: "parent", ReplyCount: 0}}, {{ID: "reply", ParentID: "unverified-parent", ReplyCount: 1}}} {
		if got := client.readCurrentReplyMetadata(context.Background(), "team", "viewer-token", roots); len(got) != 0 {
			t.Fatal("unverified parent was used")
		}
	}
	if calls.Load() != 0 {
		t.Fatal("unverified parent made an upstream call")
	}
}

func TestCurrentReplyMetadataCapsCombinedReplies(t *testing.T) {
	response := map[string][]privateReplyMetadata{}
	for parent := range 2 {
		alias := fmt.Sprintf("r%d", parent)
		for i := range 150 {
			response[alias] = append(response[alias], privateReplyMetadata{ID: fmt.Sprintf("reply-%d-%d", parent, i), Status: "active"})
		}
	}
	got := currentReplyMetadata([]string{"parent-0", "parent-1"}, nil, response)
	if len(got) != 200 || got[199].ID != "reply-1-49" {
		t.Fatalf("reply result exceeds the shared budget: count=%d", len(got))
	}
}
