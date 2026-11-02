package riidoaiserver

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
)

func cliAgentReadTargetFromRequest(r *http.Request) (cliAgentReadTarget, bool) {
	target, ok := cliAgentReadPath(r.URL.Path)
	if !ok || r.Method != http.MethodGet || r.RequestURI != r.URL.RequestURI() || r.URL.RawPath != "" || r.URL.ForceQuery || r.Header.Get("Origin") != "" || r.Header.Get("Upgrade") != "" {
		return cliAgentReadTarget{}, false
	}
	if r.URL.RawQuery != "" {
		if target.agent != "" || !strings.HasPrefix(r.URL.RawQuery, "cursor=") {
			return cliAgentReadTarget{}, false
		}
		raw := strings.TrimPrefix(r.URL.RawQuery, "cursor=")
		page, err := strconv.Atoi(raw)
		if err != nil || page < 0 || strconv.Itoa(page) != raw {
			return cliAgentReadTarget{}, false
		}
		target.page = page
	}
	return target, true
}

func cliAgentReadRequestToken(r *http.Request, req AuthorizationRequest) (string, bool) {
	target, ok := cliAgentReadTargetFromRequest(r)
	if !ok || req.Resource != AuthorizationResourceCLIAgentRead || req.Action != AuthorizationActionRead || req.WorkspaceID != target.workspace || req.TaskID != target.task || req.AgentID != target.agent || req.DeviceID != "" {
		return "", false
	}
	values := r.Header.Values("Authorization")
	if len(values) != 1 || !strings.HasPrefix(values[0], "Bearer "+CLIAgentReadTokenPrefix) || len(r.Header.Values(aiAgentTokenHeader)) != 0 || len(r.Header.Values(deviceIDHeader)) != 0 || len(r.Header.Values(deviceSecretHeader)) != 0 {
		return "", false
	}
	token := strings.TrimPrefix(values[0], "Bearer ")
	random := strings.TrimPrefix(token, CLIAgentReadTokenPrefix)
	raw, err := base64.RawURLEncoding.DecodeString(random)
	return token, len(random) == 43 && err == nil && len(raw) == 32 && base64.RawURLEncoding.EncodeToString(raw) == random
}
