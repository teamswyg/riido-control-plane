package riidoaiserver

import (
	"context"
	"strconv"
	"strings"
	"time"
)

// assignmentAgentDirectoryView answers `riido agents list|show`. ok is false
// when `show` names an agent outside the requester's visible directory;
// absent and hidden agents are refused alike so existence is not revealed.
func (s Server) assignmentAgentDirectoryView(ctx context.Context, assignment Assignment, requesterAdmin bool, req AgentAssignmentContextRequest, now time.Time) (any, []AssignmentContextSource, bool, error) {
	entries, err := s.assignmentAgentDirectory(ctx, assignment, requesterAdmin)
	if err != nil {
		return nil, nil, false, err
	}
	if req.View == AssignmentContextViewAgent {
		target := strings.TrimSpace(req.Target)
		for _, entry := range entries {
			if target != "" && entry.AgentID == target {
				return AgentDirectoryDetail{Agent: entry, FetchedAt: now}, []AssignmentContextSource{agentDirectorySource(entry)}, true, nil
			}
		}
		return nil, nil, false, nil
	}
	page, err := strconv.Atoi(strings.TrimSpace(req.Cursor))
	if err != nil || page < 0 {
		page = 0
	}
	out := AgentDirectoryPage{Agents: []AgentDirectoryEntry{}, Total: len(entries), Cursor: strconv.Itoa(page), FetchedAt: now}
	start := page * agentDirectoryPageSize
	if start < len(entries) {
		end := min(start+agentDirectoryPageSize, len(entries))
		out.Agents = entries[start:end]
		if end < len(entries) {
			out.NextCursor = strconv.Itoa(page + 1)
		}
	}
	sources := make([]AssignmentContextSource, 0, len(out.Agents))
	for _, entry := range out.Agents {
		sources = append(sources, agentDirectorySource(entry))
	}
	return out, sources, true, nil
}
