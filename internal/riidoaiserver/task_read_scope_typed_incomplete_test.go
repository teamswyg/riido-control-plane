package riidoaiserver

import (
	"encoding/json"
	"testing"
)

func TestTaskReadScopeTypedIncompleteMetadataPreservesUnverifiedPrefix(t *testing.T) {
	selector, _ := normalizeReadScopePartSelector(AssignmentContentComponent, "metadata", nil)
	raw := `{"details":{"id":"work","title":"not yet complete"}}`
	version := readScopePartVersion(AssignmentContentComponent, "work", "team", "metadata", "riido_metadata", raw, selector)
	first := typedJSONFixture("work", "metadata", version, raw, 0, &selector)
	retainTypedPrefix(first, []byte(raw[:12]))
	work := &taskReadScopeWork{ID: "work", Kind: AssignmentContentComponent, TeamID: "team", ContentPage: first}
	resource := AssignmentScopedResource{Kind: work.Kind, ID: work.ID, TeamID: work.TeamID}
	client := &AIAgentPrivateTaskContextClient{readScopeEvidence: &taskReadEvidenceIndex{complete: false, values: map[string]json.RawMessage{}, selected: map[string]string{"component\x00work": "team"}}}
	builder, _ := newTaskReadScopeBuilder()
	manifest, err := client.captureReadScopeWorkMetadata(t.Context(), "", builder, resource, work)
	if err != nil || manifest.Complete || len(manifest.Cursors) != 1 || work.ContentPage == nil || len(work.Details) != 0 {
		t.Fatalf("unfinished metadata promoted: %+v %v", manifest, err)
	}
	var payload taskReadBodyPage
	if json.Unmarshal(builder.capture.Pages[0].Data, &payload) != nil || payload.Content != raw[:12] || payload.Complete == nil || *payload.Complete || payload.NextOffset == nil || *payload.NextOffset != 12 {
		t.Fatal("prefix or byte continuation changed")
	}
}

func TestTaskReadScopeMetadataRestorationKeepsUnsupportedPartIncomplete(t *testing.T) {
	incomplete, complete := false, true
	work := taskReadScopeWork{ID: "work", Kind: AssignmentContentComponent, TeamID: "team", ContentPage: &taskReadBodyPage{}, PartsStatus: []*taskReadPartStatus{{Name: "github", Status: "unsupported", Complete: &incomplete}, {Name: "details", Status: "deferred", Complete: &incomplete}, {Name: "comments", Status: "empty", Complete: &complete}}}
	raw := []byte(`{"details":{"id":"work","title":"restored"},"links":[],"github":[],"repositories":[],"figma":null,"priorities":[],"checklists":[],"attachments":[],"sprints":[],"origin":null}`)
	if err := applyReadScopeWorkMetadata(&work, raw); err != nil {
		t.Fatal(err)
	}
	if work.PartsStatus[0].Status != "unsupported" || *work.PartsStatus[0].Complete || readScopeMetadataComplete(work.PartsStatus) {
		t.Fatal("unsupported metadata promoted to available")
	}
	if work.PartsStatus[1].Status != "available" || !*work.PartsStatus[1].Complete {
		t.Fatal("deferred complete slot not restored")
	}
}
