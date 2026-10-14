package main

import (
	"encoding/json"
	"os"
	"strings"
)

type openAPISpec struct {
	Paths map[string]map[string]openAPIOperation `json:"paths"`
}

func isHTTPMethod(value string) bool {
	switch strings.ToLower(value) {
	case "get", "post", "put", "patch", "delete", "head", "options", "trace":
		return true
	default:
		return false
	}
}

type openAPIOperation struct {
	OperationID string            `json:"operationId"`
	Client      openAPIClientMeta `json:"x-riido-client"`
}

type openAPIClientMeta struct {
	GeneratedPath string `json:"generated_path"`
}

type generatedOperation struct {
	Method string
	Path   string
}

func loadOpenAPIGenerated(path string) (map[string]generatedOperation, operationCounts, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, operationCounts{}, err
	}
	var spec openAPISpec
	if err := json.Unmarshal(body, &spec); err != nil {
		return nil, operationCounts{}, err
	}
	ops := map[string]generatedOperation{}
	var counts operationCounts
	for route, byMethod := range spec.Paths {
		for method, op := range byMethod {
			if !isHTTPMethod(method) || op.Client.GeneratedPath == "" {
				continue
			}
			counts.Total++
			if strings.HasPrefix(op.Client.GeneratedPath, "v2.") {
				counts.V2++
			} else {
				counts.V1++
			}
			ops[op.Client.GeneratedPath] = generatedOperation{strings.ToUpper(method), route}
		}
	}
	return ops, counts, nil
}
