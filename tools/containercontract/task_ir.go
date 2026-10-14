package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func verifyTaskDefinitionIR(contract imageContract) error {
	task, err := loadTaskDefinitionIR(contract.FargateTaskDefinitionIR)
	if err != nil {
		return err
	}
	if task.RuntimePlatform.OperatingSystemFamily != "LINUX" {
		return fmt.Errorf("runtime platform OS = %q, want LINUX", task.RuntimePlatform.OperatingSystemFamily)
	}
	if err := verifyTaskPorts(task, contract.Final.ExposedPorts); err != nil {
		return err
	}
	return verifyTaskEnv(task, contract.Final.Env)
}

func loadTaskDefinitionIR(path string) (taskDefinitionIR, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return taskDefinitionIR{}, err
	}
	var task taskDefinitionIR
	dec := json.NewDecoder(strings.NewReader(string(body)))
	if err := dec.Decode(&task); err != nil {
		return taskDefinitionIR{}, fmt.Errorf("decode fargate task definition IR: %w", err)
	}
	var extra struct{}
	if err := dec.Decode(&extra); err != io.EOF {
		return taskDefinitionIR{}, errors.New("decode fargate task definition IR: trailing data")
	}
	return task, nil
}

func verifyTaskPorts(task taskDefinitionIR, exposedPorts []int) error {
	taskPorts := map[int]bool{}
	for _, mapping := range task.Container.PortMappings {
		taskPorts[mapping.ContainerPort] = true
	}
	for _, port := range exposedPorts {
		if !taskPorts[port] {
			return fmt.Errorf("task definition does not expose container port %d", port)
		}
	}
	return nil
}

func verifyTaskEnv(task taskDefinitionIR, env map[string]string) error {
	taskEnv := map[string]string{}
	for _, value := range task.Container.Environment {
		taskEnv[value.Name] = value.Value
	}
	for key, want := range env {
		if taskEnv[key] != want {
			return fmt.Errorf("task definition env %s = %q, want %q", key, taskEnv[key], want)
		}
	}
	return nil
}
