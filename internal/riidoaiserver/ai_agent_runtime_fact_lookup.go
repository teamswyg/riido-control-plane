package riidoaiserver

import (
	"strings"
	"time"
)

func (s *DevelopmentAIAgentClientStore) LookupAgentRuntimeFact(
	agentID string,
) (AgentRuntimeBinding, RuntimeRecord, bool) {
	if s == nil || strings.TrimSpace(agentID) == "" {
		return AgentRuntimeBinding{}, RuntimeRecord{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	agent, ok := s.agents[strings.TrimSpace(agentID)]
	if !ok {
		return AgentRuntimeBinding{}, RuntimeRecord{}, false
	}
	return s.agentRuntimeFactLocked(agent, time.Now().UTC())
}
