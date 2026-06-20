package riidoaiserver

import (
	"context"
)

type AIAgentClientEventSubscriber interface {
	SubscribeAIAgentClientEvents(ctx context.Context, principal AuthorizationResult) ([]ClientStreamEvent, <-chan ClientStreamEvent, func(), error)
}
