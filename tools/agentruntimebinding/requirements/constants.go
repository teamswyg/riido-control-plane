package requirements

const (
	DefaultManifest = "docs/20-domain/agent-runtime-binding.riido.json"
	ManifestSchema  = "riido-agent-runtime-binding.v1"
	EvidenceSchema  = "riido-agent-runtime-binding-evidence.v1"
	ExpectedID      = "agent-runtime-binding"
	ExpectedTask    = "RIID-4665"
)

var RequiredFields = []string{"agent_id", "daemon_id", "runtime_id", "runtime_provider"}

var RequiredRules = []string{
	"unique-agent",
	"assignment-provider-match",
	"daemon-id-match",
	"runtime-id-match",
	"device-id-if-bound",
	"nil-registry-no-gate",
}

var RequiredDeviceRules = []string{
	"machine-device-id",
	"connected-workspaces",
	"cross-workspace-bindings",
	"legacy-runtime-prune",
}
