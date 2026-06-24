package riidoaiserver

import "errors"

func (s Server) resolveAgentRuntimeFact(agentID string) (AgentRuntimeBinding, RuntimeRecord, error) {
	if factRegistry, ok := s.aiAgent.(AgentRuntimeFactRegistry); ok {
		binding, runtime, found := factRegistry.LookupAgentRuntimeFact(agentID)
		if !found {
			return AgentRuntimeBinding{}, RuntimeRecord{}, errors.New("ai agent runtime binding is not configured")
		}
		return binding, runtime, nil
	}
	registry, ok := s.aiAgent.(AgentRegistry)
	if !ok {
		return AgentRuntimeBinding{}, RuntimeRecord{}, errors.New("ai agent runtime registry is not configured")
	}
	binding, found := registry.LookupAgent(agentID)
	if !found {
		return AgentRuntimeBinding{}, RuntimeRecord{}, errors.New("ai agent runtime binding is not configured")
	}
	return binding, RuntimeRecord{}, nil
}
