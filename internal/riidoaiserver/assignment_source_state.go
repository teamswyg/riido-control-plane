package riidoaiserver

import (
	"context"
	"log"
)

// AssignmentSourceState is comparison metadata decided once at capture time,
// so repeated change reads in one run agree and do not call the monolith.
type AssignmentSourceState struct {
	ResourceKind AssignmentContentKind `json:"resource_kind"`
	SourceID     string                `json:"source_id"`
	State        string                `json:"state"`
}

type agentSourceStateRef struct {
	Kind   AssignmentContentKind `json:"kind"`
	ID     string                `json:"id"`
	TeamID string                `json:"team_id"`
}

// assignmentSourceStates asks only about sources of the server-selected
// previous capture in the same conversation that this capture no longer has.
// Every failure returns nil: the comparison then keeps the unknown omission.
func (s Server) assignmentSourceStates(ctx context.Context, snapshot AssignmentContextSnapshot) []AssignmentSourceState {
	config, store := s.config.AgentSourceState, s.config.AssignmentContext
	if !config.enabled() || store == nil || snapshot.PreviousAssignmentID == "" || snapshot.PreviousThreadID == "" || snapshot.ConversationID == "" ||
		snapshot.Context.ReadScope == nil || snapshot.WorkspaceID == "" || snapshot.RequesterPrincipalID == "" {
		return nil
	}
	parent, found, err := store.LoadAssignmentContextSnapshot(ctx, snapshot.PreviousAssignmentID)
	if err != nil || !found || parent.AssignmentID != snapshot.PreviousAssignmentID || parent.ThreadID != snapshot.PreviousThreadID ||
		parent.ConversationID != snapshot.ConversationID || parent.Context.ReadScope == nil || parent.TriggerKind == AssignmentContextTriggerAgentMention ||
		!sameAssignmentContextScope(snapshot, parent) {
		return nil
	}
	refs := absentAssignmentSources(parent, snapshot)
	if len(refs) == 0 || len(refs) > assignmentSourceStateLimit {
		return nil
	}
	states, err := config.lookup(ctx, snapshot.WorkspaceID, snapshot.RequesterPrincipalID, refs)
	if err != nil {
		log.Printf("event=assignment_source_state_failed assignment_id=%q sources=%d error=%q", snapshot.AssignmentID, len(refs), err.Error())
		return nil
	}
	return states
}

func absentAssignmentSources(parent, current AssignmentContextSnapshot) []agentSourceStateRef {
	present := map[string]bool{}
	for _, resource := range current.Context.ReadScope.Resources {
		present[string(resource.Kind)+"\x00"+resource.ID] = true
	}
	refs := []agentSourceStateRef{}
	for _, resource := range parent.Context.ReadScope.Resources {
		key := string(resource.Kind) + "\x00" + resource.ID
		if present[key] || resource.ID == "" || resource.TeamID == "" || !validAssignmentContentKind(resource.Kind) {
			continue
		}
		present[key] = true
		refs = append(refs, agentSourceStateRef{Kind: resource.Kind, ID: resource.ID, TeamID: resource.TeamID})
	}
	return refs
}
