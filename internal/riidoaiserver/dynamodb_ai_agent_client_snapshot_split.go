package riidoaiserver

const dynamoDBAIAgentClientSnapshotSplitStorageVersion = "riido-ai-agent-client-snapshot-split.v1"

const (
	dynamoDBAIAgentClientSnapshotPartOnboardingAgentCreates = "ONBOARDING_AGENT_CREATES"
	dynamoDBAIAgentClientSnapshotPartTaskRequestReceipts    = "TASK_REQUEST_RECEIPTS"
	dynamoDBAIAgentClientSnapshotPartDevices                = "DEVICES"
	dynamoDBAIAgentClientSnapshotPartDeviceCredentials      = "DEVICE_CREDENTIALS"
	dynamoDBAIAgentClientSnapshotPartDeviceConnections      = "DEVICE_CONNECTION_GRANTS"
	dynamoDBAIAgentClientSnapshotPartDaemons                = "DAEMONS"
	dynamoDBAIAgentClientSnapshotPartAgents                 = "AGENTS"
	dynamoDBAIAgentClientSnapshotPartFixtures               = "FIXTURES"
	dynamoDBAIAgentClientSnapshotPartTaskAgentMemberships   = "TASK_AGENT_MEMBERSHIPS"
	dynamoDBAIAgentClientSnapshotPartTaskThreads            = "TASK_THREADS"
	dynamoDBAIAgentClientSnapshotPartTaskThreadMessages     = "TASK_THREAD_MESSAGES"
	dynamoDBAIAgentClientSnapshotPartEvents                 = "EVENTS"
	dynamoDBAIAgentClientSnapshotPartOnboardingCreated      = "ONBOARDING_CREATED_IDS"
	dynamoDBAIAgentClientSnapshotPartOnboardingDone         = "ONBOARDING_COMPLETIONS"
	dynamoDBAIAgentClientSnapshotPartOnboardingThreads      = "ONBOARDING_COMPLETION_THREADS"
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

// Optional parts are absent from snapshots written before they existed. A missing
// part decodes to its zero value.
var dynamoDBAIAgentClientSnapshotOptionalPartNames = []string{
	dynamoDBAIAgentClientSnapshotPartOnboardingAgentCreates,
	dynamoDBAIAgentClientSnapshotPartOnboardingCreated,
	dynamoDBAIAgentClientSnapshotPartOnboardingDone,
	dynamoDBAIAgentClientSnapshotPartOnboardingThreads,
	dynamoDBAIAgentClientSnapshotPartTaskRequestReceipts,
	dynamoDBAIAgentClientSnapshotPartTaskAgentMemberships,
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
