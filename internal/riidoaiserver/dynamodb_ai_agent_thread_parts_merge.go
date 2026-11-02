package riidoaiserver

// Preserve committed messages when this writer only changed other state.
// Conflicting local message changes must reload and reapply. The returned hash
// keeps tracking the caller's unchanged cache after the save.
func (s *DynamoDBAIAgentClientSnapshot) mergeCommittedThreadParts(items []map[string]map[string]string, requested AIAgentClientSnapshot, snapshot *AIAgentClientSnapshot) (map[string]string, error) {
	current := dynamoDBAIAgentClientSnapshotCurrentItem(items)
	if current == nil {
		return nil, nil
	}
	var hashes map[string]string
	if err := decodeConversationManifestHashes([]byte(dynamoDBStringValue(current, "part_hashes_json")), &hashes); err != nil {
		return nil, err
	}
	const name = dynamoDBAIAgentClientSnapshotPartTaskThreadMessages
	baseline := s.partHashes[name]
	if baseline == hashes[name] {
		return nil, nil
	}
	part, err := encodeDynamoDBAIAgentClientSnapshotPart(name, requested.TaskThreadMessages)
	if err != nil {
		return nil, err
	}
	// Older snapshots can omit the optional message part entirely.
	absentUnchanged := baseline == "" && len(requested.TaskThreadMessages) == 0
	if part.hash != baseline && !absentUnchanged {
		return nil, errTaskAgentSnapshotChanged
	}
	parts := dynamoDBAIAgentClientSnapshotItemsByPart(items)
	var committed AIAgentClientSnapshot
	if err := decodeCommittedAIAgentClientPart(name, current, parts[name], items, &committed); err != nil {
		return nil, err
	}
	snapshot.TaskThreadMessages = committed.TaskThreadMessages
	return map[string]string{name: part.hash}, nil
}
