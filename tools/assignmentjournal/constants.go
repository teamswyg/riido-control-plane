package main

const (
	defaultManifest = "docs/20-domain/assignment-operation-journal.riido.json"
	manifestSchema  = "riido-assignment-operation-journal.v1"
	evidenceSchema  = "riido-assignment-operation-journal-evidence.v1"
	expectedID      = "assignment-operation-journal"
	expectedTask    = "RIID-4669"
)

var requiredPorts = []string{
	"AssignmentOperationStore",
	"AssignmentOperationLoader",
	"AssignmentQueueReader",
	"AssignmentClaimer",
	"AssignmentActiveLeaseStore",
	"AssignmentProjectionReader",
}

var requiredRecords = []string{
	"AssignmentOperationRecord",
	"AssignmentProjection",
	"AssignmentActiveLease",
	"AssignmentClaimResult",
}

var requiredReplayRules = []string{
	"validate-before-apply",
	"stable-replay-order",
	"dedupe-task-events",
	"cross-task-same-seq",
	"track-next-sequences",
	"sort-replayed-events",
	"rebuild-current-assignment",
	"rebuild-agent-index",
}

var requiredConstants = []string{
	"AssignmentOperationSchemaVersion",
	"AssignmentProjectionSchemaVersion",
	"AssignmentAgentActiveSchemaVersion",
	"DefaultAssignmentActiveLeaseSeconds",
}
