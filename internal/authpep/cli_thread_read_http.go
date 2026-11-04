package authpep

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func (a *CLIThreadReadAuthorizer) Authorize(ctx context.Context, token string, req AuthorizationRequest) (AuthorizationResult, error) {
	if ctx.Err() != nil {
		return AuthorizationResult{}, ctx.Err()
	}
	if a == nil || a.client == nil {
		return AuthorizationResult{}, errCLIThreadReadUnavailable
	}
	if !validCLIThreadReadRequest(token, req) {
		return AuthorizationResult{}, ErrAuthorizationForbidden
	}
	body, err := json.Marshal(map[string]string{"ticket": token, "resource": a.resource, "purpose": cliThreadReadPurpose, "workspace_id": req.WorkspaceID, "task_id": req.TaskID})
	if err != nil {
		return AuthorizationResult{}, errCLIThreadReadUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, a.origin+"/internal/cli/thread-read/authorize", bytes.NewReader(body))
	if err != nil {
		return AuthorizationResult{}, errCLIThreadReadUnavailable
	}
	r.Header.Set("Authorization", "Bearer "+a.token)
	r.Header.Set("Content-Type", "application/json")
	response, err := a.client.Do(r)
	if err != nil {
		return AuthorizationResult{}, errCLIThreadReadUnavailable
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusForbidden {
		return AuthorizationResult{}, ErrAuthorizationForbidden
	}
	if response.StatusCode != http.StatusOK {
		return AuthorizationResult{}, errCLIThreadReadUnavailable
	}
	content, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil || content != "application/json" {
		return AuthorizationResult{}, errCLIThreadReadUnavailable
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, (8<<10)+1))
	if err != nil || len(raw) > 8<<10 || ctx.Err() != nil {
		return AuthorizationResult{}, errCLIThreadReadUnavailable
	}
	return a.decode(raw, req)
}

func validCLIThreadReadRequest(token string, req AuthorizationRequest) bool {
	if req.Resource != riidoaiserver.AuthorizationResourceCLIThreadRead || req.Action != AuthorizationActionRead || req.AgentID != "" || req.DeviceID != "" || !cliThreadReadID(req.WorkspaceID) || !cliThreadReadID(req.TaskID) || !strings.HasPrefix(token, riidoaiserver.CLIThreadReadTokenPrefix) {
		return false
	}
	random := strings.TrimPrefix(token, riidoaiserver.CLIThreadReadTokenPrefix)
	if len(random) != 43 {
		return false
	}
	raw, err := base64.RawURLEncoding.DecodeString(random)
	return err == nil && len(raw) == 32 && base64.RawURLEncoding.EncodeToString(raw) == random
}
