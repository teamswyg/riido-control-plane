package riidoaiserver

import (
	"log"
	"strings"
)

const (
	daemonProfileMismatchLog = "event=daemon_profile_mismatch action=reject_and_prune"
	daemonProfilePrunedLog   = "event=daemon_profile_pruned action=delete_stale_profile_records"
)

func logDaemonProfileMismatch(
	deviceID, daemonID, expected, actual string,
	pruned daemonProfilePruneResult,
) {
	log.Printf(
		daemonProfileMismatchLog+" device_id=%q daemon_id=%q expected_profile=%q actual_profile=%q pruned_daemons=%d pruned_runtimes=%d",
		strings.TrimSpace(deviceID),
		strings.TrimSpace(daemonID),
		strings.TrimSpace(expected),
		strings.TrimSpace(actual),
		pruned.daemons,
		pruned.runtimes,
	)
}

func logDaemonProfilePruned(
	deviceID, expected, reason string,
	pruned daemonProfilePruneResult,
) {
	if pruned.daemons == 0 && pruned.runtimes == 0 {
		return
	}
	log.Printf(
		daemonProfilePrunedLog+" device_id=%q expected_profile=%q reason=%q pruned_daemons=%d pruned_runtimes=%d",
		strings.TrimSpace(deviceID),
		strings.TrimSpace(expected),
		strings.TrimSpace(reason),
		pruned.daemons,
		pruned.runtimes,
	)
}
