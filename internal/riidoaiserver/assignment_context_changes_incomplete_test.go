package riidoaiserver

import (
	"strings"
	"testing"
)

func TestContextChangesIncompleteAndUnknownAreNotUnchanged(t *testing.T) {
	part := AssignmentContentManifest{Part: AssignmentContentEditor, Status: "available", Complete: true, Digest: "sha256:" + strings.Repeat("a", 64)}
	resource := AssignmentScopedResource{Kind: AssignmentContentComponent, ID: "work", Title: "Current"}
	if _, changed := assignmentPartChange(resource, part, part, true); changed {
		t.Fatal("same complete content changed")
	}
	for _, mutate := range []func(*AssignmentContentManifest){
		func(p *AssignmentContentManifest) { p.Complete = false },
		func(p *AssignmentContentManifest) { p.Digest = "" },
		func(p *AssignmentContentManifest) { p.Digest = "invalid" },
		func(p *AssignmentContentManifest) { p.Digest = "sha256:" + strings.Repeat("z", 64) },
		func(p *AssignmentContentManifest) { p.Status = "unavailable" },
	} {
		bad := part
		mutate(&bad)
		for _, pair := range [][2]AssignmentContentManifest{{part, bad}, {bad, part}} {
			change, changed := assignmentPartChange(resource, pair[0], pair[1], true)
			if !changed || change.Status != "unavailable" {
				t.Fatal("incomplete comparison reported unchanged or updated")
			}
		}
	}
}
