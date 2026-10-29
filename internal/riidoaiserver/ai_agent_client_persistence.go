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
	"sync"
	"time"
)

const (
	defaultAIAgentClientSnapshotReloadInterval        = time.Minute
	defaultAIAgentClientHeartbeatSnapshotSaveInterval = 5 * time.Minute
)

type AIAgentClientSnapshotStore interface {
	LoadAIAgentClientSnapshot(ctx context.Context) (AIAgentClientSnapshot, bool, error)
	SaveAIAgentClientSnapshot(ctx context.Context, snapshot AIAgentClientSnapshot) error
}

type AIAgentClientSnapshot struct {
	OnboardingAgentCreates      map[string]onboardingAgentCreateReceipt      `json:"onboarding_agent_creates,omitempty"`
	TaskRequestReceipts         map[string]AIAgentTaskActionResponse         `json:"task_request_receipts,omitempty"`
	Version                     int64                                        `json:"version,omitempty"`
	ExpectedVersion             int64                                        `json:"-"`
	SchemaVersion               string                                       `json:"schema_version"`
	SavedAt                     time.Time                                    `json:"saved_at"`
	WorkspaceID                 string                                       `json:"workspace_id"`
	Devices                     []DeviceRecord                               `json:"devices"`
	DeviceCredentials           []AIAgentClientDeviceCredentialSnapshot      `json:"device_credentials"`
	DeviceConnectionGrants      []DeviceConnectionGrant                      `json:"device_connection_grants"`
	Daemons                     []DeviceDaemonRecord                         `json:"daemons"`
	Agents                      []AgentClientRecord                          `json:"agents"`
	OnboardingCreatedIDs        []string                                     `json:"onboarding_created_agent_ids,omitempty"`
	OnboardingCompletions       map[string]string                            `json:"onboarding_completed_tasks,omitempty"`
	OnboardingCompletionThreads map[string]AIAgentTaskThreadRecord           `json:"onboarding_completion_threads,omitempty"`
	Fixtures                    []AgentOnboardingFixture                     `json:"fixtures"`
	TaskAgentMemberships        map[string]map[string]AIAgentTaskMembership  `json:"task_agent_memberships"`
	TaskThreads                 map[string][]AIAgentTaskThreadRecord         `json:"task_threads"`
	TaskThreadMessages          map[string][]AIAgentTaskThreadHistoryMessage `json:"task_thread_messages,omitempty"`
	Events                      []AIAgentClientEventSnapshot                 `json:"events"`
	NextDeviceCredentialSeq     int                                          `json:"next_device_credential_seq"`
	NextDaemonCommand           int                                          `json:"next_daemon_command"`
}

type AIAgentClientEventSnapshot struct {
	Seq       int64           `json:"seq"`
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
}

type PersistentAIAgentClientStore struct {
	*DevelopmentAIAgentClientStore
	snapshotStore                 AIAgentClientSnapshotStore
	operationMu                   sync.Mutex
	version                       int64
	snapshotReloadMu              sync.Mutex
	snapshotReloadInterval        time.Duration
	snapshotHeartbeatSaveInterval time.Duration
	snapshotLastReloadAt          time.Time
	snapshotLastSavedAt           time.Time
	snapshotClock                 func() time.Time
	snapshotReloadRequired        bool
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
	store := &PersistentAIAgentClientStore{
		DevelopmentAIAgentClientStore: base,
		snapshotStore:                 snapshotStore,
		snapshotReloadInterval:        defaultAIAgentClientSnapshotReloadInterval,
		snapshotHeartbeatSaveInterval: defaultAIAgentClientHeartbeatSnapshotSaveInterval,
		snapshotClock:                 time.Now,
	}
	if snapshot, ok, err := snapshotStore.LoadAIAgentClientSnapshot(ctx); err != nil {
		return nil, err
	} else if ok {
		if err := base.restoreSnapshot(snapshot); err != nil {
			return nil, err
		}
		profilePruned := base.pruneUnexpectedDaemonProfiles("snapshot_restore")
		repaired := base.repairStaleActiveTaskThreads(store.snapshotNow().UTC())
		store.version = snapshot.Version
		store.markSnapshotObserved(snapshot.SavedAt)
		if repaired || profilePruned {
			return store, store.saveSnapshot(ctx)
		}
		return store, nil
	}
	if err := store.saveSnapshot(ctx); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *PersistentAIAgentClientStore) ConfigureSnapshotCadence(reloadInterval, heartbeatSaveInterval time.Duration) error {
	if s == nil {
		return nil
	}
	if reloadInterval < 0 {
		return errors.New("riidoaiserver: AI Agent client snapshot reload interval must not be negative")
	}
	if heartbeatSaveInterval < 0 {
		return errors.New("riidoaiserver: AI Agent client heartbeat snapshot save interval must not be negative")
	}
	s.snapshotReloadMu.Lock()
	defer s.snapshotReloadMu.Unlock()
	if reloadInterval > 0 {
		s.snapshotReloadInterval = reloadInterval
	}
	if heartbeatSaveInterval > 0 {
		s.snapshotHeartbeatSaveInterval = heartbeatSaveInterval
	}
	return nil
}

