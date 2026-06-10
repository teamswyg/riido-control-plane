package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Split-item AI-agent-client persistence (rev3 plan): the monolithic snapshot is
// stored as three independent DynamoDB items so the daemon hot path
// (runtime-snapshot) reads/writes only the small "core" item — never the fat
// event log or the task threads. Events live in their own item (event append
// rewrites only events, not threads). Threads are one item (complete in-memory
// invariant preserved; per-task sharding is a follow-up). Cross-item and
// migration writes are always <= 3 items, so a single TransactWriteItems makes
// them atomic — no 100-item limit, no migration marker needed.
const (
	dynamoDBAIAgentClientCorePK    = "AI_AGENT_CLIENT#core"
	dynamoDBAIAgentClientEventsPK  = "AI_AGENT_CLIENT#events"
	dynamoDBAIAgentClientThreadsPK = "AI_AGENT_CLIENT#threads"
	dynamoDBAIAgentClientSplitSK   = "CURRENT"
)

// AIAgentClientCoreSnapshot is everything except events and task threads — the
// small, broadly-needed state the daemon hot path touches.
type AIAgentClientCoreSnapshot struct {
	SchemaVersion           string                                  `json:"schema_version"`
	SavedAt                 time.Time                               `json:"saved_at"`
	WorkspaceID             string                                  `json:"workspace_id"`
	Devices                 []DeviceRecord                          `json:"devices"`
	DeviceCredentials       []AIAgentClientDeviceCredentialSnapshot `json:"device_credentials"`
	Daemons                 []DeviceDaemonRecord                    `json:"daemons"`
	Agents                  []AgentClientRecord                     `json:"agents"`
	Fixtures                []AgentOnboardingFixture                `json:"fixtures"`
	NextDeviceCredentialSeq int                                     `json:"next_device_credential_seq"`
	NextDaemonCommand       int                                     `json:"next_daemon_command"`
}

// AIAgentClientEventsSnapshot is the replay log + its seq counter (decoupled
// from core so event append never writes core).
type AIAgentClientEventsSnapshot struct {
	SchemaVersion      string                       `json:"schema_version"`
	SavedAt            time.Time                    `json:"saved_at"`
	Events             []AIAgentClientEventSnapshot `json:"events"`
	NextClientEventSeq int64                        `json:"next_client_event_seq"`
}

// AIAgentClientThreadsSnapshot is all task threads as one item.
type AIAgentClientThreadsSnapshot struct {
	SchemaVersion      string                                      `json:"schema_version"`
	SavedAt            time.Time                                   `json:"saved_at"`
	TaskThreads        map[string][]AIAgentTaskThreadRecord        `json:"task_threads"`
	TaskThreadMessages map[string][]AIAgentTaskThreadMessageRecord `json:"task_thread_messages,omitempty"`
}

// AIAgentClientSplitSnapshotStore is the optional per-collection interface. A
// snapshot store that implements it (DynamoDB) lets the persistence wrapper read
// /write each collection independently; stores that don't keep using the legacy
// combined Load/Save.
type AIAgentClientSplitSnapshotStore interface {
	LoadCore(ctx context.Context) (AIAgentClientCoreSnapshot, bool, error)
	SaveCore(ctx context.Context, core AIAgentClientCoreSnapshot) error
	LoadEvents(ctx context.Context) (AIAgentClientEventsSnapshot, bool, error)
	SaveEvents(ctx context.Context, events AIAgentClientEventsSnapshot) error
	LoadThreads(ctx context.Context) (AIAgentClientThreadsSnapshot, bool, error)
	SaveThreads(ctx context.Context, threads AIAgentClientThreadsSnapshot) error
	// WriteSplitAtomic writes core+events+threads in one transaction (migration
	// and any cross-collection mutation). Atomic: no half-written split.
	WriteSplitAtomic(ctx context.Context, core AIAgentClientCoreSnapshot, events AIAgentClientEventsSnapshot, threads AIAgentClientThreadsSnapshot) error
	// WriteCoreEvents writes core+events atomically (the Sync-on-change path:
	// device update + appended event, without rewriting the threads item).
	WriteCoreEvents(ctx context.Context, core AIAgentClientCoreSnapshot, events AIAgentClientEventsSnapshot) error
}

