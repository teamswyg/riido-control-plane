package requirements

const (
	DefaultManifest = "docs/20-domain/assignment-operation-journal.riido.json"
	ManifestSchema  = "riido-assignment-operation-journal.v1"
	EvidenceSchema  = "riido-assignment-operation-journal-evidence.v1"
	ExpectedID      = "assignment-operation-journal"
	ExpectedTask    = "RIID-4669"
)

var RequiredPorts = []string{
	"AssignmentOperationStore",
	"AssignmentOperationLoader",
	"AssignmentQueueReader",
	"AssignmentClaimer",
	"AssignmentActiveLeaseStore",
	"AssignmentProjectionReader",
}

var RequiredRecords = []string{
	"AssignmentOperationRecord",
	"AssignmentProjection",
	"AssignmentActiveLease",
	"AssignmentClaimResult",
}

var RequiredReplayRules = []string{
	"validate-before-apply",
	"stable-replay-order",
	"dedupe-task-events",
	"cross-task-same-seq",
	"track-next-sequences",
	"sort-replayed-events",
	"rebuild-current-assignment",
	"rebuild-agent-index",
}

var RequiredConstants = []string{
	"AssignmentOperationSchemaVersion",
	"AssignmentProjectionSchemaVersion",
	"AssignmentAgentActiveSchemaVersion",
	"DefaultAssignmentActiveLeaseSeconds",
}
