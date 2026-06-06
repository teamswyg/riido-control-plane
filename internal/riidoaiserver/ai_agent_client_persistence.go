package riidoaiserver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

type AIAgentClientSnapshotStore interface {
	LoadAIAgentClientSnapshot(ctx context.Context) (AIAgentClientSnapshot, bool, error)
	SaveAIAgentClientSnapshot(ctx context.Context, snapshot AIAgentClientSnapshot) error
}

type AIAgentClientSnapshot struct {
	SchemaVersion           string                                  `json:"schema_version"`
	SavedAt                 time.Time                               `json:"saved_at"`
	WorkspaceID             string                                  `json:"workspace_id"`
	Devices                 []DeviceRecord                          `json:"devices"`
	DeviceCredentials       []AIAgentClientDeviceCredentialSnapshot `json:"device_credentials"`
	Daemons                 []DeviceDaemonRecord                    `json:"daemons"`
	Agents                  []AgentClientRecord                     `json:"agents"`
	Fixtures                []AgentOnboardingFixture                `json:"fixtures"`
	TaskThreads             map[string][]AIAgentTaskThreadRecord    `json:"task_threads"`
	Events                  []AIAgentClientEventSnapshot            `json:"events"`
	NextDeviceCredentialSeq int                                     `json:"next_device_credential_seq"`
	NextDaemonCommand       int                                     `json:"next_daemon_command"`
}

type AIAgentClientDeviceCredentialSnapshot struct {
	DeviceID         string    `json:"device_id"`
	MachineID        string    `json:"machine_id,omitempty"`
	SecretHashSHA256 string    `json:"secret_hash_sha256"`
	OwnerPrincipalID string    `json:"owner_principal_id"`
	WorkspaceID      string    `json:"workspace_id"`
	DisplayName      string    `json:"display_name,omitempty"`
	IssuedAt         time.Time `json:"issued_at"`
}

type AIAgentClientEventSnapshot struct {
	Seq       int64           `json:"seq"`
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
}

type PersistentAIAgentClientStore struct {
	*DevelopmentAIAgentClientStore
	snapshotStore AIAgentClientSnapshotStore
}

func OpenPersistentAIAgentClientStore(ctx context.Context, base *DevelopmentAIAgentClientStore, snapshotStore AIAgentClientSnapshotStore) (*PersistentAIAgentClientStore, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if base == nil {
		base = NewDevelopmentAIAgentClientStore()
	}
	if snapshotStore == nil {
		return nil, errors.New("riidoaiserver: ai agent client snapshot store is required")
	}
	if snapshot, ok, err := snapshotStore.LoadAIAgentClientSnapshot(ctx); err != nil {
		return nil, err
	} else if ok {
		if err := base.restoreSnapshot(snapshot); err != nil {
			return nil, err
		}
		return &PersistentAIAgentClientStore{DevelopmentAIAgentClientStore: base, snapshotStore: snapshotStore}, nil
	}
	store := &PersistentAIAgentClientStore{DevelopmentAIAgentClientStore: base, snapshotStore: snapshotStore}
	if err := store.saveSnapshot(ctx); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *PersistentAIAgentClientStore) EnrollDeviceCredential(ctx context.Context, principal AuthorizationResult, workspaceID string, req EnrollDeviceRequest) (EnrollDeviceResponse, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return EnrollDeviceResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.EnrollDeviceCredential(ctx, principal, workspaceID, req)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}

func (s *PersistentAIAgentClientStore) AuthorizeDeviceCredential(ctx context.Context, deviceID, deviceSecret string, req AuthorizationRequest) (AuthorizationResult, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return AuthorizationResult{}, err
	}
	return s.DevelopmentAIAgentClientStore.AuthorizeDeviceCredential(ctx, deviceID, deviceSecret, req)
}

func (s *PersistentAIAgentClientStore) BootstrapAIAgentClient(ctx context.Context, principal AuthorizationResult, clientKind ClientKind) (ClientBootstrapResponse, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return ClientBootstrapResponse{}, err
	}
	return s.DevelopmentAIAgentClientStore.BootstrapAIAgentClient(ctx, principal, clientKind)
}

