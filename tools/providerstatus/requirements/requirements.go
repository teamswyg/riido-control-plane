package requirements

const (
	DefaultManifest = "docs/20-domain/provider-status.riido.json"
	ManifestSchema  = "riido-provider-status.v1"
	EvidenceSchema  = "riido-provider-status-evidence.v1"
	ExpectedID      = "provider-status"
	ExpectedTask    = "RIID-4671"
)

var Surfaces = []string{
	"ProviderStatusRecord",
	"ProviderStatusSyncRequest",
	"ProviderStatusSyncResponse",
	"ProviderStatusStore",
	"ProviderStatusReader",
	"StoreSafeRoutingInput",
	"StoreSafeRoutingDecision",
}

var RoutingStatuses = []string{"available", "login-required", "unsupported", "store-blocked"}

var DistributionChannels = []string{"developer-id", "mac-app-store", "msix-sideload", "msix-store", "dev-local"}
