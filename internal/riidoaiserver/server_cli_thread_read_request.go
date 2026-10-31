package riidoaiserver

import (
	"encoding/base64"
	"net/http"
	"strings"
)

func cliThreadReadRequestToken(r *http.Request, req AuthorizationRequest) (string, bool) {
	if req.Resource != AuthorizationResourceAIAgentClient || req.Action != AuthorizationActionRead || !cliInstallationID(req.WorkspaceID) || !cliInstallationID(req.TaskID) || req.AgentID != "" || req.DeviceID != "" {
		return "", false
	}
	path := "/v3/client/workspaces/" + req.WorkspaceID + "/ai-agent/tasks/" + req.TaskID + "/threads"
	if r.Method != http.MethodGet || r.RequestURI != path || r.URL.Path != "/v3/client/ai-agent/tasks/"+req.TaskID+"/threads" || r.URL.RawPath != "" || r.URL.RawQuery != "" || r.URL.ForceQuery || r.Header.Get("Origin") != "" || r.Header.Get("Upgrade") != "" {
		return "", false
	}
	values := r.Header.Values("Authorization")
	if len(values) != 1 || !strings.HasPrefix(values[0], "Bearer "+CLIThreadReadTokenPrefix) || len(r.Header.Values(aiAgentTokenHeader)) != 0 || len(r.Header.Values(deviceIDHeader)) != 0 || len(r.Header.Values(deviceSecretHeader)) != 0 {
		return "", false
	}
	token := strings.TrimPrefix(values[0], "Bearer ")
	random := strings.TrimPrefix(token, CLIThreadReadTokenPrefix)
	if len(random) != 43 {
		return "", false
	}
	raw, err := base64.RawURLEncoding.DecodeString(random)
	return token, err == nil && len(raw) == 32 && base64.RawURLEncoding.EncodeToString(raw) == random
}
