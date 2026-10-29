package riidoaiserver

import "context"

func materializeCommentReadScope(ctx context.Context, evidence verifiedCommentDelivery) (AIAgentTaskContext, error) {
	bundle := evidence.Context.ReadScopeEvidence
	if bundle == nil || bundle.Scope == nil || evidence.TaskID == "" || evidence.PrincipalID == "" || evidence.Context.Component.ID != evidence.TaskID {
		return AIAgentTaskContext{}, taskContextReadError{"invalid"}
	}
	team := bundle.Scope.TeamID
	if evidence.Context.TeamID != "" && evidence.Context.TeamID != team {
		return AIAgentTaskContext{}, taskContextReadError{"invalid"}
	}
	index, err := newTaskReadEvidenceIndex(bundle, evidence.TaskID, team)
	if err != nil {
		return AIAgentTaskContext{}, err
	}
	client := &AIAgentPrivateTaskContextClient{readScopeEvidence: index}
	out, err := client.getTaskReadScope(ctx, AIAgentTaskContextRequest{ComponentID: evidence.TaskID}, team)
	if err == nil && !index.complete {
		out.ReadScope.Status = "partial"
	}
	return out, err
}
