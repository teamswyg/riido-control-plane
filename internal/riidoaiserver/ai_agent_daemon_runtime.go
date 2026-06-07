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
	if req.PID < 0 {
		req.PID = 0
	}
	if req.UptimeSeconds < 0 {
		req.UptimeSeconds = 0
	}
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
	startedAt := req.StartedAt
	if !startedAt.IsZero() {
		startedAt = startedAt.UTC()
	}
	if startedAt.IsZero() && req.UptimeSeconds > 0 {
		startedAt = now.Add(-time.Duration(req.UptimeSeconds) * time.Second)
	}
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
	// Connect this device to the workspace the daemon is reporting from, so every
	// member of that workspace sees the device (workspace-connection visibility).
	if ws := strings.TrimSpace(principal.WorkspaceID); ws != "" {
		device.ConnectedWorkspaceIDs = []string{ws}
	}
	device = s.upsertDeviceRuntimeSnapshotLocked(device)
	daemon := DeviceDaemonRecord{
		DeviceID:          req.DeviceID,
		OwnerPrincipalID:  principal.PrincipalID,
		DeviceDisplayName: displayName,
		DaemonID:          req.DaemonID,
		Profile:           req.Profile,
		PID:               req.PID,
		UptimeSeconds:     req.UptimeSeconds,
		StartedAt:         startedAt,
		LastSeenAt:        now,
		Availability:      DaemonAvailabilityOnline,
		ControlState:      DaemonControlStateIdle,
		SupportedActions:  []DaemonControlAction{DaemonControlActionRestart, DaemonControlActionStop},
	}
	if existing, ok := s.daemons[req.DeviceID]; ok {
		if daemon.PID == 0 {
			daemon.PID = existing.PID
		}
		if daemon.UptimeSeconds == 0 {
			daemon.UptimeSeconds = existing.UptimeSeconds
		}
		if daemon.StartedAt.IsZero() {
			daemon.StartedAt = existing.StartedAt
		}
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

func (s *DevelopmentAIAgentClientStore) LookupAgentRuntimeFact(agentID string) (AgentRuntimeBinding, RuntimeRecord, bool) {
	if s == nil {
		return AgentRuntimeBinding{}, RuntimeRecord{}, false
	}
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return AgentRuntimeBinding{}, RuntimeRecord{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	agent, ok := s.agents[agentID]
	if !ok {
		return AgentRuntimeBinding{}, RuntimeRecord{}, false
	}
	binding, ok := s.agentRuntimeBindingLocked(agent)
	if !ok {
		return AgentRuntimeBinding{}, RuntimeRecord{}, false
	}
	device, runtime, ok := s.deviceRuntimeByRuntimeIDLocked(agent.RuntimeID)
	if !ok {
		return AgentRuntimeBinding{}, RuntimeRecord{}, false
	}
	projected := projectDeviceRuntimeLiveness(device, time.Now().UTC())
	if runtime, ok = runtimeByID(projected.Runtimes, agent.RuntimeID); !ok || !runtimeAvailableForBinding(runtime) {
		return AgentRuntimeBinding{}, RuntimeRecord{}, false
	}
	return binding, copyRuntime(runtime), true
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
	now := time.Now().UTC()
	device = projectDeviceRuntimeLiveness(device, now)
	if runtime, ok = runtimeByID(device.Runtimes, agent.RuntimeID); !ok || !runtimeAvailableForBinding(runtime) {
		return AgentRuntimeBinding{}, false
	}
	daemon, ok := s.daemons[device.DeviceID]
	if !ok || strings.TrimSpace(daemon.DaemonID) == "" {
		return AgentRuntimeBinding{}, false
	}
	daemon = projectDeviceDaemonLiveness(daemon, now)
	if daemon.Availability != DaemonAvailabilityOnline {
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

func runtimeByID(runtimes []RuntimeRecord, runtimeID string) (RuntimeRecord, bool) {
	runtimeID = strings.TrimSpace(runtimeID)
	for _, runtime := range runtimes {
		if strings.TrimSpace(runtime.RuntimeID) == runtimeID {
			return copyRuntime(runtime), true
		}
	}
	return RuntimeRecord{}, false
}

func runtimeAvailableForBinding(runtime RuntimeRecord) bool {
	return runtime.Availability == RuntimeAvailabilityOnline && runtime.DetectionState == RuntimeDetectionStateDetected
}

func (s *DevelopmentAIAgentClientStore) upsertDeviceRuntimeSnapshotLocked(device DeviceRecord) DeviceRecord {
	incomingRuntimeIDs := make(map[string]struct{}, len(device.Runtimes))
	for _, runtime := range device.Runtimes {
		if runtimeID := strings.TrimSpace(runtime.RuntimeID); runtimeID != "" {
			incomingRuntimeIDs[runtimeID] = struct{}{}
		}
	}
	if len(incomingRuntimeIDs) > 0 {
		for deviceIndex := range s.devices {
			if s.devices[deviceIndex].DeviceID == device.DeviceID {
				continue
			}
			filtered := s.devices[deviceIndex].Runtimes[:0]
			for _, runtime := range s.devices[deviceIndex].Runtimes {
				if _, moving := incomingRuntimeIDs[strings.TrimSpace(runtime.RuntimeID)]; moving {
					continue
				}
				filtered = append(filtered, runtime)
			}
			s.devices[deviceIndex].Runtimes = filtered
		}
	}
	for i := range s.devices {
		if s.devices[i].DeviceID == device.DeviceID {
			merged := copyDevice(s.devices[i])
			merged.OwnerPrincipalID = device.OwnerPrincipalID
			if device.DisplayName != "" {
				merged.DisplayName = device.DisplayName
			}
			merged.DaemonLastSeenAt = device.DaemonLastSeenAt
			for _, ws := range device.ConnectedWorkspaceIDs {
				merged.ConnectedWorkspaceIDs = addConnectedWorkspace(merged.ConnectedWorkspaceIDs, ws)
			}
			runtimeIndexByID := make(map[string]int, len(merged.Runtimes))
			for runtimeIndex, runtime := range merged.Runtimes {
				runtimeIndexByID[runtime.RuntimeID] = runtimeIndex
			}
			for _, runtime := range device.Runtimes {
				runtime = copyRuntime(runtime)
				if runtimeIndex, ok := runtimeIndexByID[runtime.RuntimeID]; ok {
					merged.Runtimes[runtimeIndex] = runtime
					continue
				}
				runtimeIndexByID[runtime.RuntimeID] = len(merged.Runtimes)
				merged.Runtimes = append(merged.Runtimes, runtime)
			}
			s.devices[i] = copyDevice(merged)
			return copyDevice(merged)
		}
	}
	s.devices = append(s.devices, copyDevice(device))
	return copyDevice(device)
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
			RuntimeID:                 runtime.RuntimeID,
			DeviceID:                  deviceID,
			Kind:                      runtime.Kind,
			Availability:              runtime.Availability,
			DetectionState:            runtime.DetectionState,
			OwnerPrincipalID:          ownerPrincipalID,
			LastDetectedAt:            now,
			RequiresExperimentalOptIn: runtime.RequiresExperimentalOptIn,
			Models:                    normalizeRuntimeModels(runtime.Kind, runtime.Models),
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
