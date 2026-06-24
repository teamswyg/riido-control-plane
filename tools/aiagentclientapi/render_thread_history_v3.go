package main

import (
	"fmt"
	"strings"
)

func renderThreadHistoryV3(b *strings.Builder, v3 threadHistoryV3) {
	b.WriteString("\n## Frontend Thread History v3 Handoff\n\n")
	b.WriteString("Use v3 for initial/read refresh, v2 for mutations, and v2 SSE for live progress.\n\n")
	renderEndpointTable(b, append([]endpointRule{v3.ReadEndpoint, v3.SSEEndpoint}, v3.ActionEndpoints...))
	renderNamedRuleTable(b, "Thread History v3 Frontend Flow", v3.ImplementationSteps)
	renderShapeRules(b, v3.ResponseShapes)
	renderIdentityTable(b, v3.IdentityRules)
	renderBullets(b, "Thread History v3 Grouping Rules", v3.GroupingRules)
	renderBullets(b, "Thread History v3 Agent Snapshot Rules", v3.AgentSnapshotRules)
	renderMessageRoleTable(b, v3.MessageRoles)
	renderNamedRuleTable(b, "Thread History v3 Message Ordering", v3.OrderingRules)
	renderNamedRuleTable(b, "Thread History v3 Mutation Rules", v3.MutationRules)
	renderBullets(b, "Thread History v3 SSE Handling Rules", v3.SSEHandlingRules)
	renderList(b, "Thread History v3 Terminal States", v3.TerminalStates)
	renderBullets(b, "Thread History v3 Frontend Checklist", v3.Checklist)
}

func renderEndpointTable(b *strings.Builder, endpoints []endpointRule) {
	b.WriteString("### Endpoint Split\n\n| Name | Method | Path | Purpose | Truth role |\n| --- | --- | --- | --- | --- |\n")
	for _, endpoint := range endpoints {
		fmt.Fprintf(b, "| %s | `%s` | `%s` | %s | %s |\n", endpoint.Name, endpoint.Method, endpoint.Path, endpoint.Purpose, endpoint.TruthRole)
	}
}

func renderIdentityTable(b *strings.Builder, rules []identityRule) {
	b.WriteString("\n### Identity Rules\n\n| Name | Meaning | Frontend use |\n| --- | --- | --- |\n")
	for _, rule := range rules {
		fmt.Fprintf(b, "| `%s` | %s | %s |\n", rule.Name, rule.Meaning, rule.Use)
	}
}

func renderMessageRoleTable(b *strings.Builder, roles []messageRole) {
	b.WriteString("\n### Message Roles\n\n| Role | Meaning |\n| --- | --- |\n")
	for _, role := range roles {
		fmt.Fprintf(b, "| `%s` | %s |\n", role.Role, role.Meaning)
	}
}

func renderNamedRuleTable(b *strings.Builder, title string, rules []namedRule) {
	b.WriteString("\n### " + title + "\n\n| Rule | Detail |\n| --- | --- |\n")
	for _, rule := range rules {
		fmt.Fprintf(b, "| `%s` | %s |\n", rule.Name, rule.Detail)
	}
}

func renderShapeRules(b *strings.Builder, shapes []shapeRule) {
	b.WriteString("\n### Thread History v3 Response Model\n")
	for _, shape := range shapes {
		fmt.Fprintf(b, "\n#### `%s`\n\n%s\n\n", shape.Name, shape.Purpose)
		fmt.Fprintf(b, "- Fields: %s\n", codeList(shape.Fields))
	}
}

func renderBullets(b *strings.Builder, title string, items []string) {
	b.WriteString("\n## " + title + "\n\n")
	for _, item := range items {
		fmt.Fprintf(b, "- %s\n", item)
	}
}
