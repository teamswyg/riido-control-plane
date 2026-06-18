package riidoaiserver

const dynamoDBAIAgentClientSnapshotSplitStorageVersion = "riido-ai-agent-client-snapshot-split.v1"

const (
	dynamoDBAIAgentClientSnapshotPartDevices           = "DEVICES"
	dynamoDBAIAgentClientSnapshotPartDeviceCredentials = "DEVICE_CREDENTIALS"
	dynamoDBAIAgentClientSnapshotPartDaemons           = "DAEMONS"
	dynamoDBAIAgentClientSnapshotPartAgents            = "AGENTS"
	dynamoDBAIAgentClientSnapshotPartFixtures          = "FIXTURES"
	dynamoDBAIAgentClientSnapshotPartTaskThreads       = "TASK_THREADS"
	dynamoDBAIAgentClientSnapshotPartEvents            = "EVENTS"
)

var dynamoDBAIAgentClientSnapshotPartNames = []string{
	dynamoDBAIAgentClientSnapshotPartDevices,
	dynamoDBAIAgentClientSnapshotPartDeviceCredentials,
	dynamoDBAIAgentClientSnapshotPartDaemons,
	dynamoDBAIAgentClientSnapshotPartAgents,
	dynamoDBAIAgentClientSnapshotPartFixtures,
	dynamoDBAIAgentClientSnapshotPartTaskThreads,
	dynamoDBAIAgentClientSnapshotPartEvents,
}

type dynamoDBAIAgentClientSnapshotPart struct {
	name string
	gzip string
	hash string
}
