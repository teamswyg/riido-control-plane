package riidoaiserver

import (
	"errors"
	"sort"
	"strings"
)

func handleListAgentCatalog(state *storeState) []AgentCatalogRecord {
	ids := make([]string, 0, len(state.agentCatalog))
	for id := range state.agentCatalog {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	records := make([]AgentCatalogRecord, 0, len(ids))
	for _, id := range ids {
		records = append(records, copyAgentCatalogRecord(state.agentCatalog[id]))
	}
	return records
}

func handleGetAgentCatalog(state *storeState, agentID string) (AgentCatalogRecord, bool, error) {
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return AgentCatalogRecord{}, false, errors.New("agent_id is required")
	}
	record, ok := state.agentCatalog[agentID]
	return copyAgentCatalogRecord(record), ok, nil
}

func handleSaveAgentCatalog(state *storeState, record AgentCatalogRecord) (AgentCatalogRecord, error) {
	record = normalizeAgentCatalogRecord(record)
	if err := record.Validate(); err != nil {
		return AgentCatalogRecord{}, err
	}
	state.agentCatalog[record.AgentID] = record
	return copyAgentCatalogRecord(record), nil
}

func handleDeleteAgentCatalog(state *storeState, agentID string) (bool, error) {
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return false, errors.New("agent_id is required")
	}
	if _, ok := state.agentCatalog[agentID]; !ok {
		return false, nil
	}
	delete(state.agentCatalog, agentID)
	return true, nil
}

func copyAgentCatalogRecord(record AgentCatalogRecord) AgentCatalogRecord {
	return AgentCatalogRecord{
		AgentID:          record.AgentID,
		OwnerPrincipalID: record.OwnerPrincipalID,
		Visibility:       record.Visibility,
	}
}
