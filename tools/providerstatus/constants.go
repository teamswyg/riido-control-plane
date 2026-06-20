package main

const (
	defaultManifest = "docs/20-domain/provider-status.riido.json"
	manifestSchema  = "riido-provider-status.v1"
	evidenceSchema  = "riido-provider-status-evidence.v1"
	expectedID      = "provider-status"
	expectedTask    = "RIID-4671"
)

var requiredSurfaces = []string{
	"ProviderStatusRecord",
	"ProviderStatusSyncRequest",
	"ProviderStatusSyncResponse",
	"ProviderStatusStore",
	"ProviderStatusReader",
	"StoreSafeRoutingInput",
	"StoreSafeRoutingDecision",
}

var requiredRoutingStatuses = []string{"available", "login-required", "unsupported", "store-blocked"}

var requiredDistributionChannels = []string{"developer-id", "mac-app-store", "msix-sideload", "msix-store", "dev-local"}

var requiredValidationRules = []string{
	"agent-id-required",
	"identity-fields-trimmed",
	"daemon-id-required",
	"runtime-id-required",
	"distribution-channel-known",
	"providers-required",
	"provider-kind-required",
	"provider-kind-unique",
	"routing-status-known",
	"providers-sorted",
}

var requiredRoutingRules = []string{
	"runtime-provider-required",
	"available-allowed",
	"login-required-blocked",
	"unsupported-blocked",
	"store-blocked-blocked",
	"missing-synced-provider-blocked",
	"not-synced-allowed",
	"unknown-routing-status-rejected",
}
