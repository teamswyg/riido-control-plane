package riidoaiserver

import (
	"encoding/json"
	"fmt"
)

// Encoding is pure and independent of commit success. Retain only the most
// recent encoding per named part, under the snapshot store's command loop.
// Hashes used for write skipping still come from the durable commit observation.
func (s *DynamoDBAIAgentClientSnapshot) encodeSnapshot(snapshot AIAgentClientSnapshot, previousHashes map[string]string) ([]map[string]map[string]string, map[string]string, int64, error) {
	if s.partEncodings == nil {
		s.partEncodings = make(map[string]dynamoDBAIAgentClientSnapshotPart)
	}
	return encodeCachedSplitDynamoDBAIAgentClientSnapshot(snapshot, previousHashes, s.partEncodings)
}

func encodeDynamoDBAIAgentClientSnapshotPart(name string, value any, cache map[string]dynamoDBAIAgentClientSnapshotPart) (dynamoDBAIAgentClientSnapshotPart, error) {
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return dynamoDBAIAgentClientSnapshotPart{}, fmt.Errorf("encode AI Agent client snapshot part %s: %w", name, err)
	}
	hash := sha256Hex(jsonBytes)
	if part, ok := cache[name]; ok && part.hash == hash {
		return part, nil
	}
	gzipped, err := gzipBase64(jsonBytes)
	if err != nil {
		return dynamoDBAIAgentClientSnapshotPart{}, fmt.Errorf("gzip AI Agent client snapshot part %s: %w", name, err)
	}
	part := dynamoDBAIAgentClientSnapshotPart{
		name: name,
		gzip: gzipped,
		hash: hash,
	}
	if cache != nil {
		cache[name] = part
	}
	return part, nil
}
