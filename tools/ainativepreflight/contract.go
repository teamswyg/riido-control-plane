package main

import "regexp"

var (
	keyNames = []string{"RIIDO_AI_SERVER_EXTERNAL_AUTHZ_API_KEY", "RIIDO_AI_SERVER_COMMENT_DELIVERY_KEY", "RIIDO_AI_SERVER_AGENT_SOURCE_STATE_KEY"}
	urlPaths = map[string]string{
		"RIIDO_AI_SERVER_EXTERNAL_AUTHZ_URL":            "/internal/control-plane/authorize",
		"RIIDO_AI_SERVER_TASK_CONTEXT_GRAPHQL_URL":      "/graphql",
		"RIIDO_AI_SERVER_TASK_CONTEXT_BASE_URL":         "",
		"RIIDO_AI_SERVER_COMMENT_DELIVERY_EVIDENCE_URL": "",
		"RIIDO_AI_SERVER_AGENT_SOURCE_STATE_URL":        "",
	}
)
var nonProduction = regexp.MustCompile(`(^|[.-])(staging|stage|testnet|test|development|dev|localhost)([.-]|$)`)
