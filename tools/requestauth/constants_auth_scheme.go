package main

var authorizationSchemeRequiredSurfaces = []string{
	"AuthorizationScheme",
	"AuthorizationSchemeSelector",
	"AuthorizationSchemeRouter",
	"IssuerAuthorizationSchemeSelector",
}

var authorizationSchemeRequiredRuleGroups = map[string][]string{
	"authorization-scheme-routing": {
		"explicit-legacy-v1-and-auth-service-v2",
		"selector-classifies-but-does-not-authenticate",
		"legacy-v1-behavior-preserved",
		"v2-failure-never-downgrades",
		"v2-unavailable-fails-closed",
		"unknown-scheme-fails-closed",
	},
}
