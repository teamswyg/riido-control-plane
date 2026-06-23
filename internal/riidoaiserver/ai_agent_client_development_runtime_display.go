package riidoaiserver

func dedupeRuntimesByKindForDisplay(runtimes []RuntimeRecord) []RuntimeRecord {
	if len(runtimes) <= 1 {
		return runtimes
	}
	indexByKind := make(map[RuntimeKind]int, len(runtimes))
	out := make([]RuntimeRecord, 0, len(runtimes))
	for _, runtime := range runtimes {
		if runtime.Kind == "" {
			out = append(out, runtime)
			continue
		}
		if idx, ok := indexByKind[runtime.Kind]; ok {
			if preferRuntimeForDisplay(runtime, out[idx]) {
				out[idx] = runtime
			}
			continue
		}
		indexByKind[runtime.Kind] = len(out)
		out = append(out, runtime)
	}
	return out
}

func preferRuntimeForDisplay(candidate, current RuntimeRecord) bool {
	candidateLive := candidate.Availability == RuntimeAvailabilityOnline &&
		candidate.DetectionState == RuntimeDetectionStateDetected
	currentLive := current.Availability == RuntimeAvailabilityOnline &&
		current.DetectionState == RuntimeDetectionStateDetected
	if candidateLive != currentLive {
		return candidateLive
	}
	return candidate.LastDetectedAt.After(current.LastDetectedAt)
}
