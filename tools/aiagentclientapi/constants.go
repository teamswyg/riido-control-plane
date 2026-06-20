package main

const (
	defaultManifest = "docs/20-domain/ai-agent-client-api.riido.json"
	manifestSchema  = "riido-ai-agent-client-api.v1"
	evidenceSchema  = "riido-ai-agent-client-api-evidence.v1"
	expectedID      = "ai-agent-client-api"
	expectedTask    = "RIID-4721"
)

var requiredGeneratedPaths = []string{
	"aiAgent.bootstrap",
	"aiAgent.profileThumbnails.uploads.create",
	"aiAgent.tasks.assign",
	"aiAgent.tasks.threadMessages.create",
	"v2.aiAgent.bootstrap",
	"v2.aiAgent.agents.create",
	"v2.aiAgent.profileThumbnails.uploads.create",
	"v2.aiAgent.tasks.assignedAgentProfiles",
	"v2.aiAgent.tasks.agentAssignments.create",
	"v2.aiAgent.tasks.threadStreamSubscription",
}

var requiredRuntimeConfigKeys = []string{"RIIDO_AI_SERVER_AI_AGENT_CLIENT_DEVELOPMENT"}

var requiredPublicFields = []string{
	"workspace_id",
	"profile_thumbnail_url",
	"provider_version",
	"assigned-agent-profiles",
	"agent-assignments",
	"thread-stream-subscription",
}

var requiredDeploymentEvidence = []string{
	"not from a manual",
	"The workflow masks both values",
}
