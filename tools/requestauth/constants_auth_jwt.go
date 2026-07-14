package main

var requiredRuntimeConfigKeys = []string{
	"RIIDO_AI_SERVER_WEB_ALLOWED_ORIGINS",
	"RIIDO_AI_SERVER_AUTH_ISSUER",
	"RIIDO_AI_SERVER_AUTH_RESOURCE",
	"RIIDO_AI_SERVER_AUTH_AUTHORIZATION_PROFILE",
	"RIIDO_AI_SERVER_AUTH_INTROSPECTION_CLIENT_ID",
	"RIIDO_AI_SERVER_AUTH_INTROSPECTION_CLIENT_SECRET",
	"RIIDO_AI_SERVER_AUTH_HTTP_TIMEOUT_SECONDS",
}

var jwtRequiredRuleGroups = map[string][]string{
	"auth-jwt-pep": {
		"exact-https-issuer-and-resource", "typ-at-jwt-only", "es256-p256-only", "known-kid-only",
		"jwks-etag-max-age-60", "expired-jwks-fails-closed", "unknown-kid-one-refresh",
		"exact-authorization-profile", "canonical-non-wildcard-scopes", "auth-introspection-required",
		"claim-equivalence-required", "jwt-shaped-invalid-stops-fallback", "consumer-domain-pdp-required",
		"domain-subject-and-workspace-equivalence",
	},
}

func mergeRequiredRuleGroups(groups ...map[string][]string) map[string][]string {
	result := map[string][]string{}
	for _, group := range groups {
		for id, rules := range group {
			result[id] = rules
		}
	}
	return result
}
