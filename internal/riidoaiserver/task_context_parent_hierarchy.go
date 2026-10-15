package riidoaiserver

import "context"

func (c *AIAgentPrivateTaskContextClient) inheritParentTaskHierarchy(ctx context.Context, teamID, componentID, bearer string, snapshot *AIAgentTaskContext) map[string]privateComponentDetailResponse {
	seed := make(map[string]privateComponentDetailResponse)
	h := &snapshot.Hierarchy
	if h.ParentTask.ID == "" || h.ParentTask.ID == componentID || (h.Project.ID != "" && h.Milestone.ID != "") {
		return seed
	}
	detail, err := c.getComponentDetail(ctx, teamID, h.ParentTask.ID, bearer)
	seed[h.ParentTask.ID] = detail
	if err != nil || detail.ID != h.ParentTask.ID {
		return seed
	}
	parent := detail.toAIAgentTaskContext().Hierarchy
	if h.Project.ID == "" {
		h.Project = parent.Project
	}
	if h.Milestone.ID == "" {
		h.Milestone = parent.Milestone
	}
	return seed
}
