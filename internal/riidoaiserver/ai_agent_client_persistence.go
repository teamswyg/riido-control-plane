package riidoaiserver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

const (
	aiAgentClientSnapshotPK = "AI_AGENT_CLIENT#development"
	aiAgentClientSnapshotSK = "STATE"
)

type AIAgentClientPersistence interface {
	LoadAIAgentClientSnapshot(ctx context.Context) (AIAgentClientSnapshot, bool, error)
	SaveAIAgentClientSnapshot(ctx context.Context, snapshot AIAgentClientSnapshot) error
	Close() error
}

type DevelopmentAIAgentClientStoreConfig struct {
	Persistence AIAgentClientPersistence
}

type AIAgentClientSnapshot struct {
	SchemaVersion     string                               `json:"schema_version"`
	SavedAt           time.Time                            `json:"saved_at"`
	WorkspaceID       string                               `json:"workspace_id,omitempty"`
	Devices           []DeviceRecord                       `json:"devices"`
	Daemons           []DeviceDaemonRecord                 `json:"daemons"`
	DeviceCredentials []DeviceCredentialSnapshotRecord     `json:"device_credentials"`
	NextDeviceID      int                                  `json:"next_device_id"`
	NextDaemonCommand int                                  `json:"next_daemon_command"`
	Agents            []AgentClientRecord                  `json:"agents"`
	Fixtures          []AgentOnboardingFixture             `json:"fixtures"`
	TaskThreads       map[string][]AIAgentTaskThreadRecord `json:"task_threads"`
	Events            []AIAgentClientSnapshotEvent         `json:"events"`
}

type DeviceCredentialSnapshotRecord struct {
	DeviceID         string    `json:"device_id"`
	SecretHashHex    string    `json:"secret_hash_hex"`
	OwnerPrincipalID string    `json:"owner_principal_id"`
	WorkspaceID      string    `json:"workspace_id,omitempty"`
	DisplayName      string    `json:"display_name,omitempty"`
	IssuedAt         time.Time `json:"issued_at"`
}

type AIAgentClientSnapshotEvent struct {
	Seq       int64           `json:"seq"`
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
}

func NewDevelopmentAIAgentClientStore(ctx context.Context, config DevelopmentAIAgentClientStoreConfig) (*DevelopmentAIAgentClientStore, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	store := newEmptyAIAgentClientStore(config.Persistence)
	if config.Persistence == nil {
		return store, nil
	}
	snapshot, ok, err := config.Persistence.LoadAIAgentClientSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	if ok {
		if err := store.applySnapshot(snapshot); err != nil {
			return nil, err
		}
	}
	return store, nil
}

func (s *DevelopmentAIAgentClientStore) Close() error {
	if s == nil || s.persistence == nil {
		return nil
	}
	s.mu.Lock()
	snapshot, err := snapshotFromAIAgentClientStoreLocked(s, time.Now().UTC())
	s.mu.Unlock()
	if err != nil {
		return err
	}
	if err := s.persistence.SaveAIAgentClientSnapshot(context.Background(), snapshot); err != nil {
		return err
	}
	return s.persistence.Close()
}

func (s *DevelopmentAIAgentClientStore) saveSnapshotLocked(ctx context.Context) error {
	if s == nil || s.persistence == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	snapshot, err := snapshotFromAIAgentClientStoreLocked(s, time.Now().UTC())
	if err != nil {
		return err
	}
	return s.persistence.SaveAIAgentClientSnapshot(ctx, snapshot)
}

