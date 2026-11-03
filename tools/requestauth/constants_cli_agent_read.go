package main

var cliAgentReadRequiredRuleGroups = map[string][]string{
	"cli-agent-read": {
		"GET /v1/cli/workspaces/{workspace_id}/tasks/{task_id}/agents returns only current visible public AgentDirectoryEntry fields, 50 per page; optional canonical nonnegative integer cursor; current visible total only",
		"GET /v1/cli/workspaces/{workspace_id}/tasks/{task_id}/agents/{agent_id} returns current public detail; hidden and absent IDs are both forbidden with identical response",
		"CLI-only routes require exact single Bearer riido_cli_agent_v1_ credential; ordinary application, thread, device and alternate headers never authorize this route",
		"Agent opaque prefix is intercepted before generic or device authorization on every other resource/action; no fallback when disabled, rejected, cancelled or unavailable",
		"Fixed monolith /internal/cli/agent-read/authorize, schema riido-cli-agent-read.v1, purpose task_agents_read and current original grant/install/workspace/team/task verdict each page; thread contract remains distinct",
		"Five-second timeout, HTTPS fixed origins, no cookies/redirect/retry/authorization cache, bounded exact-ten-field UTF-8 response with no null/duplicate/unknown fields and current clock/expiry binding",
		"Independent default-off runtime flag and service credential; reject collisions with external/static/review credential representations and thread service token in both configurations",
		"No raw OAuth/cookie reaches CP and no private instruction/owner/device credential is returned; no-store and private errors",
		"No synthetic Assignment, run, quota use or run read record; policy_checked remains false and no self/precheck is fabricated; callable expresses only reported runtime availability reasons",
		"Dedicated CLI surface is not part of the generated application client SDK or its application-token contract; existing web and Agent-mode routes keep their authorization and projection behavior",
	},
}
