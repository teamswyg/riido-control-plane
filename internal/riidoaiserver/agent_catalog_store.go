package riidoaiserver

import (
	"context"
	"errors"
)

func (s *Store) ListAgentCatalog(ctx context.Context) ([]AgentCatalogRecord, error) {
	reply := make(chan listAgentCatalogResult, 1)
	if err := s.send(ctx, listAgentCatalogCmd{reply: reply}); err != nil {
		return nil, err
	}
	select {
	case res := <-reply:
		return res.records, res.err
	case <-s.done:
		return nil, errors.New("riido-control-plane store closed")
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
	case <-s.done:
		return AgentCatalogRecord{}, false, errors.New("riido-control-plane store closed")
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
	case <-s.done:
		return AgentCatalogRecord{}, errors.New("riido-control-plane store closed")
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
	case <-s.done:
		return false, errors.New("riido-control-plane store closed")
	case <-ctx.Done():
		return false, ctx.Err()
	}
}
