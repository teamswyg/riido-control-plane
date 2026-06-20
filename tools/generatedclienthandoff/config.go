package main

import (
	"errors"
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