func (s *DevelopmentAIAgentClientStore) applySnapshot(snapshot AIAgentClientSnapshot) error {
	if snapshot.SchemaVersion != AIAgentClientPersistenceSchemaVersion {
		return fmt.Errorf("unsupported ai agent client snapshot schema_version %q", snapshot.SchemaVersion)
	}
	agents := make(map[string]AgentClientRecord, len(snapshot.Agents))
	for _, agent := range snapshot.Agents {
		agent.AgentID = strings.TrimSpace(agent.AgentID)
		if agent.AgentID == "" {
			return errors.New("ai agent client snapshot agent_id is required")
		}
		agents[agent.AgentID] = agent
	}
	daemons := make(map[string]DeviceDaemonRecord, len(snapshot.Daemons))
	for _, daemon := range snapshot.Daemons {
		daemon.DeviceID = strings.TrimSpace(daemon.DeviceID)
		if daemon.DeviceID == "" {
			return errors.New("ai agent client snapshot daemon.device_id is required")
		}
		daemons[daemon.DeviceID] = copyDeviceDaemon(daemon)
	}
	credentials := make(map[string]deviceCredentialRecord, len(snapshot.DeviceCredentials))
	for _, credential := range snapshot.DeviceCredentials {
		record, err := deviceCredentialRecordFromSnapshot(credential)
		if err != nil {
			return err
		}
		credentials[record.deviceID] = record
	}
	events := make([]ClientStreamEvent, 0, len(snapshot.Events))
	for _, event := range snapshot.Events {
		typed, err := clientStreamEventFromSnapshot(event)
		if err != nil {
			return err
		}
		events = append(events, typed)
	}
	fixtures := copyAgentOnboardingFixtures(snapshot.Fixtures)
	if len(fixtures) == 0 {
		fixtures = defaultAgentOnboardingFixtures()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.workspaceID = strings.TrimSpace(snapshot.WorkspaceID)
	if s.workspaceID == "" {
		s.workspaceID = defaultAIAgentClientWorkspaceID
	}
	s.devices = copyDevices(snapshot.Devices)
	s.daemons = daemons
	s.deviceCredentials = credentials
	s.nextDeviceID = snapshot.NextDeviceID
	s.nextDaemonCommand = snapshot.NextDaemonCommand
	if s.nextDaemonCommand <= 0 {
		s.nextDaemonCommand = 1
	}
	s.agents = agents
	s.fixtures = fixtures
	s.taskThreads = copyTaskThreads(snapshot.TaskThreads)
	s.events = events
	s.subscribers = map[int]aiAgentClientSubscriber{}
	return nil
}

func snapshotFromAIAgentClientStoreLocked(s *DevelopmentAIAgentClientStore, savedAt time.Time) (AIAgentClientSnapshot, error) {
	agents := make([]AgentClientRecord, 0, len(s.agents))
	for _, agent := range s.agents {
		agents = append(agents, agent)
	}
	sort.Slice(agents, func(i, j int) bool { return agents[i].AgentID < agents[j].AgentID })

	daemons := make([]DeviceDaemonRecord, 0, len(s.daemons))
	for _, daemon := range s.daemons {
		daemons = append(daemons, copyDeviceDaemon(daemon))
	}
	sort.Slice(daemons, func(i, j int) bool { return daemons[i].DeviceID < daemons[j].DeviceID })

	credentials := make([]DeviceCredentialSnapshotRecord, 0, len(s.deviceCredentials))
	for _, credential := range s.deviceCredentials {
		credentials = append(credentials, snapshotFromDeviceCredentialRecord(credential))
	}
	sort.Slice(credentials, func(i, j int) bool { return credentials[i].DeviceID < credentials[j].DeviceID })

	events := make([]AIAgentClientSnapshotEvent, 0, len(s.events))
	for _, event := range s.events {
		snapshotEvent, err := snapshotFromClientStreamEvent(event)
		if err != nil {
			return AIAgentClientSnapshot{}, err
		}
		events = append(events, snapshotEvent)
	}

	return AIAgentClientSnapshot{
		SchemaVersion:     AIAgentClientPersistenceSchemaVersion,
		SavedAt:           savedAt.UTC(),
		WorkspaceID:       s.workspaceID,
		Devices:           copyDevices(s.devices),
		Daemons:           daemons,
		DeviceCredentials: credentials,
		NextDeviceID:      s.nextDeviceID,
		NextDaemonCommand: s.nextDaemonCommand,
		Agents:            agents,
		Fixtures:          copyAgentOnboardingFixtures(s.fixtures),
		TaskThreads:       copyTaskThreads(s.taskThreads),
		Events:            events,
	}, nil
}

func snapshotFromDeviceCredentialRecord(record deviceCredentialRecord) DeviceCredentialSnapshotRecord {
	return DeviceCredentialSnapshotRecord{
		DeviceID:         record.deviceID,
		SecretHashHex:    hex.EncodeToString(record.secretHash[:]),
		OwnerPrincipalID: record.ownerPrincipalID,
		WorkspaceID:      record.workspaceID,
		DisplayName:      record.displayName,
		IssuedAt:         record.issuedAt,
	}
}

func deviceCredentialRecordFromSnapshot(snapshot DeviceCredentialSnapshotRecord) (deviceCredentialRecord, error) {
	deviceID := strings.TrimSpace(snapshot.DeviceID)
	if deviceID == "" {
		return deviceCredentialRecord{}, errors.New("ai agent client snapshot device credential device_id is required")
	}
	rawHash, err := hex.DecodeString(strings.TrimSpace(snapshot.SecretHashHex))
	if err != nil {
		return deviceCredentialRecord{}, fmt.Errorf("decode device credential hash: %w", err)
	}
	if len(rawHash) != sha256.Size {
		return deviceCredentialRecord{}, errors.New("device credential hash length is invalid")
	}
	var hash [sha256.Size]byte
	copy(hash[:], rawHash)
	return deviceCredentialRecord{
		deviceID:         deviceID,
		secretHash:       hash,
		ownerPrincipalID: strings.TrimSpace(snapshot.OwnerPrincipalID),
		workspaceID:      strings.TrimSpace(snapshot.WorkspaceID),
		displayName:      strings.TrimSpace(snapshot.DisplayName),
		issuedAt:         snapshot.IssuedAt,
	}, nil
}

func snapshotFromClientStreamEvent(event ClientStreamEvent) (AIAgentClientSnapshotEvent, error) {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return AIAgentClientSnapshotEvent{}, err
	}
	return AIAgentClientSnapshotEvent{Seq: event.Seq, EventType: event.EventType, Payload: payload}, nil
}

