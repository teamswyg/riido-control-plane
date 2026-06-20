package main

const (
	defaultManifest = "docs/20-domain/agent-runtime-binding.riido.json"
	manifestSchema  = "riido-agent-runtime-binding.v1"
	evidenceSchema  = "riido-agent-runtime-binding-evidence.v1"
	expectedID      = "agent-runtime-binding"
	expectedTask    = "RIID-4665"
)

var requiredFields = []string{"agent_id", "daemon_id", "runtime_id", "runtime_provider"}

var requiredRules = []string{
	"unique-agent",
	"assignment-provider-match",
	"daemon-id-match",
	"runtime-id-match",
	"device-id-if-bound",
	"nil-registry-no-gate",
}

var requiredDeviceRules = []string{"machine-device-id", "connected-workspaces", "cross-workspace-bindings", "legacy-runtime-prune"}
