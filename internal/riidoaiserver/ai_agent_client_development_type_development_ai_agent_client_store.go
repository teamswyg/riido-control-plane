package riidoaiserver

import (
	"sync"
)

type DevelopmentAIAgentClientStore struct {
	mu                      sync.Mutex
	workspaceID             string
	devices                 []DeviceRecord
	deviceCredentials       map[string]deviceCredentialRecord
	nextDeviceCredentialSeq int
	daemons                 map[string]DeviceDaemonRecord
	nextDaemonCommand       int
	agents                  map[string]AgentClientRecord
	fixtures                []AgentOnboardingFixture
	taskThreads             map[string][]AIAgentTaskThreadRecord
	events                  []ClientStreamEvent
	subscribers             map[int]aiAgentClientSubscriber
	nextSubscriberID        int
}
