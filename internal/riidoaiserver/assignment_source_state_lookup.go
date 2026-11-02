package riidoaiserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

var agentSourceStateNames = map[string]string{"visible": AssignmentSourceUnselected, "deleted": AssignmentSourceDeleted, "hidden": AssignmentSourceHidden}

// lookup sends only workspace, requester and source identities. The answer
// must name the same sources in the same order with a known state.
func (config AgentSourceStateConfig) lookup(ctx context.Context, workspaceID, principalID string, refs []agentSourceStateRef) ([]AssignmentSourceState, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	payload, err := json.Marshal(map[string]any{"workspace_id": workspaceID, "principal_id": principalID, "sources": refs})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(config.BaseURL, "/")+"/internal/agent-source-states", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set(AgentSourceStateKeyHeader, config.Key)
	req.Header.Set("Content-Type", "application/json")
	client := config.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 3 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("source state request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("source state status " + http.StatusText(resp.StatusCode))
	}
	var body struct {
		States []struct {
			Kind   AssignmentContentKind `json:"kind"`
			ID     string                `json:"id"`
			TeamID string                `json:"team_id"`
			State  string                `json:"state"`
		} `json:"states"`
	}
	decoder := json.NewDecoder(io.LimitReader(resp.Body, 64<<10))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&body) != nil || !errors.Is(decoder.Decode(&struct{}{}), io.EOF) || len(body.States) != len(refs) {
		return nil, errors.New("invalid source state response")
	}
	out := make([]AssignmentSourceState, 0, len(refs))
	for i, state := range body.States {
		mapped := agentSourceStateNames[state.State]
		if state.Kind != refs[i].Kind || state.ID != refs[i].ID || state.TeamID != refs[i].TeamID || mapped == "" {
			return nil, errors.New("source state does not match the request")
		}
		out = append(out, AssignmentSourceState{ResourceKind: state.Kind, SourceID: state.ID, State: mapped})
	}
	return out, nil
}