func clientStreamEventFromSnapshot(event AIAgentClientSnapshotEvent) (ClientStreamEvent, error) {
	var payload any
	switch event.EventType {
	case AgentClientEventDeviceRuntimeSnapshot:
		var typed DeviceRuntimeSnapshotEvent
		if err := json.Unmarshal(event.Payload, &typed); err != nil {
			return ClientStreamEvent{}, err
		}
		payload = typed
	case AgentClientEventDeviceDaemonStatus:
		var typed DeviceDaemonStatusEvent
		if err := json.Unmarshal(event.Payload, &typed); err != nil {
			return ClientStreamEvent{}, err
		}
		payload = typed
	case AgentClientEventEditabilityChanged:
		var typed AgentEditabilityChangedEvent
		if err := json.Unmarshal(event.Payload, &typed); err != nil {
			return ClientStreamEvent{}, err
		}
		payload = typed
	case AgentClientEventWorkStatusChanged:
		var typed AgentWorkStatusChangedEvent
		if err := json.Unmarshal(event.Payload, &typed); err != nil {
			return ClientStreamEvent{}, err
		}
		payload = typed
	case AgentClientEventThreadProgress:
		var typed AgentThreadProgressEvent
		if err := json.Unmarshal(event.Payload, &typed); err != nil {
			return ClientStreamEvent{}, err
		}
		payload = typed
	default:
		return ClientStreamEvent{}, fmt.Errorf("unsupported ai agent client event_type %q", event.EventType)
	}
	return ClientStreamEvent{Seq: event.Seq, EventType: event.EventType, Payload: payload}, nil
}

func copyTaskThreads(in map[string][]AIAgentTaskThreadRecord) map[string][]AIAgentTaskThreadRecord {
	out := make(map[string][]AIAgentTaskThreadRecord, len(in))
	for taskID, threads := range in {
		out[taskID] = copyTaskThreadRecords(threads)
	}
	return out
}

func copyTaskThreadRecords(in []AIAgentTaskThreadRecord) []AIAgentTaskThreadRecord {
	out := make([]AIAgentTaskThreadRecord, 0, len(in))
	for _, thread := range in {
		out = append(out, copyTaskThread(thread))
	}
	return out
}

type DynamoDBAIAgentClientSnapshotConfig struct {
	Region              string
	TableName           string
	Endpoint            string
	HTTPClient          *http.Client
	Now                 func() time.Time
	CredentialsProvider AWSCredentialsProvider
}

type DynamoDBAIAgentClientSnapshot struct {
	region              string
	tableName           string
	endpoint            string
	endpointHost        string
	httpClient          *http.Client
	now                 func() time.Time
	credentialsProvider AWSCredentialsProvider
}

func NewDynamoDBAIAgentClientSnapshot(config DynamoDBAIAgentClientSnapshotConfig) (*DynamoDBAIAgentClientSnapshot, error) {
	region := strings.TrimSpace(config.Region)
	if region == "" {
		return nil, errors.New("riidoaiserver: DynamoDB AI Agent client snapshot region is required")
	}
	tableName := strings.TrimSpace(config.TableName)
	if tableName == "" {
		return nil, errors.New("riidoaiserver: DynamoDB AI Agent client snapshot table name is required")
	}
	if config.CredentialsProvider == nil {
		return nil, errors.New("riidoaiserver: DynamoDB AI Agent client snapshot credentials provider is required")
	}
	endpoint, endpointHost, err := normalizeDynamoDBEndpoint(region, strings.TrimSpace(config.Endpoint))
	if err != nil {
		return nil, err
	}
	return &DynamoDBAIAgentClientSnapshot{
		region:              region,
		tableName:           tableName,
		endpoint:            endpoint,
		endpointHost:        endpointHost,
		httpClient:          dynamoDBHTTPClient(config.HTTPClient),
		now:                 dynamoDBClock(config.Now),
		credentialsProvider: config.CredentialsProvider,
	}, nil
}

