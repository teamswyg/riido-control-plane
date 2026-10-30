package riidoaiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"
)

func (s Server) daemonAssignmentSnapshot(ctx context.Context, deviceID, secret string, signals AssignmentNotificationStore) (daemonAssignmentEvent, map[string]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	event, err := s.daemonAssignmentBindings(ctx, deviceID, secret)
	if err != nil {
		return daemonAssignmentEvent{}, nil, err
	}
	versions, err := signals.AssignmentSignalVersions(ctx, event.AgentIDs)
	return event, versions, err
}

func (s Server) daemonAssignmentBindings(ctx context.Context, deviceID, secret string) (daemonAssignmentEvent, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	principal, err := s.devices.AuthorizeDeviceCredential(ctx, deviceID, secret, AuthorizationRequest{Resource: AuthorizationResourceAgent, Action: AuthorizationActionRead, DeviceID: deviceID})
	if err != nil {
		return daemonAssignmentEvent{}, err
	}
	bindings, err := s.daemonRuntime.ListAIAgentDaemonAgentBindings(ctx, principal, deviceID)
	if err != nil {
		return daemonAssignmentEvent{}, err
	}
	event := daemonAssignmentEvent{SchemaVersion: daemonAssignmentEventsSchema, Bindings: bindings.Bindings, ConnectionRevision: bindings.ConnectionRevision, ConnectedPrincipalCount: bindings.ConnectedPrincipalCount, AgentIDs: []string{}}
	for _, binding := range event.Bindings {
		event.AgentIDs = append(event.AgentIDs, binding.AgentID)
	}
	return event, nil
}

func changedDaemonAssignmentAgents(before, after daemonAssignmentEvent, old, next map[string]string) []string {
	bindings := map[string]AgentRuntimeBinding{}
	for _, b := range before.Bindings {
		bindings[b.AgentID] = b
	}
	changed := []string{}
	for _, b := range after.Bindings {
		previous, exists := bindings[b.AgentID]
		if !exists || !reflect.DeepEqual(previous, b) || old[b.AgentID] != next[b.AgentID] {
			changed = append(changed, b.AgentID)
		}
	}
	return changed
}

func writeDaemonAssignmentEvent(w http.ResponseWriter, event daemonAssignmentEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "event: assignments\ndata: %s\n\n", body); err != nil {
		return err
	}
	return http.NewResponseController(w).Flush()
}
