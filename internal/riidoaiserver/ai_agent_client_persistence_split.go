package riidoaiserver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

// CurrentClientEventSeq returns the in-memory next-event-seq. The Sync wrapper
// compares this before/after the dev op to detect whether an event was appended,
// so a plain heartbeat persists only the small core item.
func (s *DevelopmentAIAgentClientStore) CurrentClientEventSeq() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.nextClientEventSeq
}

// restoreCorePreservingRest overwrites only the core collections (devices,
// credentials, daemons, agents, fixtures, seqs) from a core snapshot and leaves
// task threads, events, the event seq, and subscribers untouched. The daemon hot
// path (runtime-snapshot) uses this so it never reloads the heavy events/threads
// items, while in-memory threads/events stay this instance's truth.
func (s *DevelopmentAIAgentClientStore) restoreCorePreservingRest(core AIAgentClientCoreSnapshot) error {
	if core.SchemaVersion != AIAgentClientPersistenceSchemaVersion {
		return fmt.Errorf("unsupported ai agent client core snapshot schema_version %q", core.SchemaVersion)
	}
	deviceCredentials := make(map[string]deviceCredentialRecord, len(core.DeviceCredentials))
	for _, record := range core.DeviceCredentials {
		deviceID := strings.TrimSpace(record.DeviceID)
		if deviceID == "" {
			return errors.New("ai agent client core snapshot device credential device_id is required")
		}
		rawHash, err := hex.DecodeString(strings.TrimSpace(record.SecretHashSHA256))
		if err != nil || len(rawHash) != sha256.Size {
			return fmt.Errorf("ai agent client core snapshot device credential %s secret_hash_sha256 is invalid", deviceID)
		}
		var hash [sha256.Size]byte
		copy(hash[:], rawHash)
		deviceCredentials[deviceID] = deviceCredentialRecord{
			deviceID:         deviceID,
			machineID:        strings.TrimSpace(record.MachineID),
			secretHash:       hash,
			ownerPrincipalID: strings.TrimSpace(record.OwnerPrincipalID),
			workspaceID:      strings.TrimSpace(record.WorkspaceID),
			displayName:      strings.TrimSpace(record.DisplayName),
			issuedAt:         record.IssuedAt,
		}
	}
	daemons := make(map[string]DeviceDaemonRecord, len(core.Daemons))
	for _, daemon := range core.Daemons {
		deviceID := strings.TrimSpace(daemon.DeviceID)
		if deviceID == "" {
			return errors.New("ai agent client core snapshot daemon device_id is required")
		}
		daemons[deviceID] = copyDeviceDaemon(daemon)
	}
	agents := make(map[string]AgentClientRecord, len(core.Agents))
	for _, agent := range core.Agents {
		agentID := strings.TrimSpace(agent.AgentID)
		if agentID == "" {
			return errors.New("ai agent client core snapshot agent_id is required")
		}
		agents[agentID] = agent
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.workspaceID = strings.TrimSpace(core.WorkspaceID)
	s.devices = pruneLegacyRuntimeRecords(copyDevices(core.Devices))
	s.deviceCredentials = deviceCredentials
	s.nextDeviceCredentialSeq = core.NextDeviceCredentialSeq
	s.daemons = daemons
	s.nextDaemonCommand = core.NextDaemonCommand
	s.agents = agents
	s.fixtures = copyAgentOnboardingFixtures(core.Fixtures)
	s.ensureOnboardingFixtureColorsLocked()
	// taskThreads, taskThreadMessages, events, nextClientEventSeq,
	// subscribers/nextSubscriberID: preserved.
	return nil
}

// --- PersistentAIAgentClientStore split helpers ---

func (p *PersistentAIAgentClientStore) splitStore() (AIAgentClientSplitSnapshotStore, bool) {
	if p == nil || p.snapshotStore == nil {
		return nil, false
	}
	split, ok := p.snapshotStore.(AIAgentClientSplitSnapshotStore)
	return split, ok
}

// reloadAllSplit refreshes the full in-memory state from the three split items.
// Returns false (not an error) when the core item is absent (not migrated yet).
func (p *PersistentAIAgentClientStore) reloadAllSplit(ctx context.Context, split AIAgentClientSplitSnapshotStore) (bool, error) {
	core, ok, err := split.LoadCore(ctx)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	events, _, err := split.LoadEvents(ctx)
	if err != nil {
		return false, err
	}
	threads, _, err := split.LoadThreads(ctx)
	if err != nil {
		return false, err
	}
	return true, p.DevelopmentAIAgentClientStore.restoreSnapshotPreservingSubscribers(combinedFromSplit(core, events, threads))
}

// reloadCoreSplit refreshes ONLY the core collections (daemon hot path). Returns
// false when not migrated yet (core absent).
func (p *PersistentAIAgentClientStore) reloadCoreSplit(ctx context.Context, split AIAgentClientSplitSnapshotStore) (bool, error) {
	core, ok, err := split.LoadCore(ctx)
	if err != nil || !ok {
		return ok, err
	}
	return true, p.DevelopmentAIAgentClientStore.restoreCorePreservingRest(core)
}

func (p *PersistentAIAgentClientStore) saveAllSplit(ctx context.Context, split AIAgentClientSplitSnapshotStore) error {
	snap, err := p.DevelopmentAIAgentClientStore.snapshot(time.Now().UTC())
	if err != nil {
		return err
	}
	core, events, threads := splitFromCombined(snap)
	return split.WriteSplitAtomic(ctx, core, events, threads)
}

func (p *PersistentAIAgentClientStore) saveCoreSplit(ctx context.Context, split AIAgentClientSplitSnapshotStore) error {
	snap, err := p.DevelopmentAIAgentClientStore.snapshot(time.Now().UTC())
	if err != nil {
		return err
	}
	core, _, _ := splitFromCombined(snap)
	return split.SaveCore(ctx, core)
}

func (p *PersistentAIAgentClientStore) saveCoreEventsSplit(ctx context.Context, split AIAgentClientSplitSnapshotStore) error {
	snap, err := p.DevelopmentAIAgentClientStore.snapshot(time.Now().UTC())
	if err != nil {
		return err
	}
	core, events, _ := splitFromCombined(snap)
	return split.WriteCoreEvents(ctx, core, events)
}