func (s *DynamoDBAIAgentClientSnapshot) LoadAIAgentClientSnapshot(ctx context.Context) (AIAgentClientSnapshot, bool, error) {
	if s == nil {
		return AIAgentClientSnapshot{}, false, nil
	}
	credentials, err := cachedAWSCredentials(ctx, s.now, s.credentialsProvider, nil)
	if err != nil {
		return AIAgentClientSnapshot{}, false, err
	}
	payload, err := json.Marshal(struct {
		TableName      string                       `json:"TableName"`
		ConsistentRead bool                         `json:"ConsistentRead"`
		Key            map[string]map[string]string `json:"Key"`
	}{
		TableName:      s.tableName,
		ConsistentRead: true,
		Key: map[string]map[string]string{
			"pk": {"S": aiAgentClientSnapshotPK},
			"sk": {"S": aiAgentClientSnapshotSK},
		},
	})
	if err != nil {
		return AIAgentClientSnapshot{}, false, err
	}
	body, err := doDynamoDBJSON(ctx, dynamoDBRequest{
		endpoint:     s.endpoint,
		endpointHost: s.endpointHost,
		region:       s.region,
		target:       dynamoDBGetItemTarget,
		payload:      payload,
		credentials:  credentials,
		httpClient:   s.httpClient,
		now:          s.now,
	})
	if err != nil {
		return AIAgentClientSnapshot{}, false, fmt.Errorf("dynamodb load ai agent client snapshot: %w", err)
	}
	var response struct {
		Item map[string]map[string]string `json:"Item"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return AIAgentClientSnapshot{}, false, fmt.Errorf("decode DynamoDB AI Agent client snapshot response: %w", err)
	}
	if len(response.Item) == 0 {
		return AIAgentClientSnapshot{}, false, nil
	}
	rawSnapshot := response.Item["snapshot_json"]["S"]
	if rawSnapshot == "" {
		return AIAgentClientSnapshot{}, false, errors.New("decode DynamoDB AI Agent client snapshot response: snapshot_json is required")
	}
	var snapshot AIAgentClientSnapshot
	dec := json.NewDecoder(strings.NewReader(rawSnapshot))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&snapshot); err != nil {
		return AIAgentClientSnapshot{}, false, fmt.Errorf("decode AI Agent client snapshot json: %w", err)
	}
	var extra struct{}
	if err := dec.Decode(&extra); err != io.EOF {
		return AIAgentClientSnapshot{}, false, errors.New("decode AI Agent client snapshot json: trailing data")
	}
	return snapshot, true, nil
}

func (s *DynamoDBAIAgentClientSnapshot) SaveAIAgentClientSnapshot(ctx context.Context, snapshot AIAgentClientSnapshot) error {
	if s == nil {
		return nil
	}
	if snapshot.SchemaVersion != AIAgentClientPersistenceSchemaVersion {
		return fmt.Errorf("unsupported ai agent client snapshot schema_version %q", snapshot.SchemaVersion)
	}
	if snapshot.SavedAt.IsZero() {
		snapshot.SavedAt = s.now()
	}
	credentials, err := cachedAWSCredentials(ctx, s.now, s.credentialsProvider, nil)
	if err != nil {
		return err
	}
	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(struct {
		TableName string                       `json:"TableName"`
		Item      map[string]map[string]string `json:"Item"`
	}{
		TableName: s.tableName,
		Item: map[string]map[string]string{
			"pk":             {"S": aiAgentClientSnapshotPK},
			"sk":             {"S": aiAgentClientSnapshotSK},
			"schema_version": {"S": AIAgentClientPersistenceSchemaVersion},
			"snapshot_json":  {"S": string(snapshotJSON)},
			"saved_at":       {"S": snapshot.SavedAt.UTC().Format(time.RFC3339Nano)},
		},
	})
	if err != nil {
		return err
	}
	_, err = doDynamoDBJSON(ctx, dynamoDBRequest{
		endpoint:     s.endpoint,
		endpointHost: s.endpointHost,
		region:       s.region,
		target:       dynamoDBPutItemTarget,
		payload:      payload,
		credentials:  credentials,
		httpClient:   s.httpClient,
		now:          s.now,
	})
	if err != nil {
		return fmt.Errorf("dynamodb save ai agent client snapshot: %w", err)
	}
	return nil
}

func (s *DynamoDBAIAgentClientSnapshot) Close() error {
	return nil
}
