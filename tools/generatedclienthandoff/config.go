package main

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

func normalizeConfig(cfg config) (config, error) {
	if missingRequiredPaths(cfg) {
		return config{}, errors.New("openapi, dsl, ir, core, react, and out are required")
	}
	if strings.TrimSpace(cfg.SourceCommit) == "" {
		return config{}, errors.New("source-commit is required")
	}
	if strings.TrimSpace(cfg.TargetBranch) == "" {
		return config{}, errors.New("target-branch is required")
	}
	if err := validateTargetBranch(cfg.TargetBranch); err != nil {
		return config{}, err
	}
	if strings.TrimSpace(cfg.SourceRef) == "" {
		cfg.SourceRef = cfg.SourceCommit
	}
	if strings.TrimSpace(cfg.GeneratedAt) == "" {
		cfg.GeneratedAt = time.Now().UTC().Format("2006-01-02")
	}
	return cfg, nil
}

func missingRequiredPaths(cfg config) bool {
	values := []string{cfg.OpenAPI, cfg.DSL, cfg.IR, cfg.Core, cfg.React, cfg.Out}
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return true
		}
	}
	return false
}

func sourceFiles(cfg config) map[string]string {
	return map[string]string{
		"openapi": cfg.OpenAPI,
		"dsl":     cfg.DSL,
		"ir":      cfg.IR,
		"core":    cfg.Core,
		"react":   cfg.React,
	}
}

func validateTargetBranch(branch string) error {
	if strings.Contains(branch, "react-query-") && !strings.HasPrefix(branch, "RIID-") {
		return fmt.Errorf("target branch %q must be a Riido work branchName, not a generated branch", branch)
	}
	return nil
}
