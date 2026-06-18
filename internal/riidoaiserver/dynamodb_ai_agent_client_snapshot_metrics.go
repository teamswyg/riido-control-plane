package riidoaiserver

import "time"

func (s *DynamoDBAIAgentClientSnapshot) observeAIAgentClientSnapshot(operation AIAgentClientPersistenceOperation, startedAt time.Time, bytes int64, err error) {
	if s == nil || s.metrics == nil {
		return
	}
	s.metrics.ObserveAIAgentClientPersistence(AIAgentClientPersistenceObservation{
		Operation: operation,
		Duration:  time.Since(startedAt),
		Bytes:     bytes,
		Err:       err,
	})
}
