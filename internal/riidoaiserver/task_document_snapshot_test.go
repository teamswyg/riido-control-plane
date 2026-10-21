package riidoaiserver

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestTaskDocumentSnapshotOverridesOnlyVerifiedBody(t *testing.T) {
	for _, body := range []string{"<p>click-time body</p>", ""} {
		t.Run(body, func(t *testing.T) {
			reader := &assignmentHTTPRequestTaskContextReader{contextSnapshot: AIAgentTaskContext{
				Component: AIAgentTaskContextComponent{ID: "task", Title: "Verified title"},
				Document:  AIAgentTaskContextDocument{ID: "document", Content: "stale body", Source: "provider_document"},
			}}
			s := NewServer(ServerConfig{TaskContext: reader})
			composed, err := s.assignRequestWithTaskContextPromptForClientResult(context.Background(), "task", AssignRequest{}, AIAgentTaskContextRequest{
				ComponentID: "task", WorkspaceID: "workspace", BearerToken: "jwt",
				DocumentSnapshot: &AIAgentTaskDocumentSnapshot{Content: body, ContentFormat: "html"},
			})
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(composed.Request.Prompt, "stale body") || !strings.Contains(composed.Request.Prompt, "Verified title") || !strings.Contains(composed.Request.Prompt, "editor_snapshot") {
				t.Fatalf("snapshot prompt = %s", composed.Request.Prompt)
			}
			if body != "" && !strings.Contains(composed.Request.Prompt, "click-time body") {
				t.Fatal("fresh body omitted")
			}
			if len(reader.requests) != 1 || reader.requests[0].WorkspaceID != "workspace" || reader.requests[0].BearerToken != "jwt" {
				t.Fatal("request authorization scope lost")
			}
		})
	}
}

func TestTaskDocumentSnapshotFailsClosedWithoutVerifiedContext(t *testing.T) {
	s := NewServer(ServerConfig{TaskContext: &assignmentHTTPRequestTaskContextReader{err: errors.New("workspace mismatch")}})
	_, err := s.assignRequestWithTaskContextPromptForClientResult(context.Background(), "task", AssignRequest{}, AIAgentTaskContextRequest{
		ComponentID: "task", DocumentSnapshot: &AIAgentTaskDocumentSnapshot{Content: "private", ContentFormat: "html"},
	})
	if err == nil {
		t.Fatal("unverified snapshot must not execute")
	}
}

func TestTaskDocumentSnapshotValidation(t *testing.T) {
	for _, snapshot := range []*AIAgentTaskDocumentSnapshot{
		{ContentFormat: "text"},
		{Content: strings.Repeat("한", 87382), ContentFormat: "html"},
		{Content: string([]byte{0xff}), ContentFormat: "html"},
	} {
		if err := snapshot.validate(); err == nil {
			t.Fatalf("accepted invalid snapshot: format=%q size=%d", snapshot.ContentFormat, len(snapshot.Content))
		}
	}
	if err := (&AIAgentTaskDocumentSnapshot{Content: strings.Repeat("x", 256*1024), ContentFormat: "html"}).validate(); err != nil {
		t.Fatal(err)
	}
}
