package riidoaiserver

import (
	"sync"
)

type DevelopmentAIAgentClientStore struct {
	taskRequestReceipts     map[string]AIAgentTaskActionResponse
	mu                      sync.Mutex
	taskAssignedAgents      map[string][]string
	snapshotVersion         int64
	workspaceID             string
	expectedDaemonProfile   string
	daemonClientPolicy      DaemonClientCompatibilityPolicy
	devices                 []DeviceRecord
	deviceCredentials       map[string]deviceCredentialRecord
	deviceConnectionGrants  map[string]map[string]DeviceConnectionGrant
	nextDeviceCredentialSeq int
	daemons                 map[string]DeviceDaemonRecord
	nextDaemonCommand       int
	agents                  map[string]AgentClientRecord
	fixtures                []AgentOnboardingFixture
	taskThreads             map[string][]AIAgentTaskThreadRecord
	taskThreadMessages      map[string][]AIAgentTaskThreadHistoryMessage
	taskThreadProgressCache map[string]taskThreadProgressMessageCache
	taskThreadHistoryCache  map[string]taskThreadHistoryMessageCache
	eventStreamHrefs        map[string]string
	events                  []ClientStreamEvent
	subscribers             map[int]aiAgentClientSubscriber
	nextSubscriberID        int
}
