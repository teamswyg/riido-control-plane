package riidoaiserver

import "strings"

func threadMessageSuffixThreadID(suffix string) (string, bool) {
	parts := strings.Split(strings.Trim(suffix, "/"), "/")
	if len(parts) != 3 || parts[0] != "threads" || strings.TrimSpace(parts[1]) == "" || parts[2] != "messages" {
		return "", false
	}
	return parts[1], true
}

func agentAssignmentSuffixAgentID(suffix string) (string, bool) {
	parts := strings.Split(strings.Trim(suffix, "/"), "/")
	if len(parts) != 2 || parts[0] != "agent-assignments" || strings.TrimSpace(parts[1]) == "" {
		return "", false
	}
	return parts[1], true
}

func agentAssignmentStopSuffixAgentID(suffix string) (string, bool) {
	parts := strings.Split(strings.Trim(suffix, "/"), "/")
	if len(parts) != 3 || parts[0] != "agent-assignments" || strings.TrimSpace(parts[1]) == "" || parts[2] != "stop" {
		return "", false
	}
	return parts[1], true
}

func splitAIAgentClientDevicePath(path string) (string, string, bool) {
	const prefix = "/v1/client/ai-agent/devices/"
	if !strings.HasPrefix(path, prefix) {
		return "", "", false
	}
	rest := strings.Trim(strings.TrimPrefix(path, prefix), "/")
	parts := strings.Split(rest, "/")
	if len(parts) < 2 || len(parts) > 3 || strings.TrimSpace(parts[0]) == "" {
		return "", "", false
	}
	for _, part := range parts[1:] {
		if strings.TrimSpace(part) == "" {
			return "", "", false
		}
	}
	return parts[0], strings.Join(parts[1:], "/"), true
}

func splitAIAgentClientAgentPath(path string) (string, string, bool) {
	const prefix = "/v1/client/ai-agent/agents/"
	if !strings.HasPrefix(path, prefix) {
		return "", "", false
	}
	rest := strings.Trim(strings.TrimPrefix(path, prefix), "/")
	if rest == "" {
		return "", "", false
	}
	parts := strings.Split(rest, "/")
	if len(parts) < 1 || len(parts) > 3 {
		return "", "", false
	}
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			return "", "", false
		}
	}
	if len(parts) == 1 {
		return parts[0], "", true
	}
	return parts[0], strings.Join(parts[1:], "/"), true
}