// splitFromCombined projects a combined snapshot into the three sub-snapshots.
func splitFromCombined(s AIAgentClientSnapshot) (AIAgentClientCoreSnapshot, AIAgentClientEventsSnapshot, AIAgentClientThreadsSnapshot) {
	core := AIAgentClientCoreSnapshot{
		SchemaVersion:           AIAgentClientPersistenceSchemaVersion,
		SavedAt:                 s.SavedAt,
		WorkspaceID:             s.WorkspaceID,
		Devices:                 s.Devices,
		DeviceCredentials:       s.DeviceCredentials,
		Daemons:                 s.Daemons,
		Agents:                  s.Agents,
		Fixtures:                s.Fixtures,
		NextDeviceCredentialSeq: s.NextDeviceCredentialSeq,
		NextDaemonCommand:       s.NextDaemonCommand,
	}
	events := AIAgentClientEventsSnapshot{
		SchemaVersion:      AIAgentClientPersistenceSchemaVersion,
		SavedAt:            s.SavedAt,
		Events:             s.Events,
		NextClientEventSeq: s.NextClientEventSeq,
	}
	threads := AIAgentClientThreadsSnapshot{
		SchemaVersion:      AIAgentClientPersistenceSchemaVersion,
		SavedAt:            s.SavedAt,
		TaskThreads:        s.TaskThreads,
		TaskThreadMessages: s.TaskThreadMessages,
	}
	return core, events, threads
}

// combinedFromSplit reassembles a combined snapshot (used by restore paths that
// still expect the full struct, and by tests).
func combinedFromSplit(core AIAgentClientCoreSnapshot, events AIAgentClientEventsSnapshot, threads AIAgentClientThreadsSnapshot) AIAgentClientSnapshot {
	return AIAgentClientSnapshot{
		SchemaVersion:           AIAgentClientPersistenceSchemaVersion,
		SavedAt:                 core.SavedAt,
		WorkspaceID:             core.WorkspaceID,
		Devices:                 core.Devices,
		DeviceCredentials:       core.DeviceCredentials,
		Daemons:                 core.Daemons,
		Agents:                  core.Agents,
		Fixtures:                core.Fixtures,
		TaskThreads:             threads.TaskThreads,
		TaskThreadMessages:      threads.TaskThreadMessages,
		Events:                  events.Events,
		NextDeviceCredentialSeq: core.NextDeviceCredentialSeq,
		NextDaemonCommand:       core.NextDaemonCommand,
		NextClientEventSeq:      events.NextClientEventSeq,
	}
}

// --- low-level item I/O (routed through the serial loop via do()) ---

