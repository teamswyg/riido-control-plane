package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAgentDashboardMetadataUsesCurrentGraphQLAndViewerScope(t *testing.T) {
	for _, mode := range []string{"success", "wrong_workspace", "wrong_task", "forbidden", "unavailable"} {
		t.Run(mode, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != http.MethodPost || r.URL.Path != "/graphql" || r.Header.Get("Authorization") != "Bearer viewer" {
					t.Errorf("unexpected metadata request")
				}
				var body struct {
					Query     string            `json:"query"`
					Variables map[string]string `json:"variables"`
				}
				_ = json.NewDecoder(r.Body).Decode(&body)
				if body.Variables["id"] != "task" {
					t.Error("task identity changed")
				}
				if strings.Contains(body.Query, "AgentDashboardLocation") {
					workspace := "workspace"
					if mode == "wrong_workspace" {
						workspace = "other"
					}
					_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"getComponentWorkspace": map[string]any{"id": "task", "componentType": "task", "team": map[string]any{"id": "team", "workspace": map[string]string{"id": workspace}}}}})
					return
				}
				if body.Variables["team"] != "team" || !strings.Contains(body.Query, `getDocument: "false"`) {
					t.Error("detail must use verified team without document")
				}
				if mode == "forbidden" {
					_, _ = w.Write([]byte(`{"errors":[{"extensions":{"code":"FORBIDDEN"}}]}`))
					return
				}
				if mode == "unavailable" {
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}
				id := "task"
				if mode == "wrong_task" {
					id = "other"
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"getComponent": map[string]string{"id": id, "componentType": "task", "title": "실제 업무", "keyNumber": "RIID-5830"}}})
			}))
			defer server.Close()
			client, err := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: server.URL})
			if err != nil {
				t.Fatal(err)
			}
			task, err := client.GetAgentDashboardTask(context.Background(), AIAgentTaskContextRequest{ComponentID: "task", WorkspaceID: "workspace", BearerToken: "viewer"})
			switch mode {
			case "success":
				if err != nil || task.Component.Title != "실제 업무" || task.Component.KeyNumber != "RIID-5830" {
					t.Fatalf("metadata=%+v err=%v", task, err)
				}
			case "unavailable":
				if err == nil || errors.Is(err, ErrAuthorizationForbidden) {
					t.Fatalf("outage must fail, not hide task: %v", err)
				}
			default:
				if !errors.Is(err, ErrAuthorizationForbidden) {
					t.Fatalf("must reject mismatched/inaccessible task: %v", err)
				}
			}
			if mode == "wrong_workspace" && calls != 1 {
				t.Fatal("must not fetch detail outside workspace")
			}
		})
	}
}
