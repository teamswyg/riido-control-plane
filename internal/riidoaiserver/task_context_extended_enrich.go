package riidoaiserver

import (
	"context"
	"sync"
)

// enrichTaskContextFor runs the extended reads beside the regular enrichment
// under the original request deadline, including optional reply metadata.
func (c *AIAgentPrivateTaskContextClient) enrichTaskContextFor(ctx context.Context, teamID, componentID, bearer string, snapshot *AIAgentTaskContext, req AIAgentTaskContextRequest) error {
	var extended sync.WaitGroup
	if req.Extended {
		extended.Go(func() {
			snapshot.Extended = c.readExtendedTaskContext(ctx, teamID, componentID, bearer)
			if req.ReplyMetadata && snapshot.Extended.CommentStatus != "unavailable" {
				replies := c.readCurrentReplyMetadata(ctx, teamID, bearer, snapshot.Extended.Comments)
				snapshot.Extended.Comments = append(snapshot.Extended.Comments, replies...)
			}
		})
	}
	err := c.enrichTaskContext(ctx, teamID, componentID, bearer, snapshot, req.DocumentSnapshot)
	extended.Wait()
	return err
}
