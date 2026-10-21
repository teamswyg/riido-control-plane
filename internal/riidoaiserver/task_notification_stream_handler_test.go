package riidoaiserver

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type multiplexNotificationAuthorizer struct{}

func (multiplexNotificationAuthorizer) Authorize(_ context.Context, _ string, req AuthorizationRequest) (AuthorizationResult, error) {
	if !strings.HasPrefix(req.WorkspaceID, "allowed-") || req.TaskID == "revoked" {
		return AuthorizationResult{}, ErrAuthorizationForbidden
	}
	return AuthorizationResult{PrincipalID: "owner", WorkspaceID: req.WorkspaceID}, nil
}

func TestNotificationMultiplexOneConnectionAndDeliveryScope(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	server := Server{aiAgent: store}
	server.config.Authorizer = multiplexNotificationAuthorizer{}
	httpServer := httptest.NewServer(http.HandlerFunc(server.handleAIAgentClientEvents))
	defer httpServer.Close()
	query := "?notifications=1"
	for i := 0; i < 9; i++ {
		query += fmt.Sprintf("&workspace_id=allowed-%d", i)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, httpServer.URL+query, nil)
	request.Header.Set("Authorization", "Bearer local-test")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatal(response.StatusCode)
	}
	store.mu.Lock()
	if len(store.subscribers) != 9 {
		t.Fatalf("subscriptions=%d", len(store.subscribers))
	}
	emit := func(workspace, recipient, task, id string) {
		store.appendClientEventLocked(AgentClientEventTaskNotification, AgentTaskNotificationEvent{EventType: AgentClientEventTaskNotification, EventID: id, WorkspaceID: workspace, RecipientPrincipalID: recipient, TaskID: task})
	}
	emit("allowed-0", "other", "task", "other-user")
	emit("outside", "owner", "task", "outside-workspace")
	emit("allowed-1", "owner", "revoked", "revoked-task")
	emit("allowed-8", "owner", "task", "visible")
	store.mu.Unlock()
	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data:") {
			if !strings.Contains(line, `"event_id":"visible"`) {
				t.Fatalf("unexpected delivery: %s", line)
			}
			cancel()
			return
		}
	}
	t.Fatalf("missing live notification: %v", scanner.Err())
}

func TestNotificationMultiplexRejectsUnapprovedWorkspace(t *testing.T) {
	server := Server{aiAgent: NewDevelopmentAIAgentClientStore()}
	server.config.Authorizer = multiplexNotificationAuthorizer{}
	request := httptest.NewRequest(http.MethodGet, "/?notifications=1&workspace_id=allowed-0&workspace_id=outside", nil)
	request.Header.Set("Authorization", "Bearer local-test")
	response := httptest.NewRecorder()
	server.handleAIAgentClientEvents(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status=%d", response.Code)
	}
}
