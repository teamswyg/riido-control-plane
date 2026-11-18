package riidoaiserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (s Server) readExplicitAssignment(ctx context.Context, a Assignment, req AgentAssignmentContextRequest) (explicitAssignmentResult, int, error) {
	empty := explicitAssignmentResult{}
	if !validExplicitOptions(req) {
		return empty, 400, errors.New("invalid explicit read options")
	}
	view := req.View
	if view == "threads" || view == "thread" {
		view = "authorize"
	}
	wire, err := json.Marshal(struct {
		Target  string            `json:"target"`
		View    string            `json:"view"`
		Options map[string]string `json:"options,omitempty"`
	}{req.Target, view, req.Options})
	if err != nil {
		return empty, http.StatusBadRequest, err
	}
	c := s.config.CommentDelivery
	endpoint := strings.TrimRight(c.EvidenceBaseURL, "/") + "/internal/agent-assignment-read/" + url.PathEscape(a.WorkspaceID) + "/" + url.PathEscape(a.TaskID) + "/" + url.PathEscape(a.CreatedBy)
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(wire))
	if err != nil {
		return empty, 503, err
	}
	r.Header.Set(CommentDeliveryKeyHeader, c.Key)
	r.Header.Set("Content-Type", "application/json")
	client := http.Client{Timeout: 8 * time.Second}
	if c.HTTPClient != nil {
		client = *c.HTTPClient
	}
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := client.Do(r)
	if err != nil {
		return empty, 503, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		status := explicitReadHTTPStatus(resp.StatusCode)
		return empty, status, errors.New("explicit source read failed")
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, (2<<20)+1))
	if err != nil || len(raw) > 2<<20 {
		return empty, 503, errors.New("explicit source response too large")
	}
	var v explicitAssignmentResult
	if json.Unmarshal(raw, &v) != nil || !validExplicitAssignmentLocator(raw, v) || v.WorkspaceID != a.WorkspaceID || v.RootID != a.TaskID || v.PrincipalID != a.CreatedBy || v.ID == "" || !json.Valid(v.Data) {
		return empty, 503, errors.New("explicit source provenance mismatch")
	}
	switch v.Kind {
	case "task", "subtask", "project", "milestone":
	default:
		return empty, 503, errors.New("invalid explicit source kind")
	}
	return v, 200, nil
}
