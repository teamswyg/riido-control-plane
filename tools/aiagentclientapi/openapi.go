package main

import (
	"encoding/json"
	"os"
	"strings"
)

type openAPISpec struct {
	Paths map[string]map[string]openAPIOperation `json:"paths"`
}

type openAPIOperation struct {
	Client struct {
		GeneratedPath string `json:"generated_path"`
	} `json:"x-riido-client"`
}

type operationSet struct {
	Generated map[string]struct{}
	Counts    operationCounts
}

func loadOpenAPIOperations(path string) (operationSet, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return operationSet{}, err
	}
	var spec openAPISpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return operationSet{}, err
	}
	out := operationSet{Generated: map[string]struct{}{}}
	for path, methods := range spec.Paths {
		for method, operation := range methods {
			if !isHTTPMethod(method) {
				continue
			}
			out.Counts.Total++
			if strings.HasPrefix(path, "/v1/client/ai-agent") {
				out.Counts.V1++
			}
			if strings.HasPrefix(path, "/v2/client/workspaces/{workspace_id}/ai-agent") {
				out.Counts.V2++
			}
			out.Generated[strings.TrimSpace(operation.Client.GeneratedPath)] = struct{}{}
		}
	}
	return out, nil
}

func isHTTPMethod(method string) bool {
	switch strings.ToLower(method) {
	case "get", "post", "patch", "delete":
		return true
	default:
		return false
	}
}
