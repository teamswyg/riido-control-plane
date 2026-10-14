package riidoaiserver

func (m *storeOperationLatencyMetrics) observe(elapsedMS int64) {
	m.samplesTotal++
	m.totalMilliseconds += elapsedMS
	if elapsedMS > m.maxMilliseconds {
		m.maxMilliseconds = elapsedMS
	}
	m.lastMilliseconds = elapsedMS
}
