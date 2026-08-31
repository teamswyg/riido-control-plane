package riidoaiserver

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
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
			s.upsertDeviceConnectionGrantLocked(deviceID, principal, time.Now().UTC())
			revision, principalCount := s.deviceConnectionSummaryLocked(deviceID)
			log.Printf("riido_ai_agent_device event=account_connected device_id=%q workspace_id=%q connected_principal_count=%d connection_revision=%q", deviceID, principal.WorkspaceID, principalCount, revision)
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
	s.upsertDeviceConnectionGrantLocked(deviceID, principal, time.Now().UTC())
	s.devices = append(s.devices, device)
	return copyDevice(device), nil
}

func (s *DevelopmentAIAgentClientStore) SyncAIAgentDaemonRuntimeSnapshot(ctx context.Context, principal AuthorizationResult, req DeviceRuntimeSnapshotSyncRequest) (DeviceRuntimeSnapshotSyncResponse, error) {
	response, _, err := s.syncAIAgentDaemonRuntimeSnapshot(ctx, principal, req)
	return response, err
}

func (s *DevelopmentAIAgentClientStore) syncAIAgentDaemonRuntimeSnapshot(ctx context.Context, principal AuthorizationResult, req DeviceRuntimeSnapshotSyncRequest) (DeviceRuntimeSnapshotSyncResponse, bool, error) {
	if err := ctx.Err(); err != nil {
		return DeviceRuntimeSnapshotSyncResponse{}, false, err
	}
	if s == nil {
		return DeviceRuntimeSnapshotSyncResponse{}, false, errors.New("ai agent client store is not configured")
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
		return DeviceRuntimeSnapshotSyncResponse{}, false, errors.New("principal_id is required")
	}
	if req.DeviceID == "" {
		return DeviceRuntimeSnapshotSyncResponse{}, false, errors.New("device_id is required")
	}
	if req.DaemonID == "" {
		return DeviceRuntimeSnapshotSyncResponse{}, false, errors.New("daemon_id is required")
	}
	expectedProfile, changed, err := s.rejectUnexpectedDaemonProfileSnapshot(req)
	if err != nil {
		return DeviceRuntimeSnapshotSyncResponse{}, changed, err
	}
	runtimes, err := normalizeRuntimeSnapshotRecords(req.DeviceID, principal.PrincipalID, req.DaemonID, req.Profile, req.Runtimes)
	if err != nil {
		return DeviceRuntimeSnapshotSyncResponse{}, false, err
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
	profilePruned := s.pruneExpectedDaemonProfilesForSnapshotLocked(req.DeviceID, expectedProfile)
	if credential, ok := s.deviceCredentials[req.DeviceID]; ok {
		if credential.ownerPrincipalID != principal.PrincipalID {
			return DeviceRuntimeSnapshotSyncResponse{}, false, ErrAuthorizationForbidden
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
	previousDaemon, previousDaemonOK := s.daemonByIdentityLocked(req.DeviceID, req.Profile, req.DaemonID)
	daemon := DeviceDaemonRecord{
		DeviceID:          req.DeviceID,
		OwnerPrincipalID:  principal.PrincipalID,
		DeviceDisplayName: displayName,
		DaemonID:          req.DaemonID,
		Profile:           req.Profile,
		AppVersion:        strings.TrimSpace(req.AppVersion),
		PID:               req.PID,
		UptimeSeconds:     req.UptimeSeconds,
		StartedAt:         startedAt,
		LastSeenAt:        now,
		Availability:      DaemonAvailabilityOnline,
		ControlState:      DaemonControlStateIdle,
		SupportedActions:  []DaemonControlAction{DaemonControlActionRestart, DaemonControlActionStop},
	}
	if previousDaemonOK {
		if daemon.PID == 0 {
			daemon.PID = previousDaemon.PID
		}
		if daemon.UptimeSeconds == 0 {
			daemon.UptimeSeconds = previousDaemon.UptimeSeconds
		}
		if daemon.StartedAt.IsZero() {
			daemon.StartedAt = previousDaemon.StartedAt
		}
		if daemon.Profile == "" {
			daemon.Profile = previousDaemon.Profile
		}
		if daemon.AppVersion == "" {
			daemon.AppVersion = previousDaemon.AppVersion
		}
	}
	previousDevice, previousDeviceOK := s.deviceByIDLocked(req.DeviceID)
	if staleDaemonRuntimeSnapshot(previousDaemon, previousDaemonOK, daemon) {
		return DeviceRuntimeSnapshotSyncResponse{
			SchemaVersion: SchemaVersion,
			Device:        copyDevice(previousDevice),
			Daemon:        copyDeviceDaemon(previousDaemon),
		}, false, nil
	}
	device = s.upsertDeviceRuntimeSnapshotLocked(device)
	s.putDaemonLocked(daemon)
	// Only emit client stream events when something a viewer cares about actually
	// changed (runtimes, availability, connected workspaces, daemon status). A
	// plain liveness heartbeat (only last_seen_at/uptime advanced) is suppressed
	// so the SSE stream isn't flooded with identical snapshots every ~5s.
	deviceChanged := deviceRuntimeSnapshotChangedForEvent(previousDevice, previousDeviceOK, device)
	if deviceChanged {
		logProviderHealthChanges(previousDevice, previousDeviceOK, device)
		s.appendClientEventLocked(AgentClientEventDeviceRuntimeSnapshot, DeviceRuntimeSnapshotEvent{
			EventType:     AgentClientEventDeviceRuntimeSnapshot,
			SchemaVersion: SchemaVersion,
			Device:        copyDevice(device),
		})
	}
	daemonChanged := daemonStatusChangedForEvent(previousDaemon, previousDaemonOK, daemon)
	if daemonChanged {
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
	}, deviceChanged || daemonChanged || profilePruned.changed(), nil
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
	s.repairStaleActiveTaskThreadsLocked(time.Now().UTC())
	bindings := []AgentRuntimeBinding{}
	for _, agent := range s.agents {
		// The runtime binding is authoritative for daemon execution. Agent creation
		// already verifies that the caller can select the runtime, including an
		// owner using the same physical runtime from another workspace. Requiring
		// ConnectedWorkspaceIDs again here creates a split contract: creation
		// succeeds, but the daemon never sees the binding and work stays queued.
		binding, ok := s.agentRuntimeBindingLocked(agent)
		if !ok || binding.DeviceID != deviceID {
			continue
		}
		agent = s.projectAgentWorkStatusFromThreadsLocked(agent)
		s.agents[agent.AgentID] = agent
		bindings = append(bindings, binding)
	}
	sort.Slice(bindings, func(i, j int) bool {
		if bindings[i].RuntimeProvider != bindings[j].RuntimeProvider {
			return bindings[i].RuntimeProvider < bindings[j].RuntimeProvider
		}
		return bindings[i].AgentID < bindings[j].AgentID
	})
	revision, principalCount := s.deviceConnectionSummaryLocked(deviceID)
	return AgentRuntimeBindingListResponse{
		SchemaVersion:           SchemaVersion,
		Bindings:                bindings,
		ConnectionRevision:      revision,
		ConnectedPrincipalCount: principalCount,
	}, nil
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
			p.DaemonID != rt.DaemonID ||
			p.DaemonProfile != rt.DaemonProfile ||
			p.Availability != rt.Availability ||
			p.DetectionState != rt.DetectionState ||
			p.ProviderVersion != rt.ProviderVersion ||
			p.HealthStatus != rt.HealthStatus ||
			p.DiagnosticCode != rt.DiagnosticCode ||
			p.DiagnosticSummary != rt.DiagnosticSummary ||
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
		prev.AppVersion != next.AppVersion ||
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

func normalizeRuntimeSnapshotRecords(deviceID, ownerPrincipalID, daemonID, daemonProfile string, in []RuntimeSnapshotRecord) ([]RuntimeRecord, error) {
	if len(in) == 0 {
		return nil, errors.New("runtimes is required")
	}
	out := make([]RuntimeRecord, 0, len(in))
	seen := map[string]struct{}{}
	now := time.Now().UTC()
	daemonProfile, daemonID = normalizeDaemonRuntimeScope(daemonProfile, daemonID)
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
		healthStatus, diagnosticCode, diagnosticSummary, err := normalizeRuntimeProviderHealth(runtime)
		if err != nil {
			return nil, fmt.Errorf("runtimes[%d].provider_health: %w", i, err)
		}
		out = append(out, RuntimeRecord{
			RuntimeID:                 runtime.RuntimeID,
			DeviceID:                  deviceID,
			DaemonID:                  daemonID,
			DaemonProfile:             daemonProfile,
			Kind:                      runtime.Kind,
			Availability:              runtime.Availability,
			DetectionState:            runtime.DetectionState,
			ProviderVersion:           providerVersion,
			HealthStatus:              healthStatus,
			DiagnosticCode:            diagnosticCode,
			DiagnosticSummary:         diagnosticSummary,
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
		return RuntimeModelRecord{ModelID: providercatalog.DefaultCodexModelID, Label: "Codex 기본 모델", IsDefault: true}
	case RuntimeKindClaudeCode:
		return RuntimeModelRecord{ModelID: providercatalog.DefaultClaudeModelID, Label: "Claude Code 기본 모델", IsDefault: true}
	case RuntimeKindCursor:
		return RuntimeModelRecord{ModelID: providercatalog.DefaultCursorModelID, Label: "Cursor Auto", IsDefault: true}
	case RuntimeKindOpenClaw:
		return RuntimeModelRecord{ModelID: providercatalog.DefaultOpenClawModelID, Label: "OpenClaw 기본 모델", IsDefault: true}
	default:
		return RuntimeModelRecord{ModelID: providercatalog.DefaultRuntimeModelID, Label: "기본 모델", IsDefault: true}
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
