package riidoaiserver

import (
	"context"
	"sync"
)

func (c *AIAgentPrivateTaskContextClient) enrichTaskContext(ctx context.Context, teamID, componentID, bearer string, snapshot *AIAgentTaskContext) {
	if snapshot.Document.Content != "" {
		snapshot.Document.Source = "component_snapshot"
	}
	var wg sync.WaitGroup
	wg.Go(func() {
		if document, ok := c.getProviderDocumentHTML(ctx, teamID, componentID, bearer); ok {
			snapshot.Document.Content = document.HTML
			snapshot.Document.ContentFormat = "html"
			snapshot.Document.Source = "provider_document"
		}
	})
	wg.Go(func() {
		snapshot.Resources, snapshot.ResourceStatus = c.readTaskContextResources(ctx, teamID, componentID, bearer)
	})
	wg.Go(func() { c.enrichParentTaskContext(ctx, teamID, componentID, bearer, snapshot) })
	wg.Wait()
}
