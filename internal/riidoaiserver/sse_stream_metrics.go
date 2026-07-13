package riidoaiserver

import (
	"sort"
	"time"
)

type SSEStreamMetric struct {
	Route                            string `json:"route"`
	ClientSurface                    string `json:"client_surface,omitempty"`
	StreamsOpenedTotal               int64  `json:"streams_opened_total"`
	StreamsClosedTotal               int64  `json:"streams_closed_total"`
	ActiveStreams                    int64  `json:"active_streams"`
	TimeToFirstByteSamplesTotal      int64  `json:"ttfb_samples_total"`
	TimeToFirstByteTotalMilliseconds int64  `json:"ttfb_total_ms"`
	TimeToFirstByteMaxMilliseconds   int64  `json:"ttfb_max_ms"`
	TimeToFirstByteLastMilliseconds  int64  `json:"ttfb_last_ms"`
	StreamDurationSamplesTotal       int64  `json:"stream_duration_samples_total"`
	StreamDurationTotalMilliseconds  int64  `json:"stream_duration_total_ms"`
	StreamDurationMaxMilliseconds    int64  `json:"stream_duration_max_ms"`
	StreamDurationLastMilliseconds   int64  `json:"stream_duration_last_ms"`
}

type SSEStreamOpenObservation struct {
	Route           string
	ClientSurface   string
	TimeToFirstByte time.Duration
	ObservedAt      time.Time
}

type SSEStreamCloseObservation struct {
	Route         string
	ClientSurface string
	Duration      time.Duration
	ObservedAt    time.Time
}

type sseStreamKey struct {
	route         string
	clientSurface string
}

type sseStreamMetricState struct {
	metric                 SSEStreamMetric
	lastTTFBObservedAt     time.Time
	lastDurationObservedAt time.Time
}

type sseStreamMetricsSnapshot struct {
	streamsOpenedTotal     int64
	streamsClosedTotal     int64
	activeStreams          int64
	ttfb                   httpTransactionLatencyMetrics
	ttfbLastObservedAt     time.Time
	streamDuration         httpTransactionLatencyMetrics
	durationLastObservedAt time.Time
	streams                []SSEStreamMetric
}

func (m *HTTPTransactionMetrics) ObserveSSEStreamOpen(obs SSEStreamOpenObservation) {
	if m == nil {
		return
	}
	key := normalizedSSEStreamKey(obs.Route, obs.ClientSurface)
	observedAt := metricsObservedAt(obs.ObservedAt)
	elapsedMS := durationMilliseconds(obs.TimeToFirstByte)

	m.mu.Lock()
	defer m.mu.Unlock()
	m.pruneLocked(observedAt)
	bucket := m.bucketLocked(metricsBucketStart(observedAt))
	state := bucket.sseByKey[key]
	state.metric.Route = key.route
	state.metric.ClientSurface = key.clientSurface
	state.metric.StreamsOpenedTotal++
	state.metric.TimeToFirstByteSamplesTotal++
	state.metric.TimeToFirstByteTotalMilliseconds += elapsedMS
	if elapsedMS > state.metric.TimeToFirstByteMaxMilliseconds {
		state.metric.TimeToFirstByteMaxMilliseconds = elapsedMS
	}
	state.metric.TimeToFirstByteLastMilliseconds = elapsedMS
	state.lastTTFBObservedAt = observedAt
	bucket.sseByKey[key] = state
	m.activeSSEStreams[key]++
}

func (m *HTTPTransactionMetrics) ObserveSSEStreamClose(obs SSEStreamCloseObservation) {
	if m == nil {
		return
	}
	key := normalizedSSEStreamKey(obs.Route, obs.ClientSurface)
	observedAt := metricsObservedAt(obs.ObservedAt)
	elapsedMS := durationMilliseconds(obs.Duration)

	m.mu.Lock()
	defer m.mu.Unlock()
	m.pruneLocked(observedAt)
	bucket := m.bucketLocked(metricsBucketStart(observedAt))
	state := bucket.sseByKey[key]
	state.metric.Route = key.route
	state.metric.ClientSurface = key.clientSurface
	state.metric.StreamsClosedTotal++
	state.metric.StreamDurationSamplesTotal++
	state.metric.StreamDurationTotalMilliseconds += elapsedMS
	if elapsedMS > state.metric.StreamDurationMaxMilliseconds {
		state.metric.StreamDurationMaxMilliseconds = elapsedMS
	}
	state.metric.StreamDurationLastMilliseconds = elapsedMS
	state.lastDurationObservedAt = observedAt
	bucket.sseByKey[key] = state
	if m.activeSSEStreams[key] <= 1 {
		delete(m.activeSSEStreams, key)
	} else {
		m.activeSSEStreams[key]--
	}
}

