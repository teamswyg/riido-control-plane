package riidoaiserver

import (
	"sync"
	"time"
)

type AIAgentClientPersistenceOperation int

const (
	AIAgentClientPersistenceOperationUnknown AIAgentClientPersistenceOperation = iota
	AIAgentClientSnapshotLoad
	AIAgentClientSnapshotSave
)

type AIAgentClientPersistenceObservation struct {
	Operation AIAgentClientPersistenceOperation
	Duration  time.Duration
	Bytes     int64
	Err       error
}

type AIAgentClientPersistenceMetrics struct {
	mu   sync.Mutex
	load aiAgentClientPersistenceOperationMetrics
	save aiAgentClientPersistenceOperationMetrics
}

type aiAgentClientPersistenceOperationMetrics struct {
	callsTotal               int64
	errorsTotal              int64
	bytesTotal               int64
	bytesMax                 int64
	bytesLast                int64
	latencySamplesTotal      int64
	latencyTotalMilliseconds int64
	latencyMaxMilliseconds   int64
	latencyLastMilliseconds  int64
}

func NewAIAgentClientPersistenceMetrics() *AIAgentClientPersistenceMetrics {
	return &AIAgentClientPersistenceMetrics{}
}

func (op AIAgentClientPersistenceOperation) String() string {
	switch op {
	case AIAgentClientSnapshotLoad:
		return "ai_agent_client_snapshot_load"
	case AIAgentClientSnapshotSave:
		return "ai_agent_client_snapshot_save"
	default:
		return "unknown"
	}
}

func (m *AIAgentClientPersistenceMetrics) ObserveAIAgentClientPersistence(obs AIAgentClientPersistenceObservation) {
	if m == nil {
		return
	}
	elapsedMS := durationMilliseconds(obs.Duration)
	bytes := obs.Bytes
	if bytes < 0 {
		bytes = 0
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	switch obs.Operation {
	case AIAgentClientPersistenceOperationUnknown:
	case AIAgentClientSnapshotLoad:
		m.load.observe(elapsedMS, bytes, obs.Err)
	case AIAgentClientSnapshotSave:
		m.save.observe(elapsedMS, bytes, obs.Err)
	}
}

func (m *AIAgentClientPersistenceMetrics) ApplyToMetricsSnapshot(snapshot MetricsSnapshot) MetricsSnapshot {
	if m == nil {
		return snapshot
	}
	load, save := m.snapshot()
	snapshot.AIAgentClientSnapshotLoadCallsTotal = load.callsTotal
	snapshot.AIAgentClientSnapshotLoadErrorsTotal = load.errorsTotal
	snapshot.AIAgentClientSnapshotLoadBytesTotal = load.bytesTotal
	snapshot.AIAgentClientSnapshotLoadBytesMax = load.bytesMax
	snapshot.AIAgentClientSnapshotLoadBytesLast = load.bytesLast
	snapshot.AIAgentClientSnapshotLoadLatencySamplesTotal = load.latencySamplesTotal
	snapshot.AIAgentClientSnapshotLoadLatencyTotalMilliseconds = load.latencyTotalMilliseconds
	snapshot.AIAgentClientSnapshotLoadLatencyMaxMilliseconds = load.latencyMaxMilliseconds
	snapshot.AIAgentClientSnapshotLoadLatencyLastMilliseconds = load.latencyLastMilliseconds
	snapshot.AIAgentClientSnapshotSaveCallsTotal = save.callsTotal
	snapshot.AIAgentClientSnapshotSaveErrorsTotal = save.errorsTotal
	snapshot.AIAgentClientSnapshotSaveBytesTotal = save.bytesTotal
	snapshot.AIAgentClientSnapshotSaveBytesMax = save.bytesMax
	snapshot.AIAgentClientSnapshotSaveBytesLast = save.bytesLast
	snapshot.AIAgentClientSnapshotSaveLatencySamplesTotal = save.latencySamplesTotal
	snapshot.AIAgentClientSnapshotSaveLatencyTotalMilliseconds = save.latencyTotalMilliseconds
	snapshot.AIAgentClientSnapshotSaveLatencyMaxMilliseconds = save.latencyMaxMilliseconds
	snapshot.AIAgentClientSnapshotSaveLatencyLastMilliseconds = save.latencyLastMilliseconds
	return snapshot
}

func (m *AIAgentClientPersistenceMetrics) snapshot() (aiAgentClientPersistenceOperationMetrics, aiAgentClientPersistenceOperationMetrics) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.load, m.save
}

func (m *aiAgentClientPersistenceOperationMetrics) observe(elapsedMS, bytes int64, err error) {
	m.callsTotal++
	if err != nil {
		m.errorsTotal++
	}
	m.bytesTotal += bytes
	if bytes > m.bytesMax {
		m.bytesMax = bytes
	}
	m.bytesLast = bytes
	m.latencySamplesTotal++
	m.latencyTotalMilliseconds += elapsedMS
	if elapsedMS > m.latencyMaxMilliseconds {
		m.latencyMaxMilliseconds = elapsedMS
	}
	m.latencyLastMilliseconds = elapsedMS
}
