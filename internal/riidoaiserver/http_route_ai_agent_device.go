package riidoaiserver

func aiAgentClientDeviceHTTPRoute(base string, segments []string) string {
	if len(segments) == 1 {
		return base + "/devices"
	}
	if len(segments) < 3 || segments[2] != "daemon" {
		return ""
	}
	if len(segments) == 3 {
		return base + "/devices/{device_id}/daemon"
	}
	if len(segments) == 4 {
		return base + "/devices/{device_id}/daemon/{action}"
	}
	return ""
}

func aiAgentClientAgentHTTPRoute(base string, segments []string) string {
	if len(segments) == 1 {
		return base + "/agents"
	}
	if len(segments) == 2 {
		return base + "/agents/{agent_id}"
	}
	if len(segments) == 3 && segments[2] == "daemon" {
		return base + "/agents/{agent_id}/daemon"
	}
	if len(segments) == 4 && segments[2] == "daemon" {
		return base + "/agents/{agent_id}/daemon/{action}"
	}
	if len(segments) == 3 && segments[2] == "editability" {
		return base + "/agents/{agent_id}/editability"
	}
	return ""
}
