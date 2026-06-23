package riidoaiserver

func aiAgentClientHTTPRoute(_, path string) string {
	if route := aiAgentClientV1Route(path); route != "" {
		return route
	}
	return aiAgentClientV2Route(path)
}

func aiAgentClientRouteFromSegments(base string, segments []string) string {
	if len(segments) == 0 {
		return ""
	}
	switch segments[0] {
	case "bootstrap", "devices", "events":
		return aiAgentClientTopLevelRoute(base, segments)
	case "onboarding":
		return aiAgentClientOnboardingRoute(base, segments)
	case "profile-thumbnails":
		return aiAgentClientProfileThumbnailRoute(base, segments)
	case "tasks":
		return aiAgentClientTaskHTTPRoute(base, segments[1:])
	case "agent-assignments":
		return aiAgentClientWorkspaceAssignmentRoute(base, segments)
	case "threads":
		return aiAgentClientThreadHTTPRoute(base, segments)
	case "agents":
		return aiAgentClientAgentHTTPRoute(base, segments)
	default:
		return ""
	}
}

func aiAgentClientTopLevelRoute(base string, segments []string) string {
	if len(segments) == 1 {
		return base + "/" + segments[0]
	}
	if segments[0] == "devices" {
		return aiAgentClientDeviceHTTPRoute(base, segments)
	}
	return ""
}

func aiAgentClientOnboardingRoute(base string, segments []string) string {
	if len(segments) == 2 && segments[1] == "fixtures" {
		return base + "/onboarding/fixtures"
	}
	return ""
}

func aiAgentClientProfileThumbnailRoute(base string, segments []string) string {
	if len(segments) == 2 && segments[1] == "uploads" {
		return base + "/profile-thumbnails/uploads"
	}
	return ""
}
