package riidoaiserver

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

type AIAgentDaemonRuntimeStore interface {
	SyncAIAgentDaemonRuntimeSnapshot(ctx context.Context, principal AuthorizationResult, req DeviceRuntimeSnapshotSyncRequest) (DeviceRuntimeSnapshotSyncResponse, error)
	ListAIAgentDaemonAgentBindings(ctx context.Context, principal AuthorizationResult, deviceID string) (AgentRuntimeBindingListResponse, error)
	ConnectAIAgentDevice(ctx context.Context, principal AuthorizationResult, machineID string) (DeviceRecord, error)
}

// ConnectAIAgentDevice connects the machine's device to the principal's
// workspace, so every member of that workspace can see the device and its
// runtimes. It is authorized by the user's token (proving workspace membership)
// and identifies the device by machine id (the same key the daemon enrolls
// under) — it does NOT issue or rotate the device secret, so a running daemon is
// unaffected. The device runtimes themselves are reported by the daemon; this
// only adds the workspace to the shared device row's connected set.
func (s *DevelopmentAIAgentClientStore) ConnectAIAgentDevice(ctx context.Context, principal AuthorizationResult, machineID string) (DeviceRecord, error) {
	if err := ctx.Err(); err != nil {
		return DeviceRecord{}, err
	}
	principal.PrincipalID = strings.TrimSpace(principal.PrincipalID)
	principal.WorkspaceID = strings.TrimSpace(principal.WorkspaceID)
	machineID = strings.TrimSpace(machineID)
	if principal.PrincipalID == "" {
		return DeviceRecord{}, errors.New("principal_id is required")
	}
	if principal.WorkspaceID == "" {
		return DeviceRecord{}, errors.New("workspace_id is required")
	}
	if machineID == "" {
		return DeviceRecord{}, errors.New("machine_id is required")
	}
	deviceID := deviceIDForMachine(machineID)

	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.devices {
		if s.devices[i].DeviceID == deviceID {
			s.devices[i].ConnectedWorkspaceIDs = addConnectedWorkspace(s.devices[i].ConnectedWorkspaceIDs, principal.WorkspaceID)
			return copyDevice(s.devices[i]), nil
		}
	}
	// The machine has not enrolled/reported yet; create a minimal connected row.
	// Runtimes fill in when the daemon reports its snapshot.
	device := DeviceRecord{
		DeviceID:              deviceID,
		OwnerPrincipalID:      principal.PrincipalID,
		DaemonLastSeenAt:      time.Now().UTC(),
		ConnectedWorkspaceIDs: []string{principal.WorkspaceID},
	}
	s.devices = append(s.devices, device)
	return copyDevice(device), nil
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
	previousDevice, previousDeviceOK := s.deviceByIDLocked(req.DeviceID)
	device = s.upsertDeviceRuntimeSnapshotLocked(device)
	previousDaemon, previousDaemonOK := s.daemons[req.DeviceID]
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
	// Only emit client stream events when something a viewer cares about actually
	// changed (runtimes, availability, connected workspaces, daemon status). A
	// plain liveness heartbeat (only last_seen_at/uptime advanced) is suppressed
	// so the SSE stream isn't flooded with identical snapshots every ~5s.
	if deviceRuntimeSnapshotChangedForEvent(previousDevice, previousDeviceOK, device) {
		s.appendClientEventLocked(AgentClientEventDeviceRuntimeSnapshot, DeviceRuntimeSnapshotEvent{
			EventType:     AgentClientEventDeviceRuntimeSnapshot,
			SchemaVersion: SchemaVersion,
			Device:        copyDevice(device),
		})
	}
	if daemonStatusChangedForEvent(previousDaemon, previousDaemonOK, daemon) {
		s.appendClientEventLocked(AgentClientEventDeviceDaemonStatus, DeviceDaemonStatusEvent{
			EventType:     AgentClientEventDeviceDaemonStatus,
			SchemaVersion: SchemaVersion,
			Daemon:        copyDeviceDaemon(daemon),
		})
	}
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
	scope := s.workspaceScope(principal)
	for _, agent := range s.agents {
		// A physical machine connects to many workspaces under one device
		// credential, so surface bindings for agents in ANY workspace this device
		// is connected to — not just the credential's enroll workspace. Otherwise
		// an agent assigned from another connected workspace is never polled and
		// its assignment stays queued forever. The binding.DeviceID guard below
		// still restricts the result to agents bound to THIS device's runtimes.
		agentWS := s.agentWorkspaceID(agent)
		if agentWS != scope && !deviceConnectedToWorkspace(device, agentWS) {
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

// deviceRuntimeSnapshotChangedForEvent reports whether anything a client cares
// about changed between the previously stored device and the new one. Volatile
// liveness fields (DaemonLastSeenAt, per-runtime LastDetectedAt) are ignored so
// a plain heartbeat does not emit a redundant snapshot event.
func deviceRuntimeSnapshotChangedForEvent(prev DeviceRecord, prevOK bool, next DeviceRecord) bool {
	if !prevOK {
		return true
	}
	if prev.OwnerPrincipalID != next.OwnerPrincipalID || prev.DisplayName != next.DisplayName {
		return true
	}
	if !equalStringSetIgnoreOrder(prev.ConnectedWorkspaceIDs, next.ConnectedWorkspaceIDs) {
		return true
	}
	if len(prev.Runtimes) != len(next.Runtimes) {
		return true
	}
	prevByID := make(map[string]RuntimeRecord, len(prev.Runtimes))
	for _, rt := range prev.Runtimes {
		prevByID[rt.RuntimeID] = rt
	}
	for _, rt := range next.Runtimes {
		p, ok := prevByID[rt.RuntimeID]
		if !ok {
			return true
		}
		if p.Kind != rt.Kind ||
			p.Availability != rt.Availability ||
			p.DetectionState != rt.DetectionState ||
			p.ProviderVersion != rt.ProviderVersion ||
			p.HasAssignedAgent != rt.HasAssignedAgent ||
			p.RequiresExperimentalOptIn != rt.RequiresExperimentalOptIn ||
			!equalRuntimeModelRecords(p.Models, rt.Models) {
			return true
		}
	}
	return false
}

// daemonStatusChangedForEvent ignores LastSeenAt/UptimeSeconds/StartedAt (which
// advance every heartbeat) and reports only meaningful daemon status changes.
func daemonStatusChangedForEvent(prev DeviceDaemonRecord, prevOK bool, next DeviceDaemonRecord) bool {
	if !prevOK {
		return true
	}
	if prev.Availability != next.Availability ||
		prev.ControlState != next.ControlState ||
		prev.Profile != next.Profile ||
		prev.PID != next.PID ||
		prev.DaemonID != next.DaemonID ||
		prev.DeviceDisplayName != next.DeviceDisplayName ||
		prev.OwnerPrincipalID != next.OwnerPrincipalID {
		return true
	}
	if len(prev.SupportedActions) != len(next.SupportedActions) {
		return true
	}
	for i := range prev.SupportedActions {
		if prev.SupportedActions[i] != next.SupportedActions[i] {
			return true
		}
	}
	return false
}

func equalRuntimeModelRecords(a, b []RuntimeModelRecord) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalStringSetIgnoreOrder(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := make(map[string]int, len(a))
	for _, v := range a {
		seen[v]++
	}
	for _, v := range b {
		if seen[v] == 0 {
			return false
		}
		seen[v]--
	}
	return true
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
		providerVersion, err := normalizeRuntimeProviderVersion(runtime.ProviderVersion)
		if err != nil {
			return nil, fmt.Errorf("runtimes[%d].provider_version: %w", i, err)
		}
		out = append(out, RuntimeRecord{
			RuntimeID:                 runtime.RuntimeID,
			DeviceID:                  deviceID,
			Kind:                      runtime.Kind,
			Availability:              runtime.Availability,
			DetectionState:            runtime.DetectionState,
			ProviderVersion:           providerVersion,
			OwnerPrincipalID:          ownerPrincipalID,
			LastDetectedAt:            now,
			RequiresExperimentalOptIn: runtime.RequiresExperimentalOptIn,
			Models:                    normalizeRuntimeModels(runtime.Kind, runtime.Models),
		})
	}
	return out, nil
}

func normalizeRuntimeProviderVersion(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	if utf8.RuneCountInString(value) > 128 {
		return "", errors.New("must be 128 characters or fewer")
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return "", errors.New("must not contain control characters")
		}
	}
	return value, nil
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
