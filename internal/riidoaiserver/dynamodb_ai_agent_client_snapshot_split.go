package riidoaiserver

const dynamoDBAIAgentClientSnapshotSplitStorageVersion = "riido-ai-agent-client-snapshot-split.v1"

const (
	dynamoDBAIAgentClientSnapshotPartDevices            = "DEVICES"
	dynamoDBAIAgentClientSnapshotPartDeviceCredentials  = "DEVICE_CREDENTIALS"
	dynamoDBAIAgentClientSnapshotPartDeviceConnections  = "DEVICE_CONNECTION_GRANTS"
	dynamoDBAIAgentClientSnapshotPartDaemons            = "DAEMONS"
	dynamoDBAIAgentClientSnapshotPartAgents             = "AGENTS"
	dynamoDBAIAgentClientSnapshotPartFixtures           = "FIXTURES"
	dynamoDBAIAgentClientSnapshotPartTaskThreads        = "TASK_THREADS"
	dynamoDBAIAgentClientSnapshotPartTaskThreadMessages = "TASK_THREAD_MESSAGES"
	dynamoDBAIAgentClientSnapshotPartEvents             = "EVENTS"
)

var dynamoDBAIAgentClientSnapshotRequiredPartNames = []string{
	dynamoDBAIAgentClientSnapshotPartDevices,
	dynamoDBAIAgentClientSnapshotPartDeviceCredentials,
	dynamoDBAIAgentClientSnapshotPartDaemons,
	dynamoDBAIAgentClientSnapshotPartAgents,
	dynamoDBAIAgentClientSnapshotPartFixtures,
	dynamoDBAIAgentClientSnapshotPartTaskThreads,
	dynamoDBAIAgentClientSnapshotPartEvents,
}

var dynamoDBAIAgentClientSnapshotOptionalPartNames = []string{
	dynamoDBAIAgentClientSnapshotPartDeviceConnections,
	dynamoDBAIAgentClientSnapshotPartTaskThreadMessages,
}

var dynamoDBAIAgentClientSnapshotPartNames = append(
	append([]string{}, dynamoDBAIAgentClientSnapshotRequiredPartNames...),
	dynamoDBAIAgentClientSnapshotOptionalPartNames...,
)

type dynamoDBAIAgentClientSnapshotPart struct {
	name string
	gzip string
	hash string
}
