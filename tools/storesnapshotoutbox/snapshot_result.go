package main

import "fmt"

func reloadedResult(tc caseSpec, history int, state string, events int) caseEvidence {
	return caseEvidence{
		Name: tc.Name, Kind: tc.Kind, HistoryEvents: history,
		AssignmentState: state, TaskEvents: events, SnapshotRestored: true,
	}
}

func verifySnapshotResult(tc caseSpec, got caseEvidence) error {
	if got.HistoryEvents != tc.WantHistoryEvents ||
		got.AssignmentState != tc.WantAssignmentState ||
		got.TaskEvents != tc.WantTaskEvents {
		return fmt.Errorf("%s result=%+v", tc.Name, got)
	}
	return nil
}
