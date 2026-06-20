package main

const (
	defaultManifest = "docs/20-domain/agent-catalog-rbac.riido.json"
	manifestSchema  = "riido-agent-catalog-rbac.v1"
	evidenceSchema  = "riido-agent-catalog-rbac-evidence.v1"
	expectedID      = "agent-catalog-rbac"
	expectedTask    = "RIID-4663"
)

var requiredRoutes = []string{
	"GET /v1/agent-catalog",
	"POST /v1/agent-catalog",
	"GET /v1/agent-catalog/{agent_id}",
	"PATCH /v1/agent-catalog/{agent_id}",
	"DELETE /v1/agent-catalog/{agent_id}",
}