func (s *PersistentAIAgentClientStore) ListAIAgentOnboardingFixtures(ctx context.Context, principal AuthorizationResult) (AgentOnboardingFixtureListResponse, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return AgentOnboardingFixtureListResponse{}, err
	}
	return s.DevelopmentAIAgentClientStore.ListAIAgentOnboardingFixtures(ctx, principal)
}

func (s *PersistentAIAgentClientStore) CreateAIAgentFromOnboardingFixture(ctx context.Context, principal AuthorizationResult, fixtureID string, req CreateAgentConfigurationRequest) (AgentClientRecordResponse, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return AgentClientRecordResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.CreateAIAgentFromOnboardingFixture(ctx, principal, fixtureID, req)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}

func (s *PersistentAIAgentClientStore) ListAIAgentDevices(ctx context.Context, principal AuthorizationResult) (DeviceRuntimeListResponse, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return DeviceRuntimeListResponse{}, err
	}
	return s.DevelopmentAIAgentClientStore.ListAIAgentDevices(ctx, principal)
}

func (s *PersistentAIAgentClientStore) GetAIAgentDaemon(ctx context.Context, principal AuthorizationResult, agentID string) (DeviceDaemonDetailResponse, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return DeviceDaemonDetailResponse{}, err
	}
	return s.DevelopmentAIAgentClientStore.GetAIAgentDaemon(ctx, principal, agentID)
}

func (s *PersistentAIAgentClientStore) ControlAIAgentDaemon(ctx context.Context, principal AuthorizationResult, agentID string, action DaemonControlAction, req ControlDeviceDaemonRequest) (DeviceDaemonCommandResponse, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return DeviceDaemonCommandResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.ControlAIAgentDaemon(ctx, principal, agentID, action, req)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}

func (s *PersistentAIAgentClientStore) GetAIAgentDeviceDaemon(ctx context.Context, principal AuthorizationResult, deviceID string) (DeviceDaemonDetailResponse, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return DeviceDaemonDetailResponse{}, err
	}
	return s.DevelopmentAIAgentClientStore.GetAIAgentDeviceDaemon(ctx, principal, deviceID)
}

func (s *PersistentAIAgentClientStore) ControlAIAgentDeviceDaemon(ctx context.Context, principal AuthorizationResult, deviceID string, action DaemonControlAction, req ControlDeviceDaemonRequest) (DeviceDaemonCommandResponse, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return DeviceDaemonCommandResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.ControlAIAgentDeviceDaemon(ctx, principal, deviceID, action, req)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}

func (s *PersistentAIAgentClientStore) SyncAIAgentDaemonRuntimeSnapshot(ctx context.Context, principal AuthorizationResult, req DeviceRuntimeSnapshotSyncRequest) (DeviceRuntimeSnapshotSyncResponse, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return DeviceRuntimeSnapshotSyncResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.SyncAIAgentDaemonRuntimeSnapshot(ctx, principal, req)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}

func (s *PersistentAIAgentClientStore) ListAIAgentDaemonAgentBindings(ctx context.Context, principal AuthorizationResult, deviceID string) (AgentRuntimeBindingListResponse, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return AgentRuntimeBindingListResponse{}, err
	}
	return s.DevelopmentAIAgentClientStore.ListAIAgentDaemonAgentBindings(ctx, principal, deviceID)
}

func (s *PersistentAIAgentClientStore) ListAIAgentTaskAssignableAgents(ctx context.Context, principal AuthorizationResult, taskID string) (AgentClientListResponse, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return AgentClientListResponse{}, err
	}
	return s.DevelopmentAIAgentClientStore.ListAIAgentTaskAssignableAgents(ctx, principal, taskID)
}

func (s *PersistentAIAgentClientStore) ListWorkspaceAssignedAgentProfiles(ctx context.Context, principal AuthorizationResult) (AssignedAgentProfileMapResponse, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return AssignedAgentProfileMapResponse{}, err
	}
	return s.DevelopmentAIAgentClientStore.ListWorkspaceAssignedAgentProfiles(ctx, principal)
}

