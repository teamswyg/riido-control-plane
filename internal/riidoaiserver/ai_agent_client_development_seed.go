package riidoaiserver

import "time"

type developmentAIAgentClientSeed struct {
	devices     []DeviceRecord
	daemons     map[string]DeviceDaemonRecord
	agents      map[string]AgentClientRecord
	fixtures    []AgentOnboardingFixture
	taskThreads map[string][]AIAgentTaskThreadRecord
	messages    map[string][]AIAgentTaskThreadHistoryMessage
	events      []ClientStreamEvent
}

func NewDevelopmentAIAgentClientStore() *DevelopmentAIAgentClientStore {
	seed := newDevelopmentAIAgentClientSeed()
	return &DevelopmentAIAgentClientStore{
		workspaceID:             defaultAIAgentClientWorkspaceID,
		devices:                 seed.devices,
		deviceCredentials:       map[string]deviceCredentialRecord{},
		daemons:                 seed.daemons,
		nextDaemonCommand:       1,
		agents:                  seed.agents,
		fixtures:                seed.fixtures,
		taskThreads:             seed.taskThreads,
		taskThreadMessages:      seed.messages,
		taskThreadProgressCache: map[string]taskThreadProgressMessageCache{},
		subscribers:             map[int]aiAgentClientSubscriber{},
		events:                  seed.events,
		nextSubscriberID:        0,
		nextDeviceCredentialSeq: 0,
	}
}

func newDevelopmentAIAgentClientSeed() developmentAIAgentClientSeed {
	now := time.Date(2026, 5, 28, 6, 0, 0, 0, time.UTC)
	device := developmentPrimaryDevice(now)
	sharedDevice := developmentSharedDevice(now)
	daemon := developmentPrimaryDaemon(device, now)
	sharedDaemon := developmentSharedDaemon(sharedDevice, now)
	return developmentAIAgentClientSeed{
		devices: []DeviceRecord{device, sharedDevice},
		daemons: map[string]DeviceDaemonRecord{
			daemonStorageKey(daemon):       daemon,
			daemonStorageKey(sharedDaemon): sharedDaemon,
		},
		agents:      developmentSeedAgents(now),
		fixtures:    developmentSeedFixtures(),
		taskThreads: developmentSeedTaskThreads(now),
		messages:    map[string][]AIAgentTaskThreadHistoryMessage{},
		events:      developmentSeedEvents(device, daemon),
	}
}
