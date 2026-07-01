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
		return verifyExternalRepoTestCommand(check)
	}
	if strings.Contains(check.NextCommand, "TestDaemon.*") {
		return fmt.Errorf("readiness check %s uses daemon test wildcard that can pass with no tests", check.ID)
	}
	return verifyExternalRepoTestCommand(check)
}

func verifyExternalRepoTestCommand(check readinessCheck) error {
	if referencesDaemonRepo(check) &&
		strings.Contains(check.NextCommand, "go test") &&
		!strings.Contains(check.NextCommand, "cd ../riido-daemon") {
		return fmt.Errorf("readiness check %s must run daemon go tests from ../riido-daemon", check.ID)
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

func referencesDaemonRepo(check readinessCheck) bool {
	for _, ref := range check.EvidenceRefs {
		if strings.HasPrefix(ref.Path, "riido-daemon:") {
			return true
		}
	}
	return false
}
