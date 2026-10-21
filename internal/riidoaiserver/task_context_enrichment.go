package riidoaiserver

import (
	"context"
	"sync"
)

func (c *AIAgentPrivateTaskContextClient) enrichTaskContext(ctx context.Context, teamID, componentID, bearer string, snapshot *AIAgentTaskContext) {
	var wg sync.WaitGroup
	wg.Go(func() {
		var document *AIAgentTaskContextDocument
		var resources []AIAgentTaskContextResource
		var status string
		var reads sync.WaitGroup
		reads.Go(func() { document = c.readProviderProjection(ctx, teamID, componentID, bearer) })
		reads.Go(func() { resources, status = c.readTaskContextResources(ctx, teamID, componentID, bearer) })
		reads.Wait()
		snapshot.Document = selectTaskContextDocument(snapshot.Document, document)
		snapshot.Resources, snapshot.ResourceStatus = resources, status
	})
	wg.Go(func() { c.enrichParentTaskContext(ctx, teamID, componentID, bearer, snapshot) })
	wg.Wait()
}
