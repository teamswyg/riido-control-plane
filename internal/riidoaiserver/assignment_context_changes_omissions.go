package riidoaiserver

// appendAssignmentOmissions reports baseline parts missing from the current
// capture without any old title or body. Only a state recorded at capture time
// explains an omission; a source the requester can no longer read is left out
// without a reason, and anything else stays unknown and needs a full read.
func appendAssignmentOmissions(out *AssignmentContextChanges, baseline AssignmentContextSnapshot, present map[string]bool, states []AssignmentSourceState) {
	recorded := map[string]string{}
	for _, state := range states {
		switch state.State {
		case AssignmentSourceDeleted, AssignmentSourceUnselected, AssignmentSourceHidden:
			recorded[string(state.ResourceKind)+"\x00"+state.SourceID] = state.State
		}
	}
	for _, part := range baseline.Content {
		if present[readScopeManifestKey(part)] {
			continue
		}
		change := AssignmentContextChange{ResourceKind: part.ResourceKind, SourceID: part.SourceID, Part: part.Part, Status: "unavailable", Reason: "not_in_current_capture"}
		switch recorded[string(part.ResourceKind)+"\x00"+part.SourceID] {
		case AssignmentSourceHidden:
			continue
		case AssignmentSourceDeleted:
			change.Status, change.Reason = AssignmentSourceDeleted, ""
		case AssignmentSourceUnselected:
			change.Status, change.Reason = AssignmentSourceUnselected, ""
		default:
			out.FullRefreshRequired = true
		}
		out.Changes = append(out.Changes, change)
	}
}
