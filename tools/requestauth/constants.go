package main

const (
	defaultManifest = "docs/20-domain/request-authorization.riido.json"
	manifestSchema  = "riido-request-authorization.v1"
	evidenceSchema  = "riido-request-authorization-evidence.v1"
	expectedID      = "request-authorization"
	expectedTask    = "RIID-4664"
)

var requiredSurfaces = []string{
	"RequestAuthorizer",
	"AuthorizationRequest",
	"AuthorizationResult",
	"StaticTokenAuthorizer",
	"FallbackAuthorizer",
	"ExternalHTTPAuthorizer",
	"ExternalHTTPAuthorizerConfig",
}

var requiredResources = []string{
	"agent",
	"ai_agent_client",
	"agent_catalog",
	"component_task",
	"component_task_events",
	"metrics",
}

var requiredTransports = []string{
	"generated-client-token",
	"compat-bearer-token",
	"device-id",
	"device-secret",
	"external-api-key",
}

var requiredContractVersions = []string{
	"riido-external-authorizer-request.v1",
	"riido-external-authorizer-response.v1",
}

var requiredRuntimeConfigKeys = []string{"RIIDO_AI_SERVER_WEB_ALLOWED_ORIGINS"}

var requiredRuleGroups = map[string][]string{
	"static-token": {
		"principal-required", "exactly-one-token-material", "sha256-hash-supported",
		"duplicate-token-rejected", "scopes-required", "constant-time-compare", "scoped-deny-forbidden",
	},
	"external-authorizer": {
		"request-schema-v1", "response-schema-v1", "opaque-bearer-token",
		"ai-agent-client-workspace-required", "api-key-header-server-only",
		"response-disallow-unknown-fields", "response-size-limit", "allowed-principal-required", "admin-role-only",
	},
	"fail-closed": {
		"http-401-unauthenticated", "http-403-forbidden", "allowed-false-forbidden",
		"non-2xx-service-error", "malformed-json-service-error",
		"unsupported-schema-service-error", "invalid-role-service-error", "network-error-service-error",
	},
	"fallback": {"next-only-unauthenticated", "forbidden-stops-chain", "empty-chain-unauthenticated"},
	"cors":     {"exact-origin-allowlist", "method-allowlist", "header-allowlist", "no-browser-credentials", "unsupported-header-rejected"},
	"device-principal": {
		"device-headers-server-only", "both-device-fields-required",
		"device-auth-before-token-auth", "browser-jwt-not-daemon-auth",
	},
}
