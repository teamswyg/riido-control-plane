package main

func generatedFiles(cfg config, hashes map[string]string, ops []operationRow) map[string]string {
	return map[string]string{
		"README.generated.md":           readme(cfg, hashes, ops),
		"apiHistory.generated.ts":       apiHistory(cfg, ops),
		"contractManifest.generated.ts": contractManifest(cfg, hashes, ops),
		"index.ts":                      indexBarrel(),
		"react.ts":                      reactBarrel(),
	}
}

func indexBarrel() string {
	return "export * from './aiAgentClient';\n" +
		"export * from './apiHistory.generated';\n" +
		"export * from './contractManifest.generated';\n"
}

func reactBarrel() string {
	return "export * from './aiAgentClient.react';\n"
}