func (s *PersistentAIAgentClientStore) ListAIAgentTaskThreads(ctx context.Context, principal AuthorizationResult, taskID string) (AIAgentTaskThreadCollectionResponse, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return AIAgentTaskThreadCollectionResponse{}, err
	}
	return s.DevelopmentAIAgentClientStore.ListAIAgentTaskThreads(ctx, principal, taskID)
}

func (s *PersistentAIAgentClientStore) ReconcileAIAgentActiveThreadProjections(ctx context.Context, principal AuthorizationResult, taskID string, reader AssignmentProjectionReader) (bool, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return false, err
	}
	changed, err := s.DevelopmentAIAgentClientStore.ReconcileAIAgentActiveThreadProjections(ctx, principal, taskID, reader)
	if err != nil || !changed {
		return changed, err
	}
	return changed, s.saveSnapshot(ctx)
}

func (s *PersistentAIAgentClientStore) AssignAIAgentTask(ctx context.Context, principal AuthorizationResult, taskID string, req AssignAIAgentTaskRequest) (AIAgentTaskActionResponse, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.AssignAIAgentTask(ctx, principal, taskID, req)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}

func (s *PersistentAIAgentClientStore) CreateAIAgentTaskAgentAssignment(ctx context.Context, principal AuthorizationResult, taskID string, req AssignAIAgentTaskRequest) (AIAgentTaskActionResponse, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.CreateAIAgentTaskAgentAssignment(ctx, principal, taskID, req)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}

func (s *PersistentAIAgentClientStore) UnassignAIAgentTask(ctx context.Context, principal AuthorizationResult, taskID string, req UnassignAIAgentTaskRequest) (AIAgentTaskActionResponse, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.UnassignAIAgentTask(ctx, principal, taskID, req)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}

func (s *PersistentAIAgentClientStore) DeleteAIAgentTaskAgentAssignment(ctx context.Context, principal AuthorizationResult, taskID, agentID string, req AgentAssignmentActionRequest) (AIAgentTaskActionResponse, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.DeleteAIAgentTaskAgentAssignment(ctx, principal, taskID, agentID, req)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}

func (s *PersistentAIAgentClientStore) SubmitAIAgentTaskComment(ctx context.Context, principal AuthorizationResult, taskID string, req SubmitAIAgentTaskCommentRequest) (AIAgentTaskActionResponse, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.SubmitAIAgentTaskComment(ctx, principal, taskID, req)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}

func (s *PersistentAIAgentClientStore) CreateAIAgentTaskThreadMessage(ctx context.Context, principal AuthorizationResult, taskID, threadID string, req CreateAIAgentTaskThreadMessageRequest) (AIAgentTaskActionResponse, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.CreateAIAgentTaskThreadMessage(ctx, principal, taskID, threadID, req)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}

func (s *PersistentAIAgentClientStore) StopAIAgentTask(ctx context.Context, principal AuthorizationResult, taskID string, req StopAIAgentTaskRequest) (AIAgentTaskActionResponse, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.StopAIAgentTask(ctx, principal, taskID, req)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}

func (s *PersistentAIAgentClientStore) StopAIAgentTaskAgentAssignment(ctx context.Context, principal AuthorizationResult, taskID, agentID string, req AgentAssignmentActionRequest) (AIAgentTaskActionResponse, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.StopAIAgentTaskAgentAssignment(ctx, principal, taskID, agentID, req)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}

func (s *PersistentAIAgentClientStore) GetAIAgentTaskThreadStreamSubscription(ctx context.Context, principal AuthorizationResult, taskID string) (AIAgentTaskThreadStreamSubscriptionResponse, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return AIAgentTaskThreadStreamSubscriptionResponse{}, err
	}
	return s.DevelopmentAIAgentClientStore.GetAIAgentTaskThreadStreamSubscription(ctx, principal, taskID)
}

func (s *PersistentAIAgentClientStore) CreateAIAgent(ctx context.Context, principal AuthorizationResult, req CreateAgentConfigurationRequest) (AgentClientRecordResponse, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return AgentClientRecordResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.CreateAIAgent(ctx, principal, req)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}

