package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func run(cfg config) error {
	normalized, err := normalizeConfig(cfg)
	if err != nil {
		return err
	}
	ops, err := readOperations(normalized.OpenAPI)
	if err != nil {
		return err
	}
	previous, err := readPreviousManifest(normalized.PreviousManifest)
	if err != nil {
		return err
	}
	hashes, err := fileHashes(sourceFiles(normalized))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(normalized.Out, 0o755); err != nil {
		return err
	}
	if err := writeGeneratedFiles(normalized, hashes, ops); err != nil {
		return err
	}
	if strings.TrimSpace(normalized.PRBody) != "" {
		body := prBody(normalized, hashes, ops, previous)
		if err := os.WriteFile(normalized.PRBody, []byte(body), 0o644); err != nil {
			return fmt.Errorf("write pr body: %w", err)
		}
	}
	return nil
}

func writeGeneratedFiles(cfg config, hashes map[string]string, ops []operationRow) error {
	for name, body := range generatedFiles(cfg, hashes, ops) {
		if err := os.WriteFile(filepath.Join(cfg.Out, name), []byte(body), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", name, err)
		}
	}
	return nil
}

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