func (s *PersistentAIAgentClientStore) reloadSnapshot(ctx context.Context) error {
	if s == nil || s.snapshotStore == nil || s.DevelopmentAIAgentClientStore == nil {
		return nil
	}
	s.snapshotReloadMu.Lock()
	if !s.snapshotReloadRequired && !s.shouldReloadSnapshotLocked() {
		s.snapshotReloadMu.Unlock()
		return nil
	}
	snapshot, ok, err := s.snapshotStore.LoadAIAgentClientSnapshot(ctx)
	if err != nil || !ok {
		if err == nil && s.snapshotReloadRequired {
			err = errAIAgentSnapshotConflict
		}
		s.snapshotReloadMu.Unlock()
		return err
	}
	s.version = snapshot.Version
	if err := s.restoreSnapshotPreservingSubscribers(snapshot); err != nil {
		s.snapshotReloadMu.Unlock()
		return err
	}
	repaired := s.repairStaleActiveTaskThreads(s.snapshotNow().UTC())
	s.snapshotReloadRequired = false
	s.snapshotLastReloadAt = s.snapshotNow()
	s.snapshotReloadMu.Unlock()
	if repaired {
		return s.saveSnapshot(ctx)
	}
	return nil
}

func (s *PersistentAIAgentClientStore) saveSnapshot(ctx context.Context) error {
	if s == nil || s.snapshotStore == nil || s.DevelopmentAIAgentClientStore == nil {
		return nil
	}
	s.snapshotReloadMu.Lock()
	required := s.snapshotReloadRequired
	s.snapshotReloadMu.Unlock()
	if required {
		return errAIAgentSnapshotConflict
	}
	snapshot, err := s.snapshot(s.snapshotNow().UTC())
	if err != nil {
		return err
	}
	snapshot.ExpectedVersion = s.version
	snapshot.Version = s.version + 1
	ctx = context.WithValue(ctx, taskAgentPreserveMembershipContextKey{}, true)
	if err := s.snapshotStore.SaveAIAgentClientSnapshot(ctx, snapshot); err != nil {
		return err
	}
	s.version = snapshot.Version
	s.markSnapshotObserved(snapshot.SavedAt)
	return nil
}

func (s *PersistentAIAgentClientStore) shouldSaveHeartbeatSnapshot() bool {
	if s == nil {
		return true
	}
	s.snapshotReloadMu.Lock()
	defer s.snapshotReloadMu.Unlock()
	if s.snapshotHeartbeatSaveInterval <= 0 || s.snapshotLastSavedAt.IsZero() {
		return true
	}
	return !s.snapshotNow().Before(s.snapshotLastSavedAt.Add(s.snapshotHeartbeatSaveInterval))
}

func (s *PersistentAIAgentClientStore) shouldReloadSnapshotLocked() bool {
	if s.snapshotReloadInterval <= 0 || s.snapshotLastReloadAt.IsZero() {
		return true
	}
	return !s.snapshotNow().Before(s.snapshotLastReloadAt.Add(s.snapshotReloadInterval))
}

func (s *PersistentAIAgentClientStore) markSnapshotObserved(savedAt time.Time) {
	if s == nil {
		return
	}
	s.snapshotReloadMu.Lock()
	defer s.snapshotReloadMu.Unlock()
	now := s.snapshotNow()
	s.snapshotLastReloadAt = now
	if !savedAt.IsZero() {
		s.snapshotLastSavedAt = savedAt.UTC()
		return
	}
	s.snapshotLastSavedAt = now
}

func (s *PersistentAIAgentClientStore) snapshotNow() time.Time {
	if s != nil && s.snapshotClock != nil {
		return s.snapshotClock()
	}
	return time.Now()
}

