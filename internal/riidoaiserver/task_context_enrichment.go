package riidoaiserver

import (
	"context"
	"sync"
)

func (c *AIAgentPrivateTaskContextClient) enrichTaskContext(ctx context.Context, teamID, componentID, bearer string, snapshot *AIAgentTaskContext, editor *AIAgentTaskDocumentSnapshot) error {
	var wg sync.WaitGroup
	var documentErr error
	wg.Go(func() {
		var reads sync.WaitGroup
		reads.Go(func() {
			if editor != nil {
				editor.apply(&snapshot.Document)
				return
			}
			document, err := c.getLatestProviderDocument(ctx, teamID, componentID, bearer)
			if err != nil {
				documentErr = err
				return
			}
			snapshot.Document = selectTaskContextDocument(snapshot.Document, document)
		})
		reads.Go(func() {
			snapshot.Resources, snapshot.ResourceStatus = c.readTaskContextResources(ctx, teamID, componentID, bearer)
		})
		reads.Wait()
	})
	wg.Go(func() { c.enrichParentTaskContext(ctx, teamID, componentID, bearer, snapshot) })
	wg.Wait()
	return documentErr
}
