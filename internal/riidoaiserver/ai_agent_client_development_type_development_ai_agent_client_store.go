package riidoaiserver

import (
	"sync"
)

type DevelopmentAIAgentClientStore struct {
	onboardingCreateMu          sync.Mutex
	onboardingAgentCreates      map[string]onboardingAgentCreateReceipt
	taskRequestReceipts         map[string]AIAgentTaskActionResponse
	mu                          sync.Mutex
	executionCoverageEnrollment bool
	executionCoverageAtomic     bool
	workspaceID                 string
	expectedDaemonProfile       string
	daemonClientPolicy          DaemonClientCompatibilityPolicy
	devices                     []DeviceRecord
	deviceCredentials           map[string]deviceCredentialRecord
	deviceConnectionGrants      map[string]map[string]DeviceConnectionGrant
	nextDeviceCredentialSeq     int
	daemons                     map[string]DeviceDaemonRecord
	nextDaemonCommand           int
	agents                      map[string]AgentClientRecord
	fixtures                    []AgentOnboardingFixture
	taskAgentMemberships        map[string]map[string]AIAgentTaskMembership
	taskThreads                 map[string][]AIAgentTaskThreadRecord
	taskThreadMessages          map[string][]AIAgentTaskThreadHistoryMessage
	taskThreadProgressCache     map[string]taskThreadProgressMessageCache
	taskThreadHistoryCache      map[string]taskThreadHistoryMessageCache
	eventStreamHrefs            map[string]string
	events                      []ClientStreamEvent
	subscribers                 map[int]aiAgentClientSubscriber
	nextSubscriberID            int
	profileColorIndex           func(n int) int
	onboardingCreated           map[string]bool
	onboardingCompletions       map[string]string
	// onboardingCompletionThreads keeps the completion thread per agent, so a repeat
	// completion can be answered after the task thread list is truncated.
	onboardingCompletionThreads map[string]AIAgentTaskThreadRecord
}
