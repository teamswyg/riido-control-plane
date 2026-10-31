package main

var cliThreadReadRequiredSurfaces = []string{"CLIThreadReadAuthorizer"}

var cliThreadReadRequiredRuleGroups = map[string][]string{
	"cli-thread-read": {
		"default-off-dedicated-reader",
		"opaque-ticket-before-generic-or-device",
		"exact-v3-task-threads-get-only",
		"canonical-single-bearer-no-query-origin-or-upgrade",
		"monolith-current-grant-installation-task-check-every-read",
		"no-authorization-cache-retry-or-fallback",
		"fixed-https-origin-no-redirect-or-cookie",
		"service-credential-separate-from-app-and-installation",
		"exact-ten-field-response-and-five-second-clock-bound",
		"resource-purpose-workspace-task-and-one-minute-expiry",
		"current-owner-public-admin-thread-visibility-preserved",
		"no-store-private-errors",
	},
}
