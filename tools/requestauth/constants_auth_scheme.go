package main

var authorizationSchemeRequiredSurfaces = []string{
	"AuthorizationScheme",
	"AuthorizationSchemeSelector",
	"AuthorizationSchemeRouter",
	"IssuerAuthorizationSchemeSelector",
	"RestrictLegacyApplicationTokens",
}

var authorizationSchemeRequiredRuleGroups = map[string][]string{
	"monolith-cli-credential-boundary": {
		"classification-can-only-deny", "cli-grant-or-client-marker-never-legacy-app-token",
		"duplicate-escaped-case-variant-markers-denied", "legacy-jwt-envelope-bounded-to-sixteen-kib",
		"no-external-static-or-scope-fallback", "normal-app-and-opaque-credentials-preserved",
		"cli-resource-read-exchange-remains-separate",
	},
	"authorization-scheme-routing": {
		"explicit-legacy-v1-and-auth-service-v2",
		"selector-classifies-but-does-not-authenticate",
		"legacy-v1-behavior-preserved",
		"v2-failure-never-downgrades",
		"v2-unavailable-fails-closed",
		"unknown-scheme-fails-closed",
	},
}
