package main

const (
	defaultManifest = "docs/20-domain/context-map.riido.json"
	manifestSchema  = "riido-control-plane-context-map.v1"
	evidenceSchema  = "riido-control-plane-context-map-evidence.v1"
	expectedID      = "control-plane-context-map"
	expectedTask    = "RIID-4712"
)

var requiredOwnedContexts = []string{
	"c10-saas-control-plane",
	"c10-runtime-adapter",
	"c10-public-aws-adapter-facade",
	"c10-container-contract",
}

var requiredImportedContexts = []string{
	"assignment-polling-contract",
	"provider-distribution-vocabulary",
	"go-dependency-allowlist",
}

var requiredExternalContexts = []string{
	"customer-pc-daemon-runtime",
	"shared-contract-tags",
	"infrastructure-deployment",
	"store-app-clients",
}
