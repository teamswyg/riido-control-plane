package requirements

var ValidationRules = []string{
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

var RoutingRules = []string{
	"runtime-provider-required",
	"available-allowed",
	"login-required-blocked",
	"unsupported-blocked",
	"store-blocked-blocked",
	"missing-synced-provider-blocked",
	"not-synced-allowed",
	"unknown-routing-status-rejected",
}
