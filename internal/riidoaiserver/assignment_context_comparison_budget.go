package riidoaiserver

import (
	"encoding/json"
	"slices"
)

// Optional comparison metadata must never make an otherwise readable capture
// disappear. Keep the existing storage limit and fall back to a full read.
func boundAssignmentComparisonMetadata(snapshot AssignmentContextSnapshot, pages []AssignmentContentPage) AssignmentContextSnapshot {
	_, refs, err := packAssignmentContentPages(pages)
	if err != nil {
		return snapshot
	}
	probe := snapshot
	probe.ContentBlocks = refs
	raw, err := json.Marshal(probe)
	if err != nil {
		return snapshot
	}
	item, encodeErr := assignmentContentSnapshotItem(snapshot, refs)
	if len(raw) <= contextSnapshotItemBytes-128 && encodeErr == nil && validateContextPayloadBudget([]map[string]map[string]string{item}) == nil {
		return snapshot
	}
	clearAssignmentContextLineage(&snapshot)
	snapshot.Content = withoutAssignmentContentDigests(snapshot.Content)
	if snapshot.Context.ReadScope != nil {
		scope := *snapshot.Context.ReadScope
		scope.Resources = slices.Clone(scope.Resources)
		for i := range scope.Resources {
			scope.Resources[i].Parts = withoutAssignmentContentDigests(scope.Resources[i].Parts)
		}
		snapshot.Context.ReadScope = &scope
	}
	return snapshot
}

func withoutAssignmentContentDigests(parts []AssignmentContentManifest) []AssignmentContentManifest {
	out := slices.Clone(parts)
	for i := range out {
		out[i].Digest = ""
	}
	return out
}
