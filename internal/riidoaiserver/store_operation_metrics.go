package riidoaiserver

import (
	"sort"
	"strings"
	"sync"
	"time"
)

const unknownStoreOperation = "store_unknown"

type StoreOperationName string

const (
	StoreOperationCreateTask       StoreOperationName = "store_create_task"
	StoreOperationCreateAssignment StoreOperationName = "store_create_assignment"
	StoreOperationCancelAssignment StoreOperationName = "store_cancel_assignment"
	StoreOperationPollAssignment   StoreOperationName = "store_poll_assignment"
	StoreOperationWaitAssignment   StoreOperationName = "store_wait_assignment"
	StoreOperationLeaseAssignment  StoreOperationName = "store_lease_assignment"
	StoreOperationAppendEvent      StoreOperationName = "store_append_event"
)

func (op StoreOperationName) String() string {
	value := strings.TrimSpace(string(op))
	if value == "" {
		return unknownStoreOperation
	}
	return value
}

type StoreOperationMetric struct {
	Operation                string `json:"operation"`
	CallsTotal               int64  `json:"calls_total"`
	ErrorsTotal              int64  `json:"errors_total"`
	LatencySamplesTotal      int64  `json:"latency_samples_total"`
	LatencyTotalMilliseconds int64  `json:"latency_total_ms"`
	LatencyMaxMilliseconds   int64  `json:"latency_max_ms"`
	LatencyLastMilliseconds  int64  `json:"latency_last_ms"`
}

type StoreOperationObservation struct {
	Operation StoreOperationName
	Duration  time.Duration
	Err       error
}

type StoreOperationMetrics struct {
	mu          sync.Mutex
	byOperation map[string]StoreOperationMetric
	callsTotal  int64
	errorsTotal int64
	latency     storeOperationLatencyMetrics
}

type storeOperationLatencyMetrics struct {
	samplesTotal      int64
	totalMilliseconds int64
	maxMilliseconds   int64
	lastMilliseconds  int64
}

func NewStoreOperationMetrics() *StoreOperationMetrics {
	return &StoreOperationMetrics{
		byOperation: map[string]StoreOperationMetric{},
	}
}

func (m *StoreOperationMetrics) ObserveStoreOperation(obs StoreOperationObservation) {
	if m == nil {
		return
	}
	operation := obs.Operation.String()
	elapsedMS := durationMilliseconds(obs.Duration)

	m.mu.Lock()
	defer m.mu.Unlock()

	m.callsTotal++
	if obs.Err != nil {
		m.errorsTotal++
	}
	m.latency.observe(elapsedMS)

	metric := m.byOperation[operation]
	metric.Operation = operation
	metric.CallsTotal++
	if obs.Err != nil {
		metric.ErrorsTotal++
	}
	metric.LatencySamplesTotal++
	metric.LatencyTotalMilliseconds += elapsedMS
	if elapsedMS > metric.LatencyMaxMilliseconds {
		metric.LatencyMaxMilliseconds = elapsedMS
	}
	metric.LatencyLastMilliseconds = elapsedMS
	m.byOperation[operation] = metric
}

func (m *StoreOperationMetrics) ApplyToMetricsSnapshot(snapshot MetricsSnapshot) MetricsSnapshot {
	if m == nil {
		return snapshot
	}
	callsTotal, errorsTotal, latency, operations := m.snapshot()
	snapshot.StoreOperationCallsTotal = callsTotal
	snapshot.StoreOperationErrorsTotal = errorsTotal
	snapshot.StoreOperationLatencySamplesTotal = latency.samplesTotal
	snapshot.StoreOperationLatencyTotalMilliseconds = latency.totalMilliseconds
	snapshot.StoreOperationLatencyMaxMilliseconds = latency.maxMilliseconds
	snapshot.StoreOperationLatencyLastMilliseconds = latency.lastMilliseconds
	snapshot.StoreOperations = operations
	return snapshot
}

func (m *StoreOperationMetrics) snapshot() (int64, int64, storeOperationLatencyMetrics, []StoreOperationMetric) {
	m.mu.Lock()
	defer m.mu.Unlock()

	operations := make([]StoreOperationMetric, 0, len(m.byOperation))
	for _, metric := range m.byOperation {
		operations = append(operations, metric)
	}
	sort.Slice(operations, func(i, j int) bool {
		return operations[i].Operation < operations[j].Operation
	})
	return m.callsTotal, m.errorsTotal, m.latency, operations
}

func (m *storeOperationLatencyMetrics) observe(elapsedMS int64) {
	m.samplesTotal++
	m.totalMilliseconds += elapsedMS
	if elapsedMS > m.maxMilliseconds {
		m.maxMilliseconds = elapsedMS
	}
	m.lastMilliseconds = elapsedMS
}
