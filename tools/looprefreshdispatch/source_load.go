package main

func loadRefreshCommandSources(root string, paths []string) ([]refreshCommandEvidence, error) {
	out := make([]refreshCommandEvidence, 0, len(paths))
	for _, path := range paths {
		source, err := loadRefreshCommands(repoPath(root, path))
		if err != nil {
			return nil, err
		}
		if source.SchemaVersion != refreshCommandsSchema {
			return nil, unsupportedRefreshCommandSchema(source.SchemaVersion)
		}
		out = append(out, source)
	}
	return out, nil
}
