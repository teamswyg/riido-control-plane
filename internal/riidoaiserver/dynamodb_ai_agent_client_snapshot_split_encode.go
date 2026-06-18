package riidoaiserver

import (
	"encoding/json"
	"fmt"
)

func encodeSplitDynamoDBAIAgentClientSnapshot(snapshot AIAgentClientSnapshot, previousHashes map[string]string) ([]map[string]map[string]string, map[string]string, int64, error) {
	parts, err := encodeDynamoDBAIAgentClientSnapshotParts(snapshot)
	if err != nil {
		return nil, nil, 0, err
	}
	items := make([]map[string]map[string]string, 0, len(parts)+1)
	hashes := make(map[string]string, len(parts))
	var encodedBytes int64
	for _, part := range parts {
		hashes[part.name] = part.hash
		encodedBytes += int64(len(part.gzip))
		if previousHashes != nil && previousHashes[part.name] == part.hash {
			continue
		}
		items = append(items, dynamoDBAIAgentClientSnapshotPartItem(snapshot, part))
	}
	items = append(items, dynamoDBAIAgentClientSnapshotManifestItem(snapshot, hashes))
	return items, hashes, encodedBytes, nil
}

func encodeDynamoDBAIAgentClientSnapshotParts(snapshot AIAgentClientSnapshot) ([]dynamoDBAIAgentClientSnapshotPart, error) {
	values := []struct {
		name  string
		value any
	}{
		{name: dynamoDBAIAgentClientSnapshotPartDevices, value: snapshot.Devices},
		{name: dynamoDBAIAgentClientSnapshotPartDeviceCredentials, value: snapshot.DeviceCredentials},
		{name: dynamoDBAIAgentClientSnapshotPartDaemons, value: snapshot.Daemons},
		{name: dynamoDBAIAgentClientSnapshotPartAgents, value: snapshot.Agents},
		{name: dynamoDBAIAgentClientSnapshotPartFixtures, value: snapshot.Fixtures},
		{name: dynamoDBAIAgentClientSnapshotPartTaskThreads, value: snapshot.TaskThreads},
		{name: dynamoDBAIAgentClientSnapshotPartEvents, value: snapshot.Events},
	}
	parts := make([]dynamoDBAIAgentClientSnapshotPart, 0, len(values))
	for _, value := range values {
		part, err := encodeDynamoDBAIAgentClientSnapshotPart(value.name, value.value)
		if err != nil {
			return nil, err
		}
		parts = append(parts, part)
	}
	return parts, nil
}

func encodeDynamoDBAIAgentClientSnapshotPart(name string, value any) (dynamoDBAIAgentClientSnapshotPart, error) {
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return dynamoDBAIAgentClientSnapshotPart{}, fmt.Errorf("encode AI Agent client snapshot part %s: %w", name, err)
	}
	gzipped, err := gzipBase64(jsonBytes)
	if err != nil {
		return dynamoDBAIAgentClientSnapshotPart{}, fmt.Errorf("gzip AI Agent client snapshot part %s: %w", name, err)
	}
	return dynamoDBAIAgentClientSnapshotPart{
		name: name,
		gzip: gzipped,
		hash: sha256Hex(jsonBytes),
	}, nil
}