func (s *DevelopmentAIAgentClientStore) snapshot(savedAt time.Time) (AIAgentClientSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.taskAgentMemberships == nil {
		s.taskAgentMemberships = legacyTaskAgentMemberships(s.taskThreads)
	}
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
	sort.Slice(daemons, func(i, j int) bool {
		if daemons[i].DeviceID != daemons[j].DeviceID {
			return daemons[i].DeviceID < daemons[j].DeviceID
		}
		if daemons[i].Profile != daemons[j].Profile {
			return daemons[i].Profile < daemons[j].Profile
		}
		return daemons[i].DaemonID < daemons[j].DaemonID
	})

	agents := make([]AgentClientRecord, 0, len(s.agents))
	for _, agent := range s.agents {
		agents = append(agents, agent)
	}
	sort.Slice(agents, func(i, j int) bool { return agents[i].AgentID < agents[j].AgentID })

	taskThreads := make(map[string][]AIAgentTaskThreadRecord, len(s.taskThreads))
	for taskID, threads := range s.taskThreads {
		// Keep only the most recent threads per task (the tail is chronological),
		// then cap each thread's progress lines. Both bounds keep the single
		// AI Agent client snapshot DynamoDB item under its 400 KB hard limit;
		// uncapped, accumulated runs/lines eventually make EVERY client write
		// (device heartbeat, thread progress, assignment events) fail with a
		// DynamoDB ValidationException, freezing the projection. The live SSE
		// stream is unaffected — it delivers every line in real time.
		retainedThreads := threads
		if len(retainedThreads) > aiAgentClientThreadsPerTaskLimit {
			retainedThreads = make([]AIAgentTaskThreadRecord, 0, len(threads))
			for index, thread := range threads {
				if index >= len(threads)-aiAgentClientThreadsPerTaskLimit || taskThreadNeedsRetention(thread) {
					retainedThreads = append(retainedThreads, thread)
				}
			}
		}
		copied := make([]AIAgentTaskThreadRecord, len(retainedThreads))
		for i, thread := range retainedThreads {
			copied[i] = copyTaskThread(thread)
			s.ensureTaskThreadAgentSnapshotLocked(&copied[i], savedAt)
			copied[i].Lines = retainLatestThreadProgressLines(copied[i].Lines)
		}
		taskThreads[taskID] = copied
	}
	taskThreadMessages := make(map[string][]AIAgentTaskThreadHistoryMessage, len(s.taskThreadMessages))
	for threadID, messages := range s.taskThreadMessages {
		taskThreadMessages[threadID] = retainLatestThreadHistoryMessages(messages)
	}
	events, err := snapshotEvents(retainLatestClientReplayEvents(s.events))
	if err != nil {
		return AIAgentClientSnapshot{}, err
	}
	return AIAgentClientSnapshot{
		SchemaVersion:               AIAgentClientPersistenceSchemaVersion,
		SavedAt:                     savedAt.UTC(),
		WorkspaceID:                 s.workspaceID,
		Devices:                     copyDevices(s.devices),
		DeviceCredentials:           credentials,
		DeviceConnectionGrants:      copyDeviceConnectionGrants(s.deviceConnectionGrants),
		Daemons:                     daemons,
		Agents:                      agents,
		Fixtures:                    copyAgentOnboardingFixtures(s.fixtures),
		TaskThreads:                 taskThreads,
		TaskRequestReceipts:         cloneTaskRequestReceipts(s.taskRequestReceipts),
		OnboardingAgentCreates:      cloneOnboardingAgentCreates(s.onboardingAgentCreates),
		OnboardingCreatedIDs:        onboardingCreatedAgentIDsSnapshot(s.onboardingCreated),
		OnboardingCompletions:       cloneOnboardingCompletedTasks(s.onboardingCompletions),
		OnboardingCompletionThreads: cloneOnboardingCompletionThreads(s.onboardingCompletionThreads),
		TaskAgentMemberships:        copyTaskAgentMemberships(s.taskAgentMemberships),
		TaskThreadMessages:          taskThreadMessages,
		Events:                      events,
		NextDeviceCredentialSeq:     s.nextDeviceCredentialSeq,
		NextDaemonCommand:           s.nextDaemonCommand,
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
	deviceConnectionGrants := make(map[string]map[string]DeviceConnectionGrant)
	for _, grant := range snapshot.DeviceConnectionGrants {
		grant.DeviceID = strings.TrimSpace(grant.DeviceID)
		grant.PrincipalID = strings.TrimSpace(grant.PrincipalID)
		grant.WorkspaceID = strings.TrimSpace(grant.WorkspaceID)
		if grant.DeviceID == "" || grant.PrincipalID == "" || grant.WorkspaceID == "" {
			return errors.New("ai agent client snapshot device connection grant identity is required")
		}
		if grant.ConnectedAt.IsZero() {
			grant.ConnectedAt = grant.LastSeenAt
		}
		if grant.LastSeenAt.IsZero() {
			grant.LastSeenAt = grant.ConnectedAt
		}
		deviceGrants := deviceConnectionGrants[grant.DeviceID]
		if deviceGrants == nil {
			deviceGrants = map[string]DeviceConnectionGrant{}
			deviceConnectionGrants[grant.DeviceID] = deviceGrants
		}
		deviceGrants[deviceConnectionGrantKey(grant.DeviceID, grant.PrincipalID, grant.WorkspaceID)] = grant
	}
	// Older v2 snapshots predate explicit account connection grants. The
	// credential row proves the original account/workspace connection, so seed
	// that fact without changing the device owner or rotating its secret.
	if snapshot.DeviceConnectionGrants == nil {
		for _, credential := range deviceCredentials {
			if credential.ownerPrincipalID == "" || credential.workspaceID == "" {
				continue
			}
			key := deviceConnectionGrantKey(credential.deviceID, credential.ownerPrincipalID, credential.workspaceID)
			deviceGrants := deviceConnectionGrants[credential.deviceID]
			if deviceGrants == nil {
				deviceGrants = map[string]DeviceConnectionGrant{}
				deviceConnectionGrants[credential.deviceID] = deviceGrants
			}
			if _, ok := deviceGrants[key]; ok {
				continue
			}
			deviceGrants[key] = DeviceConnectionGrant{
				DeviceID:    credential.deviceID,
				PrincipalID: credential.ownerPrincipalID,
				WorkspaceID: credential.workspaceID,
				ConnectedAt: credential.issuedAt,
				LastSeenAt:  credential.issuedAt,
			}
		}
	}
	daemons := make(map[string]DeviceDaemonRecord, len(snapshot.Daemons))
	for _, daemon := range snapshot.Daemons {
		deviceID := strings.TrimSpace(daemon.DeviceID)
		if deviceID == "" {
			return errors.New("ai agent client snapshot daemon device_id is required")
		}
		daemons[daemonStorageKey(daemon)] = copyDeviceDaemon(daemon)
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
	taskThreadMessages := make(map[string][]AIAgentTaskThreadHistoryMessage, len(snapshot.TaskThreadMessages))
	for threadID, messages := range snapshot.TaskThreadMessages {
		threadID = strings.TrimSpace(threadID)
		if threadID == "" {
			return errors.New("ai agent client snapshot task thread message thread_id is required")
		}
		taskThreadMessages[threadID] = retainLatestThreadHistoryMessages(messages)
	}
	events, err := restoreSnapshotEvents(snapshot.Events)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	subscribers := s.subscribers
	nextSubscriberID := s.nextSubscriberID
	devices := pruneLegacyRuntimeRecords(copyDevices(snapshot.Devices))
	devices = backfillRuntimeProviderVersionsFromSeed(devices, s.devices)
	if preserveSubscribers {
		s.preserveRuntimeSnapshotLivenessLocked(devices, daemons, deviceCredentials)
	}
	s.workspaceID = strings.TrimSpace(snapshot.WorkspaceID)
	s.devices = devices
	s.deviceCredentials = deviceCredentials
	s.deviceConnectionGrants = deviceConnectionGrants
	s.nextDeviceCredentialSeq = snapshot.NextDeviceCredentialSeq
	s.daemons = daemons
	s.nextDaemonCommand = snapshot.NextDaemonCommand
	s.agents = agents
	s.fixtures = copyAgentOnboardingFixtures(snapshot.Fixtures)
	s.ensureOnboardingFixtureColorsLocked()
	s.taskThreads = taskThreads
	s.taskRequestReceipts = cloneTaskRequestReceipts(snapshot.TaskRequestReceipts)
	s.onboardingAgentCreates = cloneOnboardingAgentCreates(snapshot.OnboardingAgentCreates)
	s.onboardingCreated = onboardingCreatedAgentsFromSnapshot(snapshot.OnboardingCreatedIDs)
	backfillOnboardingCreatedFromFixtureTmpColor(s.onboardingCreated, agents)
	s.onboardingCompletions = cloneOnboardingCompletedTasks(snapshot.OnboardingCompletions)
	s.onboardingCompletionThreads = cloneOnboardingCompletionThreads(snapshot.OnboardingCompletionThreads)
	s.taskAgentMemberships = copyTaskAgentMemberships(snapshot.TaskAgentMemberships)
	if s.taskAgentMemberships == nil {
		s.taskAgentMemberships = legacyTaskAgentMemberships(taskThreads)
	}
	s.taskThreadMessages = taskThreadMessages
	s.ensureAgentProfileColorsLocked()
	s.taskThreadProgressCache = map[string]taskThreadProgressMessageCache{}
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
	case AgentClientEventTaskNotification:
		var payload AgentTaskNotificationEvent
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
