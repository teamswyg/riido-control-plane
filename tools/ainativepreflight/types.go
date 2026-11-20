package main

import "encoding/json"

type entry struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	ValueFrom string `json:"valueFrom"`
}
type container struct {
	EnvironmentFiles []json.RawMessage `json:"environmentFiles"`
	Name             string            `json:"name"`
	Environment      []entry           `json:"environment"`
	Secrets          []entry           `json:"secrets"`
}
type task struct {
	Containers []container `json:"containerDefinitions"`
}
