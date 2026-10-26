package riidoaiserver

import (
	"context"
	"errors"
	"strings"
	"time"
)

var errAgentDirectoryRequesterUnknown = errors.New("assignment requester is unknown")

// assignmentRequesterPrincipal scopes live agent reads to the person who
// started the run: public agents plus the requester's own, or every agent in
// the workspace when the requester was an admin at request time. Roles are
// not on the assignment, so without that recorded fact admin is not assumed.
func assignmentRequesterPrincipal(assignment Assignment, requesterAdmin bool) (AuthorizationResult, error) {
	requester := strings.TrimSpace(assignment.CreatedBy)
	if requester == "" {
		return AuthorizationResult{}, errAgentDirectoryRequesterUnknown
	}
	principal := AuthorizationResult{PrincipalID: requester, WorkspaceID: strings.TrimSpace(assignment.WorkspaceID)}
	if requesterAdmin {
		principal.Roles = []AgentCatalogRole{AgentCatalogRoleAdmin}
	}
	return principal, nil
}

func (s Server) assignmentAgentDirectory(ctx context.Context, assignment Assignment, requesterAdmin bool) ([]AgentDirectoryEntry, error) {
	principal, err := assignmentRequesterPrincipal(assignment, requesterAdmin)
	if err != nil {
		return nil, err
	}
	if s.aiAgent == nil {
		return nil, errors.New("agent directory is not configured")
	}
	list, err := s.aiAgent.ListAIAgentTaskAssignableAgents(ctx, principal, assignment.TaskID)
	if err != nil {
		return nil, err
	}
	entries := make([]AgentDirectoryEntry, 0, len(list.Agents))
	for _, agent := range list.Agents {
		entries = append(entries, s.agentDirectoryEntry(agent, assignment.AgentID))
	}
	return entries, nil
}

func (s Server) agentDirectoryEntry(agent AgentClientRecord, currentAgentID string) AgentDirectoryEntry {
	entry := AgentDirectoryEntry{
		AgentID: agent.AgentID, Name: strings.TrimSpace(agent.Name), Description: strings.TrimSpace(agent.Description),
		RoleStatus: AgentDirectoryRoleDescribed, WorkStatus: agent.WorkStatus, UpdatedAt: agent.UpdatedAt,
		Runtime: AgentDirectoryRuntime{Kind: agent.RuntimeKind, Connection: AgentDirectoryConnectionOffline},
	}
	if entry.Description == "" {
		entry.RoleStatus = AgentDirectoryRoleMissing
	}
	if agent.AgentID == currentAgentID {
		entry.IsCurrentAgent = true
		entry.UnavailableReasons = append(entry.UnavailableReasons, AgentDirectoryReasonSelf)
	}
	if _, _, err := s.resolveAgentRuntimeFact(agent.AgentID); err == nil {
		entry.Runtime.Connection = AgentDirectoryConnectionOnline
	} else {
		entry.UnavailableReasons = append(entry.UnavailableReasons, AgentDirectoryReasonRuntimeOffline)
	}
	entry.Callable = len(entry.UnavailableReasons) == 0
	return entry
}

func agentDirectorySource(entry AgentDirectoryEntry) AssignmentContextSource {
	source := AssignmentContextSource{Kind: "agent", ID: entry.AgentID, Part: "profile", Version: assignmentContextVersion(entry.Name + "\n" + entry.Description)}
	if !entry.UpdatedAt.IsZero() {
		source.UpdatedAt = entry.UpdatedAt.UTC().Format(time.RFC3339)
	}
	return source
}