func (m *HTTPTransactionMetrics) sseSnapshot() sseStreamMetricsSnapshot {
	if m == nil {
		return sseStreamMetricsSnapshot{}
	}
	byKey := map[sseStreamKey]sseStreamMetricState{}

	m.mu.Lock()
	m.pruneLocked(time.Now())
	for _, bucket := range m.buckets {
		for key, source := range bucket.sseByKey {
			mergeSSEStreamMetricState(byKey, key, source)
		}
	}
	for key, active := range m.activeSSEStreams {
		state := byKey[key]
		state.metric.Route = key.route
		state.metric.ClientSurface = key.clientSurface
		state.metric.ActiveStreams = active
		byKey[key] = state
	}
	m.mu.Unlock()

	snapshot := sseStreamMetricsSnapshot{streams: make([]SSEStreamMetric, 0, len(byKey))}
	for _, state := range byKey {
		metric := state.metric
		snapshot.streamsOpenedTotal += metric.StreamsOpenedTotal
		snapshot.streamsClosedTotal += metric.StreamsClosedTotal
		snapshot.activeStreams += metric.ActiveStreams
		mergeSSELatency(&snapshot.ttfb,
			&snapshot.ttfbLastObservedAt,
			metric.TimeToFirstByteSamplesTotal,
			metric.TimeToFirstByteTotalMilliseconds,
			metric.TimeToFirstByteMaxMilliseconds,
			metric.TimeToFirstByteLastMilliseconds,
			state.lastTTFBObservedAt,
		)
		mergeSSELatency(&snapshot.streamDuration,
			&snapshot.durationLastObservedAt,
			metric.StreamDurationSamplesTotal,
			metric.StreamDurationTotalMilliseconds,
			metric.StreamDurationMaxMilliseconds,
			metric.StreamDurationLastMilliseconds,
			state.lastDurationObservedAt,
		)
		snapshot.streams = append(snapshot.streams, metric)
	}
	sort.Slice(snapshot.streams, func(i, j int) bool {
		if snapshot.streams[i].Route != snapshot.streams[j].Route {
			return snapshot.streams[i].Route < snapshot.streams[j].Route
		}
		return snapshot.streams[i].ClientSurface < snapshot.streams[j].ClientSurface
	})
	if len(snapshot.streams) > metricsBreakdownLimit {
		snapshot.streams = snapshot.streams[:metricsBreakdownLimit]
	}
	return snapshot
}

func normalizedSSEStreamKey(route, clientSurface string) sseStreamKey {
	return sseStreamKey{
		route:         normalizeHTTPMetricValue(route, unknownHTTPRoute),
		clientSurface: normalizeHTTPMetricValue(clientSurface, "unknown"),
	}
}

func mergeSSEStreamMetricState(byKey map[sseStreamKey]sseStreamMetricState, key sseStreamKey, source sseStreamMetricState) {
	destination := byKey[key]
	destination.metric.Route = key.route
	destination.metric.ClientSurface = key.clientSurface
	destination.metric.StreamsOpenedTotal += source.metric.StreamsOpenedTotal
	destination.metric.StreamsClosedTotal += source.metric.StreamsClosedTotal
	destination.metric.TimeToFirstByteSamplesTotal += source.metric.TimeToFirstByteSamplesTotal
	destination.metric.TimeToFirstByteTotalMilliseconds += source.metric.TimeToFirstByteTotalMilliseconds
	if source.metric.TimeToFirstByteMaxMilliseconds > destination.metric.TimeToFirstByteMaxMilliseconds {
		destination.metric.TimeToFirstByteMaxMilliseconds = source.metric.TimeToFirstByteMaxMilliseconds
	}
	if source.lastTTFBObservedAt.After(destination.lastTTFBObservedAt) {
		destination.metric.TimeToFirstByteLastMilliseconds = source.metric.TimeToFirstByteLastMilliseconds
		destination.lastTTFBObservedAt = source.lastTTFBObservedAt
	}
	destination.metric.StreamDurationSamplesTotal += source.metric.StreamDurationSamplesTotal
	destination.metric.StreamDurationTotalMilliseconds += source.metric.StreamDurationTotalMilliseconds
	if source.metric.StreamDurationMaxMilliseconds > destination.metric.StreamDurationMaxMilliseconds {
		destination.metric.StreamDurationMaxMilliseconds = source.metric.StreamDurationMaxMilliseconds
	}
	if source.lastDurationObservedAt.After(destination.lastDurationObservedAt) {
		destination.metric.StreamDurationLastMilliseconds = source.metric.StreamDurationLastMilliseconds
		destination.lastDurationObservedAt = source.lastDurationObservedAt
	}
	byKey[key] = destination
}

func mergeSSELatency(destination *httpTransactionLatencyMetrics, lastObservedAt *time.Time, samples, total, maximum, last int64, observedAt time.Time) {
	destination.samplesTotal += samples
	destination.totalMilliseconds += total
	if maximum > destination.maxMilliseconds {
		destination.maxMilliseconds = maximum
	}
	if observedAt.After(*lastObservedAt) {
		destination.lastMilliseconds = last
		*lastObservedAt = observedAt
	}
}