func (s *PersistentAIAgentClientStore) GetAIAgentEditability(ctx context.Context, principal AuthorizationResult, agentID string) (AgentEditabilityResponse, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return AgentEditabilityResponse{}, err
	}
	return s.DevelopmentAIAgentClientStore.GetAIAgentEditability(ctx, principal, agentID)
}

func (s *PersistentAIAgentClientStore) UpdateAIAgentConfiguration(ctx context.Context, principal AuthorizationResult, agentID string, req UpdateAgentConfigurationRequest) (AgentClientRecordResponse, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return AgentClientRecordResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.UpdateAIAgentConfiguration(ctx, principal, agentID, req)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}

func (s *PersistentAIAgentClientStore) DeleteAIAgent(ctx context.Context, principal AuthorizationResult, agentID string) (DeleteAgentResponse, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return DeleteAgentResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.DeleteAIAgent(ctx, principal, agentID)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}

func (s *PersistentAIAgentClientStore) AIAgentClientEvents(ctx context.Context, principal AuthorizationResult) ([]ClientStreamEvent, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return nil, err
	}
	return s.DevelopmentAIAgentClientStore.AIAgentClientEvents(ctx, principal)
}

func (s *PersistentAIAgentClientStore) SubscribeAIAgentClientEvents(ctx context.Context, principal AuthorizationResult) ([]ClientStreamEvent, <-chan ClientStreamEvent, func(), error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return nil, nil, nil, err
	}
	return s.DevelopmentAIAgentClientStore.SubscribeAIAgentClientEvents(ctx, principal)
}

func (s *PersistentAIAgentClientStore) RecordAIAgentThreadProgress(ctx context.Context, agentID string, req AgentThreadProgressBatchRequest) (AgentThreadProgressBatchResponse, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return AgentThreadProgressBatchResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.RecordAIAgentThreadProgress(ctx, agentID, req)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}

func (s *PersistentAIAgentClientStore) RecordAIAgentAssignmentEvent(ctx context.Context, agentID string, req AgentEventRequest, event TaskEvent) error {
	if err := s.reloadSnapshot(ctx); err != nil {
		return err
	}
	if err := s.DevelopmentAIAgentClientStore.RecordAIAgentAssignmentEvent(ctx, agentID, req, event); err != nil {
		return err
	}
	return s.saveSnapshot(ctx)
}

func (s *PersistentAIAgentClientStore) reloadSnapshot(ctx context.Context) error {
	if s == nil || s.snapshotStore == nil || s.DevelopmentAIAgentClientStore == nil {
		return nil
	}
	snapshot, ok, err := s.snapshotStore.LoadAIAgentClientSnapshot(ctx)
	if err != nil || !ok {
		return err
	}
	return s.DevelopmentAIAgentClientStore.restoreSnapshotPreservingSubscribers(snapshot)
}

func (s *PersistentAIAgentClientStore) saveSnapshot(ctx context.Context) error {
	if s == nil || s.snapshotStore == nil || s.DevelopmentAIAgentClientStore == nil {
		return nil
	}
	snapshot, err := s.DevelopmentAIAgentClientStore.snapshot(time.Now().UTC())
	if err != nil {
		return err
	}
	return s.snapshotStore.SaveAIAgentClientSnapshot(ctx, snapshot)
}