func (s *DynamoDBAIAgentClientSnapshot) do(ctx context.Context, target string, payload []byte) ([]byte, error) {
	if s == nil {
		return nil, errors.New("riidoaiserver: nil DynamoDB AI Agent client snapshot store")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	reply := make(chan dynamoDBRawResult, 1)
	select {
	case s.commands <- dynamoDBAIAgentClientSnapshotCommand{ctx: ctx, raw: &dynamoDBRawOp{target: target, payload: payload}, rawDone: reply}:
	case <-s.done:
		return nil, errors.New("riidoaiserver: DynamoDB AI Agent client snapshot store closed")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	select {
	case r := <-reply:
		return r.body, r.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *DynamoDBAIAgentClientSnapshot) getGzipItem(ctx context.Context, pk string) ([]byte, bool, error) {
	payload, err := json.Marshal(struct {
		TableName      string                       `json:"TableName"`
		ConsistentRead bool                         `json:"ConsistentRead"`
		Key            map[string]map[string]string `json:"Key"`
	}{
		TableName:      s.tableName,
		ConsistentRead: true,
		Key:            map[string]map[string]string{"pk": {"S": pk}, "sk": {"S": dynamoDBAIAgentClientSplitSK}},
	})
	if err != nil {
		return nil, false, err
	}
	body, err := s.do(ctx, dynamoDBGetItemTarget, payload)
	if err != nil {
		return nil, false, fmt.Errorf("dynamodb get %s: %w", pk, err)
	}
	var resp struct {
		Item map[string]map[string]string `json:"Item"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, false, fmt.Errorf("decode %s response: %w", pk, err)
	}
	if len(resp.Item) == 0 {
		return nil, false, nil
	}
	gz := resp.Item["snapshot_gzip"]["S"]
	if gz == "" {
		return nil, false, fmt.Errorf("decode %s: snapshot_gzip is required", pk)
	}
	raw, err := gunzipBase64(gz)
	if err != nil {
		return nil, false, fmt.Errorf("decode %s gzip: %w", pk, err)
	}
	return raw, true, nil
}

func (s *DynamoDBAIAgentClientSnapshot) gzipItemAttrs(pk string, jsonBytes []byte) (map[string]map[string]string, error) {
	gz, err := gzipBase64(jsonBytes)
	if err != nil {
		return nil, err
	}
	return map[string]map[string]string{
		"pk":             {"S": pk},
		"sk":             {"S": dynamoDBAIAgentClientSplitSK},
		"schema_version": {"S": AIAgentClientPersistenceSchemaVersion},
		"snapshot_gzip":  {"S": gz},
		"saved_at":       {"S": s.now().UTC().Format(time.RFC3339Nano)},
	}, nil
}

func (s *DynamoDBAIAgentClientSnapshot) putGzipItem(ctx context.Context, pk string, jsonBytes []byte) error {
	attrs, err := s.gzipItemAttrs(pk, jsonBytes)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(struct {
		TableName string                       `json:"TableName"`
		Item      map[string]map[string]string `json:"Item"`
	}{TableName: s.tableName, Item: attrs})
	if err != nil {
		return err
	}
	if _, err := s.do(ctx, dynamoDBPutItemTarget, payload); err != nil {
		return fmt.Errorf("dynamodb put %s: %w", pk, err)
	}
	return nil
}

// --- per-collection methods (AIAgentClientSplitSnapshotStore) ---

func (s *DynamoDBAIAgentClientSnapshot) LoadCore(ctx context.Context) (AIAgentClientCoreSnapshot, bool, error) {
	raw, ok, err := s.getGzipItem(ctx, dynamoDBAIAgentClientCorePK)
	if err != nil || !ok {
		return AIAgentClientCoreSnapshot{}, ok, err
	}
	var core AIAgentClientCoreSnapshot
	if err := json.Unmarshal(raw, &core); err != nil {
		return AIAgentClientCoreSnapshot{}, false, fmt.Errorf("decode core snapshot json: %w", err)
	}
	return core, true, nil
}

func (s *DynamoDBAIAgentClientSnapshot) SaveCore(ctx context.Context, core AIAgentClientCoreSnapshot) error {
	core.SchemaVersion = AIAgentClientPersistenceSchemaVersion
	b, err := json.Marshal(core)
	if err != nil {
		return err
	}
	return s.putGzipItem(ctx, dynamoDBAIAgentClientCorePK, b)
}

func (s *DynamoDBAIAgentClientSnapshot) LoadEvents(ctx context.Context) (AIAgentClientEventsSnapshot, bool, error) {
	raw, ok, err := s.getGzipItem(ctx, dynamoDBAIAgentClientEventsPK)
	if err != nil || !ok {
		return AIAgentClientEventsSnapshot{}, ok, err
	}
	var ev AIAgentClientEventsSnapshot
	if err := json.Unmarshal(raw, &ev); err != nil {
		return AIAgentClientEventsSnapshot{}, false, fmt.Errorf("decode events snapshot json: %w", err)
	}
	return ev, true, nil
}

func (s *DynamoDBAIAgentClientSnapshot) SaveEvents(ctx context.Context, ev AIAgentClientEventsSnapshot) error {
	ev.SchemaVersion = AIAgentClientPersistenceSchemaVersion
	b, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	return s.putGzipItem(ctx, dynamoDBAIAgentClientEventsPK, b)
}

func (s *DynamoDBAIAgentClientSnapshot) LoadThreads(ctx context.Context) (AIAgentClientThreadsSnapshot, bool, error) {
	raw, ok, err := s.getGzipItem(ctx, dynamoDBAIAgentClientThreadsPK)
	if err != nil || !ok {
		return AIAgentClientThreadsSnapshot{}, ok, err
	}
	var th AIAgentClientThreadsSnapshot
	if err := json.Unmarshal(raw, &th); err != nil {
		return AIAgentClientThreadsSnapshot{}, false, fmt.Errorf("decode threads snapshot json: %w", err)
	}
	return th, true, nil
}

func (s *DynamoDBAIAgentClientSnapshot) SaveThreads(ctx context.Context, th AIAgentClientThreadsSnapshot) error {
	th.SchemaVersion = AIAgentClientPersistenceSchemaVersion
	b, err := json.Marshal(th)
	if err != nil {
		return err
	}
	return s.putGzipItem(ctx, dynamoDBAIAgentClientThreadsPK, b)
}

type splitGzipEntry struct {
	pk        string
	jsonBytes []byte
}

func (s *DynamoDBAIAgentClientSnapshot) transactPutGzip(ctx context.Context, entries []splitGzipEntry) error {
	type txPut struct {
		Put struct {
			TableName string                       `json:"TableName"`
			Item      map[string]map[string]string `json:"Item"`
		} `json:"Put"`
	}
	items := make([]txPut, len(entries))
	for i, e := range entries {
		attrs, err := s.gzipItemAttrs(e.pk, e.jsonBytes)
		if err != nil {
			return err
		}
		items[i].Put.TableName = s.tableName
		items[i].Put.Item = attrs
	}
	payload, err := json.Marshal(struct {
		TransactItems []txPut `json:"TransactItems"`
	}{TransactItems: items})
	if err != nil {
		return err
	}
	if _, err := s.do(ctx, dynamoDBTransactWriteTarget, payload); err != nil {
		return fmt.Errorf("dynamodb transact-write (%d items): %w", len(entries), err)
	}
	return nil
}

func (s *DynamoDBAIAgentClientSnapshot) WriteSplitAtomic(ctx context.Context, core AIAgentClientCoreSnapshot, ev AIAgentClientEventsSnapshot, th AIAgentClientThreadsSnapshot) error {
	core.SchemaVersion = AIAgentClientPersistenceSchemaVersion
	ev.SchemaVersion = AIAgentClientPersistenceSchemaVersion
	th.SchemaVersion = AIAgentClientPersistenceSchemaVersion
	coreJSON, err := json.Marshal(core)
	if err != nil {
		return err
	}
	evJSON, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	thJSON, err := json.Marshal(th)
	if err != nil {
		return err
	}
	return s.transactPutGzip(ctx, []splitGzipEntry{
		{pk: dynamoDBAIAgentClientCorePK, jsonBytes: coreJSON},
		{pk: dynamoDBAIAgentClientEventsPK, jsonBytes: evJSON},
		{pk: dynamoDBAIAgentClientThreadsPK, jsonBytes: thJSON},
	})
}

func (s *DynamoDBAIAgentClientSnapshot) WriteCoreEvents(ctx context.Context, core AIAgentClientCoreSnapshot, ev AIAgentClientEventsSnapshot) error {
	core.SchemaVersion = AIAgentClientPersistenceSchemaVersion
	ev.SchemaVersion = AIAgentClientPersistenceSchemaVersion
	coreJSON, err := json.Marshal(core)
	if err != nil {
		return err
	}
	evJSON, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	return s.transactPutGzip(ctx, []splitGzipEntry{
		{pk: dynamoDBAIAgentClientCorePK, jsonBytes: coreJSON},
		{pk: dynamoDBAIAgentClientEventsPK, jsonBytes: evJSON},
	})
}
