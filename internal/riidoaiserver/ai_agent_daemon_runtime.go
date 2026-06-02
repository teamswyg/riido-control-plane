package riidoaiserver

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

type AIAgentDaemonRuntimeStore interface {
	SyncAIAgentDaemonRuntimeSnapshot(ctx context.Context, principal AuthorizationResult, req DeviceRuntimeSnapshotSyncRequest) (DeviceRuntimeSnapshotSyncResponse, error)
	ListAIAgentDaemonAgentBindings(ctx context.Context, principal AuthorizationResult, deviceID string) (AgentRuntimeBindingListResponse, error)
}

func (s *DevelopmentAIAgentClientStore) SyncAIAgentDaemonRuntimeSnapshot(ctx context.Context, principal AuthorizationResult, req DeviceRuntimeSnapshotSyncRequest) (DeviceRuntimeSnapshotSyncResponse, error) {
	if err := ctx.Err(); err != nil {
		return DeviceRuntimeSnapshotSyncResponse{}, err
	}
	if s == nil {
		return DeviceRuntimeSnapshotSyncResponse{}, errors.New("ai agent client store is not configured")
	}
	principal.PrincipalID = strings.TrimSpace(principal.PrincipalID)
	principal.WorkspaceID = strings.TrimSpace(principal.WorkspaceID)
	req.DaemonID = strings.TrimSpace(req.DaemonID)
	req.DeviceID = strings.TrimSpace(req.DeviceID)
	req.DeviceDisplayName = strings.TrimSpace(req.DeviceDisplayName)
	req.Profile = strings.TrimSpace(req.Profile)
	if principal.PrincipalID == "" {
		return DeviceRuntimeSnapshotSyncResponse{}, errors.New("principal_id is required")
	}
	if req.DeviceID == "" {
		return DeviceRuntimeSnapshotSyncResponse{}, errors.New("device_id is required")
	}
	if req.DaemonID == "" {
		return DeviceRuntimeSnapshotSyncResponse{}, errors.New("daemon_id is required")
	}
	runtimes, err := normalizeRuntimeSnapshotRecords(req.DeviceID, principal.PrincipalID, req.Runtimes)
	if err != nil {
		return DeviceRuntimeSnapshotSyncResponse{}, err
	}
	now := time.Now().UTC()
	displayName := req.DeviceDisplayName
	if displayName == "" {
		displayName = "Riido Desktop"
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if credential, ok := s.deviceCredentials[req.DeviceID]; ok {
		if credential.ownerPrincipalID != principal.PrincipalID {
			return DeviceRuntimeSnapshotSyncResponse{}, ErrAuthorizationForbidden
		}
		if principal.WorkspaceID == "" {
			principal.WorkspaceID = credential.workspaceID
		}
		if displayName == "Riido Desktop" && credential.displayName != "" {
			displayName = credential.displayName
		}
	}
	if existing, ok := s.deviceByIDLocked(req.DeviceID); ok && req.DeviceDisplayName == "" && existing.DisplayName != "" {
		displayName = existing.DisplayName
	}
	for i := range runtimes {
		runtimes[i].HasAssignedAgent = s.runtimeHasAssignedAgentLocked(runtimes[i].RuntimeID)
	}
	device := DeviceRecord{
		DeviceID:         req.DeviceID,
		OwnerPrincipalID: principal.PrincipalID,
		DisplayName:      displayName,
		DaemonLastSeenAt: now,
		Runtimes:         runtimes,
	}
	s.upsertDeviceRuntimeSnapshotLocked(device)
	daemon := DeviceDaemonRecord{
		DeviceID:          req.DeviceID,
		OwnerPrincipalID:  principal.PrincipalID,
		DeviceDisplayName: displayName,
		DaemonID:          req.DaemonID,
		Profile:           req.Profile,
		LastSeenAt:        now,
		Availability:      DaemonAvailabilityOnline,
		ControlState:      DaemonControlStateIdle,
		SupportedActions:  []DaemonControlAction{DaemonControlActionRestart, DaemonControlActionStop},
	}
	if existing, ok := s.daemons[req.DeviceID]; ok {
		daemon.PID = existing.PID
		daemon.UptimeSeconds = existing.UptimeSeconds
		daemon.StartedAt = existing.StartedAt
		if daemon.Profile == "" {
			daemon.Profile = existing.Profile
		}
	}
	s.daemons[req.DeviceID] = daemon
	s.appendClientEventLocked(AgentClientEventDeviceRuntimeSnapshot, DeviceRuntimeSnapshotEvent{
		EventType:     AgentClientEventDeviceRuntimeSnapshot,
		SchemaVersion: SchemaVersion,
		Device:        copyDevice(device),
	})
	s.appendClientEventLocked(AgentClientEventDeviceDaemonStatus, DeviceDaemonStatusEvent{
		EventType:     AgentClientEventDeviceDaemonStatus,
		SchemaVersion: SchemaVersion,
		Daemon:        copyDeviceDaemon(daemon),
	})
	return DeviceRuntimeSnapshotSyncResponse{
		SchemaVersion: SchemaVersion,
		Device:        copyDevice(device),
		Daemon:        copyDeviceDaemon(daemon),
	}, nil
}

func (s *DevelopmentAIAgentClientStore) ListAIAgentDaemonAgentBindings(ctx context.Context, principal AuthorizationResult, deviceID string) (AgentRuntimeBindingListResponse, error) {
	if err := ctx.Err(); err != nil {
		return AgentRuntimeBindingListResponse{}, err
	}
	if s == nil {
		return AgentRuntimeBindingListResponse{}, errors.New("ai agent client store is not configured")
	}
	principal.PrincipalID = strings.TrimSpace(principal.PrincipalID)
	deviceID = strings.TrimSpace(deviceID)
	if principal.PrincipalID == "" {
		return AgentRuntimeBindingListResponse{}, errors.New("principal_id is required")
	}
	if deviceID == "" {
		return AgentRuntimeBindingListResponse{}, errors.New("device_id is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if credential, ok := s.deviceCredentials[deviceID]; ok {
		if credential.ownerPrincipalID != principal.PrincipalID {
			return AgentRuntimeBindingListResponse{}, ErrAuthorizationForbidden
		}
		if principal.WorkspaceID == "" {
			principal.WorkspaceID = credential.workspaceID
		}
	}
	device, ok := s.deviceByIDLocked(deviceID)
	if !ok || device.OwnerPrincipalID != principal.PrincipalID {
		return AgentRuntimeBindingListResponse{}, ErrAuthorizationForbidden
	}
	bindings := make([]AgentRuntimeBinding, 0, len(s.agents))
	for _, agent := range s.agents {
		if s.agentWorkspaceID(agent) != s.workspaceScope(principal) {
			continue
		}
		binding, ok := s.agentRuntimeBindingLocked(agent)
		if !ok || binding.DeviceID != deviceID {
			continue
		}
		bindings = append(bindings, binding)
	}
	sort.Slice(bindings, func(i, j int) bool {
		if bindings[i].RuntimeProvider != bindings[j].RuntimeProvider {
			return bindings[i].RuntimeProvider < bindings[j].RuntimeProvider
		}
		return bindings[i].AgentID < bindings[j].AgentID
	})
	return AgentRuntimeBindingListResponse{SchemaVersion: SchemaVersion, Bindings: bindings}, nil
}

func (s *DevelopmentAIAgentClientStore) LookupAgent(agentID string) (AgentRuntimeBinding, bool) {
	if s == nil {
		return AgentRuntimeBinding{}, false
	}
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return AgentRuntimeBinding{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	agent, ok := s.agents[agentID]
	if !ok {
		return AgentRuntimeBinding{}, false
	}
	return s.agentRuntimeBindingLocked(agent)
}

func (s *DevelopmentAIAgentClientStore) agentRuntimeBindingLocked(agent AgentClientRecord) (AgentRuntimeBinding, bool) {
	agent.AgentID = strings.TrimSpace(agent.AgentID)
	agent.RuntimeID = strings.TrimSpace(agent.RuntimeID)
	if agent.AgentID == "" || agent.RuntimeID == "" {
		return AgentRuntimeBinding{}, false
	}
	device, runtime, ok := s.deviceRuntimeByRuntimeIDLocked(agent.RuntimeID)
	if !ok {
		return AgentRuntimeBinding{}, false
	}
	daemon, ok := s.daemons[device.DeviceID]
	if !ok || strings.TrimSpace(daemon.DaemonID) == "" {
		return AgentRuntimeBinding{}, false
	}
	provider := runtimeProviderForAIAgentRuntime(agent.RuntimeKind)
	if provider == "" {
		provider = runtimeProviderForAIAgentRuntime(runtime.Kind)
	}
	if provider == "" {
		return AgentRuntimeBinding{}, false
	}
	return normalizeAgentRuntimeBinding(AgentRuntimeBinding{
		AgentID:         agent.AgentID,
		DaemonID:        daemon.DaemonID,
		DeviceID:        device.DeviceID,
		RuntimeID:       runtime.RuntimeID,
		RuntimeProvider: provider,
	}), true
}

func (s *DevelopmentAIAgentClientStore) upsertDeviceRuntimeSnapshotLocked(device DeviceRecord) {
	for i := range s.devices {
		if s.devices[i].DeviceID == device.DeviceID {
			s.devices[i] = copyDevice(device)
			return
		}
	}
	s.devices = append(s.devices, copyDevice(device))
}

func (s *DevelopmentAIAgentClientStore) deviceRuntimeByRuntimeIDLocked(runtimeID string) (DeviceRecord, RuntimeRecord, bool) {
	runtimeID = strings.TrimSpace(runtimeID)
	for _, device := range s.devices {
		for _, runtime := range device.Runtimes {
			if runtime.RuntimeID == runtimeID {
				return copyDevice(device), copyRuntime(runtime), true
			}
		}
	}
	return DeviceRecord{}, RuntimeRecord{}, false
}

func (s *DevelopmentAIAgentClientStore) runtimeHasAssignedAgentLocked(runtimeID string) bool {
	runtimeID = strings.TrimSpace(runtimeID)
	if runtimeID == "" {
		return false
	}
	for _, agent := range s.agents {
		if strings.TrimSpace(agent.RuntimeID) == runtimeID {
			return true
		}
	}
	return false
}

func normalizeRuntimeSnapshotRecords(deviceID, ownerPrincipalID string, in []RuntimeSnapshotRecord) ([]RuntimeRecord, error) {
	if len(in) == 0 {
		return nil, errors.New("runtimes is required")
	}
	out := make([]RuntimeRecord, 0, len(in))
	seen := map[string]struct{}{}
	now := time.Now().UTC()
	for i, runtime := range in {
		runtime.RuntimeID = strings.TrimSpace(runtime.RuntimeID)
		if runtime.RuntimeID == "" {
			return nil, fmt.Errorf("runtimes[%d].runtime_id is required", i)
		}
		if _, ok := seen[runtime.RuntimeID]; ok {
			return nil, fmt.Errorf("runtimes[%d].runtime_id duplicates %s", i, runtime.RuntimeID)
		}
		seen[runtime.RuntimeID] = struct{}{}
		runtime.Kind = normalizeRuntimeKind(runtime.Kind)
		if runtime.Kind == "" {
			return nil, fmt.Errorf("runtimes[%d].kind is required", i)
		}
		if runtime.Availability == "" {
			runtime.Availability = RuntimeAvailabilityOnline
		}
		if runtime.Availability != RuntimeAvailabilityOnline && runtime.Availability != RuntimeAvailabilityOffline {
			return nil, fmt.Errorf("runtimes[%d].availability is invalid", i)
		}
		if runtime.DetectionState == "" {
			runtime.DetectionState = RuntimeDetectionStateDetected
		}
		if runtime.DetectionState != RuntimeDetectionStateDetected &&
			runtime.DetectionState != RuntimeDetectionStateMissing &&
			runtime.DetectionState != RuntimeDetectionStateError {
			return nil, fmt.Errorf("runtimes[%d].detection_state is invalid", i)
		}
		out = append(out, RuntimeRecord{
			RuntimeID:        runtime.RuntimeID,
			DeviceID:         deviceID,
			Kind:             runtime.Kind,
			Availability:     runtime.Availability,
			DetectionState:   runtime.DetectionState,
			OwnerPrincipalID: ownerPrincipalID,
			LastDetectedAt:   now,
			Models:           normalizeRuntimeModels(runtime.Kind, runtime.Models),
		})
	}
	return out, nil
}

func normalizeRuntimeKind(kind RuntimeKind) RuntimeKind {
	switch RuntimeKind(strings.TrimSpace(string(kind))) {
	case RuntimeKindCodex:
		return RuntimeKindCodex
	case RuntimeKindClaudeCode, RuntimeKind("claude"):
		return RuntimeKindClaudeCode
	case RuntimeKindCursor:
		return RuntimeKindCursor
	case RuntimeKindOpenClaw:
		return RuntimeKindOpenClaw
	default:
		return ""
	}
}

func normalizeRuntimeModels(kind RuntimeKind, models []RuntimeModelRecord) []RuntimeModelRecord {
	out := make([]RuntimeModelRecord, 0, len(models))
	seen := map[string]struct{}{}
	hasDefault := false
	for _, model := range models {
		model.ModelID = strings.TrimSpace(model.ModelID)
		model.Label = strings.TrimSpace(model.Label)
		if model.ModelID == "" {
			continue
		}
		if _, ok := seen[model.ModelID]; ok {
			continue
		}
		seen[model.ModelID] = struct{}{}
		if model.Label == "" {
			model.Label = model.ModelID
		}
		if model.IsDefault {
			hasDefault = true
		}
		out = append(out, model)
	}
	if len(out) == 0 {
		return []RuntimeModelRecord{defaultRuntimeModel(kind)}
	}
	if !hasDefault {
		out[0].IsDefault = true
	}
	return out
}

func defaultRuntimeModel(kind RuntimeKind) RuntimeModelRecord {
	switch kind {
	case RuntimeKindCodex:
		return RuntimeModelRecord{ModelID: "codex-default", Label: "Codex 기본 모델", IsDefault: true}
	case RuntimeKindClaudeCode:
		return RuntimeModelRecord{ModelID: "claude-default", Label: "Claude Code 기본 모델", IsDefault: true}
	case RuntimeKindCursor:
		return RuntimeModelRecord{ModelID: "cursor-auto", Label: "Cursor Auto", IsDefault: true}
	case RuntimeKindOpenClaw:
		return RuntimeModelRecord{ModelID: "openclaw-default", Label: "OpenClaw 기본 모델", IsDefault: true}
	default:
		return RuntimeModelRecord{ModelID: "runtime-default", Label: "기본 모델", IsDefault: true}
	}
}

func runtimeProviderForAIAgentRuntime(kind RuntimeKind) string {
	switch normalizeRuntimeKind(kind) {
	case RuntimeKindCodex:
		return "codex"
	case RuntimeKindClaudeCode:
		return "claude"
	case RuntimeKindCursor:
		return "cursor"
	case RuntimeKindOpenClaw:
		return "openclaw"
	default:
		return ""
	}
}
