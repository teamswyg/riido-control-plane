package riidoaiserver

type subscriberEventCache struct {
	key   string
	ready bool
	event ClientStreamEvent
	ok    bool
}

func (cache *subscriberEventCache) eventFor(
	s *DevelopmentAIAgentClientStore,
	subscriber aiAgentClientSubscriber,
	event ClientStreamEvent,
	fanoutEvent ClientStreamEvent,
	progressFanout bool,
) (ClientStreamEvent, bool) {
	if cache.ready && cache.key == subscriber.visibilityKey {
		return cache.event, cache.ok
	}
	cache.key = subscriber.visibilityKey
	cache.event, cache.ok = clientEventForSubscriberLocked(
		s, subscriber.principal, event, fanoutEvent, progressFanout,
	)
	cache.ready = true
	return cache.event, cache.ok
}
