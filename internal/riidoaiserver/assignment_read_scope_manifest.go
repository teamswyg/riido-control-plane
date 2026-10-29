package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"errors"
)

func validateAssignmentReadScopeManifest(snapshot AssignmentContextSnapshot) error {
	scope := snapshot.Context.ReadScope
	if scope == nil {
		return nil
	}
	invalid := errors.New("assignment scoped manifest inconsistent")
	if scope.SchemaVersion != taskReadScopeSchema || scope.RootID != snapshot.Context.Component.ID || scope.TeamID == "" || scope.TeamID != snapshot.Context.TeamID {
		return invalid
	}
	index := map[string]AssignmentContentManifest{}
	for _, part := range snapshot.Content {
		key := readScopeManifestKey(part)
		if _, exists := index[key]; exists {
			return invalid
		}
		index[key] = part
	}
	resources := map[string]bool{}
	for _, resource := range scope.Resources {
		key := string(resource.Kind) + "\x00" + resource.ID
		if resource.ID == "" || resource.TeamID == "" || resources[key] {
			return invalid
		}
		if resource.Kind == AssignmentContentComponent && resource.ID == scope.RootID && resource.TeamID != scope.TeamID {
			return invalid
		}
		resources[key] = true
		for _, part := range resource.Parts {
			key := readScopeManifestKey(part)
			stored, ok := index[key]
			if !ok || part.ResourceKind != resource.Kind || part.SourceID != resource.ID {
				return invalid
			}
			left, leftErr := json.Marshal(part)
			right, rightErr := json.Marshal(stored)
			if leftErr != nil || rightErr != nil || !bytes.Equal(left, right) {
				return invalid
			}
			delete(index, key)
		}
	}
	if len(index) != 0 {
		return invalid
	}
	return nil
}

func readScopeManifestKey(part AssignmentContentManifest) string {
	return string(part.ResourceKind) + "\x00" + part.SourceID + "\x00" + string(part.Part)
}
