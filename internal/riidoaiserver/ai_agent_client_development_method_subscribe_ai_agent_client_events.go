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
	s.compensateRecentTerminalProgressLocked()
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
		subscriber, ok := s.subscribers[id]
		if ok {
			delete(s.subscribers, id)
			close(subscriber.events)
		}
		s.mu.Unlock()
		if ok {
			logAIAgentClientSubscriberDeliverySummary(subscriber)
		}
	}
	return history, events, cancel, nil
}
