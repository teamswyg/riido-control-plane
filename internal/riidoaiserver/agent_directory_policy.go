package riidoaiserver

import "context"

// AgentMentionPrecheck reports why mentioning targetAgentID from the current
// run would be refused now (for example budget_exhausted, depth_limit, cycle,
// session_conflict). RIID-5892 owns the policy; the directory only shows it.
type AgentMentionPrecheck func(ctx context.Context, current Assignment, targetAgentID string) ([]string, error)

// applyAgentMentionPolicy adds RIID-5892 reasons to entries when a precheck is
// configured and reports whether the policy was applied. A failing precheck
// fails the read rather than presenting unchecked agents as checked.
func (s Server) applyAgentMentionPolicy(ctx context.Context, current Assignment, entries []AgentDirectoryEntry) (bool, error) {
	precheck := s.config.AgentMentionPrecheck
	if precheck == nil {
		return false, nil
	}
	for i := range entries {
		if entries[i].IsCurrentAgent {
			continue
		}
		reasons, err := precheck(ctx, current, entries[i].AgentID)
		if err != nil {
			return false, err
		}
		entries[i].UnavailableReasons = append(entries[i].UnavailableReasons, reasons...)
		entries[i].Callable = len(entries[i].UnavailableReasons) == 0
	}
	return true, nil
}
