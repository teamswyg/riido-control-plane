package riidoaiserver

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
