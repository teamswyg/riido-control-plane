package riidoaiserver

func encodeSplitDynamoDBAIAgentClientSnapshot(snapshot AIAgentClientSnapshot, previousHashes map[string]string) ([]map[string]map[string]string, map[string]string, int64, error) {
	return encodeCachedSplitDynamoDBAIAgentClientSnapshot(snapshot, previousHashes, nil)
}

func encodeCachedSplitDynamoDBAIAgentClientSnapshot(snapshot AIAgentClientSnapshot, previousHashes map[string]string, cache map[string]dynamoDBAIAgentClientSnapshotPart) ([]map[string]map[string]string, map[string]string, int64, error) {
	parts, err := encodeDynamoDBAIAgentClientSnapshotParts(snapshot, cache)
	if err != nil {
		return nil, nil, 0, err
	}
	items := make([]map[string]map[string]string, 0, len(parts)+1)
	hashes := make(map[string]string, len(parts))
	var encodedBytes int64
	for _, part := range parts {
		hashes[part.name] = part.hash
		encodedBytes += int64(len(part.gzip))
		// Always couple membership to CURRENT, including unchanged membership.
		// A different writer may have replaced the physical part since our load.
		if part.name != dynamoDBAIAgentClientSnapshotPartTaskAgentMemberships && previousHashes != nil && previousHashes[part.name] == part.hash {
			continue
		}
		items = append(items, dynamoDBAIAgentClientSnapshotPartItem(snapshot, part))
	}
	items = append(items, dynamoDBAIAgentClientSnapshotManifestItem(snapshot, hashes))
	return items, hashes, encodedBytes, nil
}

func encodeDynamoDBAIAgentClientSnapshotParts(snapshot AIAgentClientSnapshot, cache map[string]dynamoDBAIAgentClientSnapshotPart) ([]dynamoDBAIAgentClientSnapshotPart, error) {
	values := []struct {
		name  string
		value any
	}{
		{name: dynamoDBAIAgentClientSnapshotPartDevices, value: snapshot.Devices},
		{name: dynamoDBAIAgentClientSnapshotPartDeviceCredentials, value: snapshot.DeviceCredentials},
		{name: dynamoDBAIAgentClientSnapshotPartDeviceConnections, value: snapshot.DeviceConnectionGrants},
		{name: dynamoDBAIAgentClientSnapshotPartDaemons, value: snapshot.Daemons},
		{name: dynamoDBAIAgentClientSnapshotPartAgents, value: snapshot.Agents},
		{name: dynamoDBAIAgentClientSnapshotPartFixtures, value: snapshot.Fixtures},
		{name: dynamoDBAIAgentClientSnapshotPartTaskRequestReceipts, value: snapshot.TaskRequestReceipts},
		{name: dynamoDBAIAgentClientSnapshotPartTaskThreads, value: snapshot.TaskThreads},
		{name: dynamoDBAIAgentClientSnapshotPartTaskAgentMemberships, value: snapshot.TaskAgentMemberships},
		{name: dynamoDBAIAgentClientSnapshotPartTaskThreadMessages, value: snapshot.TaskThreadMessages},
		{name: dynamoDBAIAgentClientSnapshotPartEvents, value: snapshot.Events},
		{name: dynamoDBAIAgentClientSnapshotPartOnboardingAgentCreates, value: snapshot.OnboardingAgentCreates},
		{name: dynamoDBAIAgentClientSnapshotPartOnboardingCreated, value: snapshot.OnboardingCreatedIDs},
		{name: dynamoDBAIAgentClientSnapshotPartOnboardingDone, value: snapshot.OnboardingCompletions},
		{name: dynamoDBAIAgentClientSnapshotPartOnboardingThreads, value: snapshot.OnboardingCompletionThreads},
	}
	parts := make([]dynamoDBAIAgentClientSnapshotPart, 0, len(values))
	for _, value := range values {
		part, err := encodeDynamoDBAIAgentClientSnapshotPart(value.name, value.value, cache)
		if err != nil {
			return nil, err
		}
		parts = append(parts, part)
	}
	return parts, nil
}
