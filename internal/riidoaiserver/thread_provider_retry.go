package riidoaiserver

// Only a terminal recovery decision establishes that preflight stopped. An
// intermediate checking/backoff receipt can precede a successful dispatch.
func confirmedCodexPreStartFailure(p AssignmentProjection, req *AssignRequest) bool {
	r := p.Recovery
	if r == nil || len(r.Receipts) == 0 || p.Assignment.RuntimeProvider != "codex" || req.RuntimeProvider != "codex" {
		return false
	}
	c := r.Context
	if c.Provider != "codex" || c.OutcomeCertainty != "not_started" || (c.State != "unavailable" && c.State != "user_action_required") {
		return false
	}
	last := r.Receipts[len(r.Receipts)-1]
	return (last.State == AssignmentReady || last.State == AssignmentLeased) && last.Seq > 0 && last.Seq < p.LastEventSeq
}
