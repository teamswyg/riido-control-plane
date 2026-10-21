package riidoaiserver

import (
	"context"
	"sync"
)

func (c *AIAgentPrivateTaskContextClient) enrichTaskContext(ctx context.Context, teamID, componentID, bearer string, snapshot *AIAgentTaskContext, editor *AIAgentTaskDocumentSnapshot) error {
	if snapshot.Document.Content != "" {
		snapshot.Document.Source = "component_snapshot"
	}
	var wg sync.WaitGroup
	var documentErr error
	wg.Go(func() {
		if editor != nil {
			editor.apply(&snapshot.Document)
			return
		}
		document, err := c.getLatestProviderDocument(ctx, teamID, componentID, bearer)
		if err != nil {
			documentErr = err
			return
		}
		snapshot.Document.Content, snapshot.Document.ContentFormat, snapshot.Document.Source = document.Content, document.ContentFormat, document.Source
	})
	wg.Go(func() {
		snapshot.Resources, snapshot.ResourceStatus = c.readTaskContextResources(ctx, teamID, componentID, bearer)
	})
	wg.Go(func() { c.enrichParentTaskContext(ctx, teamID, componentID, bearer, snapshot) })
	wg.Wait()
	return documentErr
}
