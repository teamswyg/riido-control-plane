package riidoaiserver

import (
	"context"
)

func (s *DevelopmentAIAgentClientStore) SubscribeAIAgentClientEvents(
	ctx context.Context,
	principal AuthorizationResult,
) ([]ClientStreamEvent, <-chan ClientStreamEvent, func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	history := s.clientEventsForPrincipalLocked(principal)
	s.nextSubscriberID++
	id := s.nextSubscriberID
	events := make(chan ClientStreamEvent, 32)
	s.subscribers[id] = aiAgentClientSubscriber{
		principal:     principal,
		visibilityKey: subscriberVisibilityKey(principal),
		events:        events,
	}
	cancel := func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if subscriber, ok := s.subscribers[id]; ok {
			delete(s.subscribers, id)
			close(subscriber.events)
		}
	}
	return history, events, cancel, nil
}
