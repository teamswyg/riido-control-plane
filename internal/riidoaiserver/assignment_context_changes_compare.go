package riidoaiserver

import (
	"encoding/hex"
	"strings"
)

func compareAssignmentContextParts(baseline, current AssignmentContextSnapshot) AssignmentContextChanges {
	out := AssignmentContextChanges{SnapshotID: current.AssignmentID, Since: baseline.AssignmentID, Changes: []AssignmentContextChange{}, Complete: true}
	previous := map[string]AssignmentContentManifest{}
	for _, part := range baseline.Content {
		previous[readScopeManifestKey(part)] = part
	}
	present := map[string]bool{}
	for _, resource := range current.Context.ReadScope.Resources {
		for _, part := range resource.Parts {
			key := readScopeManifestKey(part)
			present[key] = true
			prior, existed := previous[key]
			change, changed := assignmentPartChange(resource, part, prior, existed)
			if changed {
				out.Changes = append(out.Changes, change)
			}
			if changed && change.Status == "unavailable" {
				out.FullRefreshRequired = true
			}
		}
	}
	for _, part := range baseline.Content {
		if present[readScopeManifestKey(part)] {
			continue
		}
		out.Changes = append(out.Changes, AssignmentContextChange{ResourceKind: part.ResourceKind, SourceID: part.SourceID, Part: part.Part, Status: "unavailable", Reason: "not_in_current_capture"})
		out.FullRefreshRequired = true
	}
	if current.Truncated || baseline.Truncated || current.Context.ReadScope.Status != "available" || baseline.Context.ReadScope.Status != "available" {
		out.FullRefreshRequired = true
	}
	out.Complete = !out.FullRefreshRequired
	if out.FullRefreshRequired {
		out.NextAction = "riido context get"
	}
	return out
}

func assignmentPartChange(resource AssignmentScopedResource, part, prior AssignmentContentManifest, existed bool) (AssignmentContextChange, bool) {
	change := AssignmentContextChange{ResourceKind: part.ResourceKind, SourceID: part.SourceID, Part: part.Part, Title: resource.Title, TeamID: resource.TeamID, KeyNumber: resource.KeyNumber, Reasons: resource.Reasons, Current: &part, Status: "updated"}
	if !assignmentPartComparable(part) || (existed && !assignmentPartComparable(prior)) {
		change.Status, change.Reason = "unavailable", "incomplete_comparison"
		return change, true
	}
	return change, !existed || part.Digest != prior.Digest
}

func assignmentPartComparable(part AssignmentContentManifest) bool {
	if !part.Complete || (part.Status != "available" && part.Status != "empty") || !strings.HasPrefix(part.Digest, "sha256:") || len(part.Digest) != 71 {
		return false
	}
	_, err := hex.DecodeString(part.Digest[7:])
	return err == nil
}
