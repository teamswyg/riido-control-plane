package riidoaiserver

import "context"

type AgentCatalogStore interface {
	ListAgentCatalog(ctx context.Context) ([]AgentCatalogRecord, error)
	GetAgentCatalog(ctx context.Context, agentID string) (AgentCatalogRecord, bool, error)
	SaveAgentCatalog(ctx context.Context, record AgentCatalogRecord) (AgentCatalogRecord, error)
	DeleteAgentCatalog(ctx context.Context, agentID string) (bool, error)
}
