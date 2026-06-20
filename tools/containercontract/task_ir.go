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
