package riidoaiserver

import (
	"context"
	"errors"
	"sort"
	"strings"
)

func (s *Store) ListAgentCatalog(ctx context.Context) ([]AgentCatalogRecord, error) {
	reply := make(chan listAgentCatalogResult, 1)
	if err := s.send(ctx, listAgentCatalogCmd{reply: reply}); err != nil {
		return nil, err
	}
	select {
	case res := <-reply:
		return res.records, res.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *Store) GetAgentCatalog(ctx context.Context, agentID string) (AgentCatalogRecord, bool, error) {
	reply := make(chan getAgentCatalogResult, 1)
	if err := s.send(ctx, getAgentCatalogCmd{agentID: agentID, reply: reply}); err != nil {
		return AgentCatalogRecord{}, false, err
	}
	select {
	case res := <-reply:
		return res.record, res.ok, res.err
	case <-ctx.Done():
		return AgentCatalogRecord{}, false, ctx.Err()
	}
}

func (s *Store) SaveAgentCatalog(ctx context.Context, record AgentCatalogRecord) (AgentCatalogRecord, error) {
	reply := make(chan saveAgentCatalogResult, 1)
	if err := s.send(ctx, saveAgentCatalogCmd{record: record, reply: reply}); err != nil {
		return AgentCatalogRecord{}, err
	}
	select {
	case res := <-reply:
		return res.record, res.err
	case <-ctx.Done():
		return AgentCatalogRecord{}, ctx.Err()
	}
}

func (s *Store) DeleteAgentCatalog(ctx context.Context, agentID string) (bool, error) {
	reply := make(chan deleteAgentCatalogResult, 1)
	if err := s.send(ctx, deleteAgentCatalogCmd{agentID: agentID, reply: reply}); err != nil {
		return false, err
	}
	select {
	case res := <-reply:
		return res.deleted, res.err
	case <-ctx.Done():
		return false, ctx.Err()
	}
}

type listAgentCatalogCmd struct {
	reply chan listAgentCatalogResult
}

type listAgentCatalogResult struct {
	records []AgentCatalogRecord
	err     error
}

type getAgentCatalogCmd struct {
	agentID string
	reply   chan getAgentCatalogResult
}

type getAgentCatalogResult struct {
	record AgentCatalogRecord
	ok     bool
	err    error
}

type saveAgentCatalogCmd struct {
	record AgentCatalogRecord
	reply  chan saveAgentCatalogResult
}

type saveAgentCatalogResult struct {
	record AgentCatalogRecord
	err    error
}

type deleteAgentCatalogCmd struct {
	agentID string
	reply   chan deleteAgentCatalogResult
}

type deleteAgentCatalogResult struct {
	deleted bool
	err     error
}

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
