package riidoaiserver

type CreateAgentCatalogRequest struct {
	AgentID    string                 `json:"agent_id"`
	Visibility AgentCatalogVisibility `json:"visibility"`
}

type UpdateAgentCatalogRequest struct {
	Visibility AgentCatalogVisibility `json:"visibility"`
}

type AgentCatalogListResponse struct {
	SchemaVersion string               `json:"schema_version"`
	Agents        []AgentCatalogRecord `json:"agents"`
}

type AgentCatalogRecordResponse struct {
	SchemaVersion string             `json:"schema_version"`
	Agent         AgentCatalogRecord `json:"agent"`
}
