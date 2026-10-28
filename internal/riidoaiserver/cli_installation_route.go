package riidoaiserver

import (
	"net/http"
	"strings"
)

func cliInstallationHTTPRoute(method, path string) string {
	if method != http.MethodPost {
		return ""
	}
	switch path {
	case "/v1/internal/cli-installations/link-intents", "/v1/internal/cli-installations/policy":
		return path
	}
	if strings.HasPrefix(path, "/v1/daemon/cli-installation-links/") && strings.HasSuffix(path, "/confirm") {
		return "/v1/daemon/cli-installation-links/{intent_id}/confirm"
	}
	return ""
}
