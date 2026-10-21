package riidoaiserver

import (
	"context"
	"strings"
	"testing"
)

func TestEditorSnapshotPreservesParentBundlesAndResources(t *testing.T) {
	fixture := newContextBundleFixture(t, true)
	client, err := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: fixture.rest.URL, GraphQLURL: fixture.graph.URL})
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := client.GetAIAgentTaskContextForRequest(context.Background(), AIAgentTaskContextRequest{
		ComponentID: "task", WorkspaceID: "workspace", BearerToken: "SYNTHETIC_VIEWER",
		DocumentSnapshot: &AIAgentTaskDocumentSnapshot{Content: "<p>Click-time editor body</p>", ContentFormat: "html"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Document.Source != "editor_snapshot" || snapshot.Document.Content != "<p>Click-time editor body</p>" {
		t.Fatalf("editor body lost: %+v", snapshot.Document)
	}
	if snapshot.Hierarchy.Project.Document == nil || !strings.Contains(snapshot.Hierarchy.Project.Document.Content, "STRUCTURED_BODY") || snapshot.ResourceStatus != "available" {
		t.Fatalf("parent/resource enrichment lost: %+v", snapshot)
	}
	fixture.mu.Lock()
	defer fixture.mu.Unlock()
	if fixture.reads["projection:task"] != 0 || fixture.reads["resources:task"] != 1 || fixture.reads["projection:project"] != 1 {
		t.Fatalf("snapshot skipped unrelated context or read stale current body: %+v", fixture.reads)
	}
}
