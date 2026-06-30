package main

import (
	"fmt"
	"regexp"
	"strings"
)

var nextCommandPlaceholderPattern = regexp.MustCompile(`<[A-Za-z][A-Za-z0-9_.:-]*>`)

func verifyNextCommand(check readinessCheck) error {
	if nextCommandPlaceholderPattern.MatchString(check.NextCommand) {
		return fmt.Errorf("readiness check %s next command must not contain placeholder tokens", check.ID)
	}
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
