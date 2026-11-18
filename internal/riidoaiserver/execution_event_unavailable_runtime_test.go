package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestAdmittedExecutionReportsWhileRuntimeIsUnavailable(t *testing.T) {
	ctx := context.Background()
	registry := NewDevelopmentAIAgentClientStore()
	owner := AuthorizationResult{PrincipalID: "owner-user", WorkspaceID: "workspace-a"}
	enrollment := enrollTestDevice(t, ctx, registry, owner, owner.WorkspaceID, "machine-report", "Report Mac")
	runtimeID := enrollment.DeviceID + ":cursor"
	syncTestRuntime(t, ctx, registry, owner, enrollment.DeviceID, enrollment.DeviceID, runtimeID)
	created, err := registry.createAIAgent(ctx, owner, CreateAgentConfigurationRequest{Name: "Reporting agent", RuntimeID: runtimeID, Visibility: AgentVisibilityPrivate}, "")
	if err != nil {
		t.Fatal(err)
	}
	store := NewStoreWithConfig(StoreConfig{AgentRegistry: registry, OperationStore: &reportingOperationStore{}})
	defer store.Close()
	a, err := store.AssignTask(ctx, "task", AssignRequest{ComponentID: "component", WorkspaceID: owner.WorkspaceID, AgentID: created.Agent.AgentID, RuntimeProvider: "claude", Prompt: "fixture", CreatedBy: owner.PrincipalID})
	if err != nil {
		t.Fatal(err)
	}
	pollRequest := PollRequest{DaemonID: enrollment.DeviceID, DeviceID: enrollment.DeviceID, RuntimeID: runtimeID}
	poll, err := store.PollAgent(ctx, created.Agent.AgentID, pollRequest)
	if err != nil || poll.Action != PollStart {
		t.Fatalf("poll: %+v %v", poll, err)
	}
	registry.mu.Lock()
	for i := range registry.devices {
		for j := range registry.devices[i].Runtimes {
			if registry.devices[i].Runtimes[j].RuntimeID == runtimeID {
				registry.devices[i].Runtimes[j].Availability = RuntimeAvailabilityOffline
				registry.devices[i].Runtimes[j].DetectionState = RuntimeDetectionStateMissing
			}
		}
	}
	registry.mu.Unlock()
	if _, err = store.PollAgent(ctx, created.Agent.AgentID, pollRequest); !errors.Is(err, ErrAgentBindingValidation) {
		t.Fatalf("unavailable runtime admitted new work: %v", err)
	}
	server := Server{assignment: store, devices: registry}
	for _, changed := range []string{"device", "daemon", "runtime"} {
		request := AgentEventRequest{AssignmentID: a.ID, TaskID: a.TaskID, DaemonID: enrollment.DeviceID, DeviceID: enrollment.DeviceID, RuntimeID: runtimeID, EventType: EventRiidoLog}
		switch changed {
		case "device":
			request.DeviceID = "foreign"
		case "daemon":
			request.DaemonID = "foreign"
		case "runtime":
			request.RuntimeID = "foreign"
		}
		w := reportFixtureEvent(server, created.Agent.AgentID, enrollment.DeviceID, enrollment.DeviceSecret, request)
		if w.Code == http.StatusOK {
			t.Fatalf("reporting accepted a changed %s binding", changed)
		}
	}
	for _, event := range []string{EventAssignmentRunning, EventRiidoLog, EventAssignmentCompleted} {
		request := AgentEventRequest{AssignmentID: a.ID, TaskID: a.TaskID, DaemonID: enrollment.DeviceID, DeviceID: enrollment.DeviceID, RuntimeID: runtimeID, EventType: event, Message: "final answer"}
		if event == EventAssignmentRunning {
			request.State = AssignmentRunning
		}
		if event == EventAssignmentCompleted {
			request.State = AssignmentCompleted
		}
		w := reportFixtureEvent(server, created.Agent.AgentID, enrollment.DeviceID, enrollment.DeviceSecret, request)
		if w.Code != http.StatusOK {
			t.Fatalf("admitted %s rejected during probe change: %d %s", event, w.Code, w.Body.String())
		}
	}
}