func (s *DevelopmentAIAgentClientStore) snapshot(savedAt time.Time) (AIAgentClientSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	credentials := make([]AIAgentClientDeviceCredentialSnapshot, 0, len(s.deviceCredentials))
	for _, record := range s.deviceCredentials {
		credentials = append(credentials, AIAgentClientDeviceCredentialSnapshot{
			DeviceID:         record.deviceID,
			MachineID:        record.machineID,
			SecretHashSHA256: hex.EncodeToString(record.secretHash[:]),
			OwnerPrincipalID: record.ownerPrincipalID,
			WorkspaceID:      record.workspaceID,
			DisplayName:      record.displayName,
			IssuedAt:         record.issuedAt,
		})
	}
	sort.Slice(credentials, func(i, j int) bool { return credentials[i].DeviceID < credentials[j].DeviceID })

	daemons := make([]DeviceDaemonRecord, 0, len(s.daemons))
	for _, daemon := range s.daemons {
		daemons = append(daemons, copyDeviceDaemon(daemon))
	}
	sort.Slice(daemons, func(i, j int) bool { return daemons[i].DeviceID < daemons[j].DeviceID })

	agents := make([]AgentClientRecord, 0, len(s.agents))
	for _, agent := range s.agents {
		agents = append(agents, agent)
	}
	sort.Slice(agents, func(i, j int) bool { return agents[i].AgentID < agents[j].AgentID })

	taskThreads := make(map[string][]AIAgentTaskThreadRecord, len(s.taskThreads))
	for taskID, threads := range s.taskThreads {
		copied := make([]AIAgentTaskThreadRecord, len(threads))
		for i, thread := range threads {
			copied[i] = copyTaskThread(thread)
		}
		taskThreads[taskID] = copied
	}
	events, err := snapshotEvents(retainLatestClientReplayEvents(s.events))
	if err != nil {
		return AIAgentClientSnapshot{}, err
	}
	return AIAgentClientSnapshot{
		SchemaVersion:           AIAgentClientPersistenceSchemaVersion,
		SavedAt:                 savedAt.UTC(),
		WorkspaceID:             s.workspaceID,
		Devices:                 copyDevices(s.devices),
		DeviceCredentials:       credentials,
		Daemons:                 daemons,
		Agents:                  agents,
		Fixtures:                copyAgentOnboardingFixtures(s.fixtures),
		TaskThreads:             taskThreads,
		Events:                  events,
		NextDeviceCredentialSeq: s.nextDeviceCredentialSeq,
		NextDaemonCommand:       s.nextDaemonCommand,
	}, nil
}

func (s *DevelopmentAIAgentClientStore) restoreSnapshot(snapshot AIAgentClientSnapshot) error {
	return s.restoreSnapshotWithSubscriberMode(snapshot, false)
}

func (s *DevelopmentAIAgentClientStore) restoreSnapshotPreservingSubscribers(snapshot AIAgentClientSnapshot) error {
	return s.restoreSnapshotWithSubscriberMode(snapshot, true)
}

