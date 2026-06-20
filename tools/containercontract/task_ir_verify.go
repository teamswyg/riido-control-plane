package main

import "fmt"

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
