package requirements

const (
	DefaultManifest = "docs/20-domain/agent-catalog-rbac.riido.json"
	ManifestSchema  = "riido-agent-catalog-rbac.v1"
	EvidenceSchema  = "riido-agent-catalog-rbac-evidence.v1"
	ExpectedID      = "agent-catalog-rbac"
	ExpectedTask    = "RIID-4663"
)

var RequiredRoutes = []string{
	"GET /v1/agent-catalog",
	"POST /v1/agent-catalog",
	"GET /v1/agent-catalog/{agent_id}",
	"PATCH /v1/agent-catalog/{agent_id}",
	"DELETE /v1/agent-catalog/{agent_id}",
}
