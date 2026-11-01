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
)

func (a *CLIThreadReadAuthorizer) Authorize(ctx context.Context, token string, req AuthorizationRequest) (AuthorizationResult, error) {
	if ctx.Err() != nil {
		return AuthorizationResult{}, ctx.Err()
	}
	if a == nil || a.client == nil {
		return AuthorizationResult{}, errCLIThreadReadUnavailable
	}
	if !validCLITaskReadRequest(token, req, a.contract) {
		return AuthorizationResult{}, ErrAuthorizationForbidden
	}
	body, err := json.Marshal(map[string]string{"ticket": token, "resource": a.resource, "purpose": a.contract.purpose, "workspace_id": req.WorkspaceID, "task_id": req.TaskID})
	if err != nil {
		return AuthorizationResult{}, errCLIThreadReadUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, a.origin+a.contract.path, bytes.NewReader(body))
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
	return validCLITaskReadRequest(token, req, cliThreadReadContract())
}

func validCLITaskReadRequest(token string, req AuthorizationRequest, contract cliTaskReadContract) bool {
	if req.Resource != contract.resource || req.Action != AuthorizationActionRead || (req.AgentID != "" && (!contract.allowAgentID || !cliThreadReadID(req.AgentID))) || req.DeviceID != "" || !cliThreadReadID(req.WorkspaceID) || !cliThreadReadID(req.TaskID) || !strings.HasPrefix(token, contract.prefix) {
		return false
	}
	random := strings.TrimPrefix(token, contract.prefix)
	if len(random) != 43 {
		return false
	}
	raw, err := base64.RawURLEncoding.DecodeString(random)
	return err == nil && len(raw) == 32 && base64.RawURLEncoding.EncodeToString(raw) == random
}
