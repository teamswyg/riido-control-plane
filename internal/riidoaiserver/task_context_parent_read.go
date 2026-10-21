package riidoaiserver

import (
	"context"
	"sync"
)

func (c *AIAgentPrivateTaskContextClient) enrichParentTaskContext(ctx context.Context, teamID, componentID, bearer string, snapshot *AIAgentTaskContext) {
	seed := c.inheritParentTaskHierarchy(ctx, teamID, componentID, bearer, snapshot)
	groups := make(map[string][]*AIAgentTaskContextReference)
	for _, ref := range []*AIAgentTaskContextReference{&snapshot.Hierarchy.Project, &snapshot.Hierarchy.Milestone, &snapshot.Hierarchy.ParentTask} {
		if ref.ID == "" || ref.ID == componentID {
			continue
		}
		ref.DocumentStatus = "unavailable"
		groups[ref.ID] = append(groups[ref.ID], ref)
	}
	var wg sync.WaitGroup
	for id, refs := range groups {
		wg.Go(func() {
			detail, found := seed[id]
			if !found {
				var err error
				detail, err = c.getComponentDetail(ctx, teamID, id, bearer)
				if err != nil {
					return
				}
			}
			if detail.ID != id {
				return
			}
			parent := detail.toAIAgentTaskContext()
			parent.Document = selectTaskContextDocument(parent.Document, c.readProviderProjection(ctx, teamID, id, bearer))
			resources, status := c.readTaskContextResources(ctx, teamID, id, bearer)
			for _, ref := range refs {
				if parent.Document.Status != "unavailable" {
					document := parent.Document
					ref.Document = &document
					ref.DocumentStatus = parent.Document.Status
				}
				ref.Resources, ref.ResourceStatus = resources, status
			}
		})
	}
	wg.Wait()
}
