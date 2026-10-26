package riidoaiserver

import (
	"context"
	"sync"
)

// enrichTaskContextFor runs the extended reads beside the regular enrichment
// so the flag adds upstream calls but not latency.
func (c *AIAgentPrivateTaskContextClient) enrichTaskContextFor(ctx context.Context, teamID, componentID, bearer string, snapshot *AIAgentTaskContext, req AIAgentTaskContextRequest) error {
	var extended sync.WaitGroup
	if req.Extended {
		extended.Go(func() { snapshot.Extended = c.readExtendedTaskContext(ctx, teamID, componentID, bearer) })
	}
	err := c.enrichTaskContext(ctx, teamID, componentID, bearer, snapshot, req.DocumentSnapshot)
	extended.Wait()
	return err
}
