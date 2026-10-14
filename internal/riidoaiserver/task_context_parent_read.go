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
			parent.Document.Source = "component_snapshot"
			if document, ok := c.getProviderDocumentHTML(ctx, teamID, id, bearer); ok {
				parent.Document.Content = document.HTML
				parent.Document.ContentFormat = "html"
				parent.Document.Source = "provider_document"
			}
			resources, status := c.readTaskContextResources(ctx, teamID, id, bearer)
			for _, ref := range refs {
				if parent.Document.Content != "" {
					document := parent.Document
					ref.Document = &document
					ref.DocumentStatus = "available"
				}
				ref.Resources, ref.ResourceStatus = resources, status
			}
		})
	}
	wg.Wait()
}
