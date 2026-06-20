package main

var requiredClientMetadataFields = map[string][]string{
	"getAIAgentClientBootstrap":       {"cache_tag"},
	"listAIAgentTaskAssignableAgents": {"cache_tag"},
	"assignAIAgentTask":               {"invalidates"},
	"unassignAIAgentTask":             {"invalidates"},
	"deleteAIAgent":                   {"invalidates"},
	"createAIAgentV2":                 {"invalidates"},
	"getAIAgentClientBootstrapV2":     {"cache_tag"},
	"submitAIAgentTaskComment":        {"invalidates"},
	"createAIAgentTaskThreadMessage":  {"invalidates"},
	"stopAIAgentTask":                 {"invalidates"},
}