func (s *DevelopmentAIAgentClientStore) restoreSnapshotWithSubscriberMode(snapshot AIAgentClientSnapshot, preserveSubscribers bool) error {
	if snapshot.SchemaVersion != AIAgentClientPersistenceSchemaVersion {
		return fmt.Errorf("unsupported ai agent client snapshot schema_version %q", snapshot.SchemaVersion)
	}
	deviceCredentials := make(map[string]deviceCredentialRecord, len(snapshot.DeviceCredentials))
	for _, record := range snapshot.DeviceCredentials {
		deviceID := strings.TrimSpace(record.DeviceID)
		if deviceID == "" {
			return errors.New("ai agent client snapshot device credential device_id is required")
		}
		rawHash, err := hex.DecodeString(strings.TrimSpace(record.SecretHashSHA256))
		if err != nil || len(rawHash) != sha256.Size {
			return fmt.Errorf("ai agent client snapshot device credential %s secret_hash_sha256 is invalid", deviceID)
		}
		var hash [sha256.Size]byte
		copy(hash[:], rawHash)
		deviceCredentials[deviceID] = deviceCredentialRecord{
			deviceID:         deviceID,
			machineID:        strings.TrimSpace(record.MachineID),
			secretHash:       hash,
			ownerPrincipalID: strings.TrimSpace(record.OwnerPrincipalID),
			workspaceID:      strings.TrimSpace(record.WorkspaceID),
			displayName:      strings.TrimSpace(record.DisplayName),
			issuedAt:         record.IssuedAt,
		}
	}
	daemons := make(map[string]DeviceDaemonRecord, len(snapshot.Daemons))
	for _, daemon := range snapshot.Daemons {
		deviceID := strings.TrimSpace(daemon.DeviceID)
		if deviceID == "" {
			return errors.New("ai agent client snapshot daemon device_id is required")
		}
		daemons[deviceID] = copyDeviceDaemon(daemon)
	}
	agents := make(map[string]AgentClientRecord, len(snapshot.Agents))
	for _, agent := range snapshot.Agents {
		agentID := strings.TrimSpace(agent.AgentID)
		if agentID == "" {
			return errors.New("ai agent client snapshot agent_id is required")
		}
		agents[agentID] = agent
	}
	taskThreads := make(map[string][]AIAgentTaskThreadRecord, len(snapshot.TaskThreads))
	for taskID, threads := range snapshot.TaskThreads {
		taskID = strings.TrimSpace(taskID)
		if taskID == "" {
			return errors.New("ai agent client snapshot task thread task_id is required")
		}
		copied := make([]AIAgentTaskThreadRecord, len(threads))
		for i, thread := range threads {
			copied[i] = copyTaskThread(thread)
		}
		taskThreads[taskID] = copied
	}
	events, err := restoreSnapshotEvents(snapshot.Events)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	subscribers := s.subscribers
	nextSubscriberID := s.nextSubscriberID
	s.workspaceID = strings.TrimSpace(snapshot.WorkspaceID)
	s.devices = pruneLegacyRuntimeRecords(copyDevices(snapshot.Devices))
	s.deviceCredentials = deviceCredentials
	s.nextDeviceCredentialSeq = snapshot.NextDeviceCredentialSeq
	s.daemons = daemons
	s.nextDaemonCommand = snapshot.NextDaemonCommand
	s.agents = agents
	s.fixtures = copyAgentOnboardingFixtures(snapshot.Fixtures)
	s.ensureOnboardingFixtureColorsLocked()
	s.taskThreads = taskThreads
	s.events = events
	if preserveSubscribers {
		s.subscribers = subscribers
		s.nextSubscriberID = nextSubscriberID
	} else {
		s.subscribers = map[int]aiAgentClientSubscriber{}
		s.nextSubscriberID = 0
	}
	return nil
}

func snapshotEvents(events []ClientStreamEvent) ([]AIAgentClientEventSnapshot, error) {
	out := make([]AIAgentClientEventSnapshot, 0, len(events))
	for _, event := range events {
		payload, err := json.Marshal(event.Payload)
		if err != nil {
			return nil, fmt.Errorf("snapshot ai agent client event %s: %w", event.EventType, err)
		}
		out = append(out, AIAgentClientEventSnapshot{
			Seq:       event.Seq,
			EventType: event.EventType,
			Payload:   payload,
		})
	}
	return out, nil
}

func restoreSnapshotEvents(events []AIAgentClientEventSnapshot) ([]ClientStreamEvent, error) {
	out := make([]ClientStreamEvent, 0, len(events))
	for _, event := range events {
		payload, err := restoreSnapshotEventPayload(event.EventType, event.Payload)
		if err != nil {
			return nil, fmt.Errorf("restore ai agent client event %s: %w", event.EventType, err)
		}
		out = append(out, ClientStreamEvent{
			Seq:       event.Seq,
			EventType: event.EventType,
			Payload:   payload,
		})
	}
	return out, nil
}

func restoreSnapshotEventPayload(eventType string, raw json.RawMessage) (any, error) {
	switch eventType {
	case AgentClientEventDeviceRuntimeSnapshot:
		var payload DeviceRuntimeSnapshotEvent
		return payload, json.Unmarshal(raw, &payload)
	case AgentClientEventDeviceDaemonStatus:
		var payload DeviceDaemonStatusEvent
		return payload, json.Unmarshal(raw, &payload)
	case AgentClientEventEditabilityChanged:
		var payload AgentEditabilityChangedEvent
		return payload, json.Unmarshal(raw, &payload)
	case AgentClientEventWorkStatusChanged:
		var payload AgentWorkStatusChangedEvent
		return payload, json.Unmarshal(raw, &payload)
	case AgentClientEventThreadProgress:
		var payload AgentThreadProgressEvent
		return payload, json.Unmarshal(raw, &payload)
	default:
		var payload map[string]any
		return payload, json.Unmarshal(raw, &payload)
	}
}
