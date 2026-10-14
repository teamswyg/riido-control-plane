package riidoaiserver

import "strings"

func unmatchedHTTPRoute(path string) string {
	segments := unmatchedHTTPRouteSegments(path)
	if len(segments) == 0 {
		return unmatchedHTTPRoutePrefix + "/"
	}
	first := strings.ToLower(segments[0])
	switch first {
	case "favicon.ico", "robots.txt", "manifest.json", "assetlinks.json", "apple-app-site-association":
		return unmatchedHTTPRoutePrefix + "/" + first
	case ".well-known":
		return unmatchedHTTPRoutePrefix + "/.well-known"
	case "v1", "v2":
		return unmatchedHTTPRoutePrefix + "/" + first + "/" + unmatchedHTTPAPISegment(segments)
	default:
		if strings.Contains(first, ".") {
			return unmatchedHTTPRoutePrefix + "/{asset}"
		}
		return unmatchedHTTPRoutePrefix + "/{other}"
	}
}

func unmatchedHTTPRouteSegments(path string) []string {
	trimmed := strings.Trim(strings.TrimSpace(path), "/")
	if trimmed == "" {
		return nil
	}
	parts := strings.Split(trimmed, "/")
	segments := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			segments = append(segments, part)
		}
	}
	return segments
}

func unmatchedHTTPAPISegment(segments []string) string {
	if len(segments) < 2 {
		return "{other}"
	}
	second := strings.ToLower(segments[1])
	switch second {
	case "agent-catalog", "agents", "client", "component-tasks", "daemon", "desktop":
		return second
	default:
		return "{other}"
	}
}

func excludedHTTPTransactionMetricsPath(path string) bool {
	switch path {
	case "/healthz", "/readyz", "/metrics":
		return true
	default:
		return false
	}
}
