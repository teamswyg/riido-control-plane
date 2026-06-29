package main

import (
	"fmt"
	"strings"
)

func verifyNextCommand(check readinessCheck) error {
	if !daemonResilienceCheck(check.ID) {
		return nil
	}
	if strings.Contains(check.NextCommand, "TestDaemon.*") {
		return fmt.Errorf("readiness check %s uses daemon test wildcard that can pass with no tests", check.ID)
	}
	return nil
}

func daemonResilienceCheck(id string) bool {
	switch id {
	case "daemon_network_disconnect_waiting", "all_servers_down_daemon_behavior":
		return true
	default:
		return false
	}
}
