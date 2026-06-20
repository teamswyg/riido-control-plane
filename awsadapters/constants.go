package awsadapters

import internal "github.com/teamswyg/riido-control-plane/internal/riidoaiserver"

const (
	SchemaVersion                         = internal.SchemaVersion
	StoreSnapshotSchemaVersion            = internal.StoreSnapshotSchemaVersion
	OutboxRecordSchemaVersion             = internal.OutboxRecordSchemaVersion
	AIAgentClientPersistenceSchemaVersion = internal.AIAgentClientPersistenceSchemaVersion

	AssignmentQueued      = internal.AssignmentQueued
	EventAssignmentQueued = internal.EventAssignmentQueued
)
