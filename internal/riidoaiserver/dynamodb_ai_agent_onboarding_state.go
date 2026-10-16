package riidoaiserver

import "strings"

// Keep the last observed values, rather than just hashes: ordinary cached writers
// may alter another profile while another replica commits a one-time completion.
func onboardingOwnedSnapshot(snapshot AIAgentClientSnapshot) AIAgentClientSnapshot {
	return AIAgentClientSnapshot{
		OnboardingCreatedIDs:        append([]string(nil), snapshot.OnboardingCreatedIDs...),
		OnboardingCompletions:       cloneOnboardingCompletedTasks(snapshot.OnboardingCompletions),
		OnboardingCompletionThreads: cloneOnboardingCompletionThreads(snapshot.OnboardingCompletionThreads),
		TaskRequestReceipts:         onboardingReceipts(snapshot.TaskRequestReceipts),
	}
}

func onboardingReceipts(receipts map[string]AIAgentTaskActionResponse) map[string]AIAgentTaskActionResponse {
	result := cloneTaskRequestReceipts(receipts)
	for key := range result {
		if !strings.HasPrefix(key, onboardingCompletionReceiptKeyPrefix) {
			delete(result, key)
		}
	}
	return result
}

// Apply only changes made since this adapter last observed onboarding state.
// CURRENT and these parts are read atomically; the existing CURRENT condition
// fences a competing commit between this read and the transaction write.
func (s *DynamoDBAIAgentClientSnapshot) mergeCommittedOnboardingSnapshot(items []map[string]map[string]string, requested *AIAgentClientSnapshot) error {
	current := dynamoDBAIAgentClientSnapshotCurrentItem(items)
	parts := dynamoDBAIAgentClientSnapshotItemsByPart(items)
	var committed AIAgentClientSnapshot
	for _, name := range onboardingOwnedPartNames {
		if err := decodeCommittedAIAgentClientPart(name, current, parts[name], &committed); err != nil {
			return err
		}
	}
	baseline := s.onboardingBaseline
	forgetDeletedOnboardingIncarnations(baseline, *requested, &committed)
	created, err := mergeOnboardingMap(onboardingCreatedAgentsFromSnapshot(baseline.OnboardingCreatedIDs), onboardingCreatedAgentsFromSnapshot(requested.OnboardingCreatedIDs), onboardingCreatedAgentsFromSnapshot(committed.OnboardingCreatedIDs))
	if err != nil {
		return err
	}
	completed, err := mergeOnboardingMap(baseline.OnboardingCompletions, requested.OnboardingCompletions, committed.OnboardingCompletions)
	if err != nil {
		return err
	}
	threads, err := mergeOnboardingMap(baseline.OnboardingCompletionThreads, requested.OnboardingCompletionThreads, committed.OnboardingCompletionThreads)
	if err != nil {
		return err
	}
	receipts, err := mergeOnboardingMap(baseline.TaskRequestReceipts, onboardingReceipts(requested.TaskRequestReceipts), onboardingReceipts(committed.TaskRequestReceipts))
	if err != nil {
		return err
	}
	requested.OnboardingCreatedIDs = onboardingCreatedAgentIDsSnapshot(created)
	requested.OnboardingCompletions = completed
	requested.OnboardingCompletionThreads = threads
	requested.TaskRequestReceipts = cloneTaskRequestReceipts(requested.TaskRequestReceipts)
	if requested.TaskRequestReceipts == nil {
		requested.TaskRequestReceipts = map[string]AIAgentTaskActionResponse{}
	}
	for key := range requested.TaskRequestReceipts {
		if strings.HasPrefix(key, onboardingCompletionReceiptKeyPrefix) {
			delete(requested.TaskRequestReceipts, key)
		}
	}
	for key, receipt := range receipts {
		requested.TaskRequestReceipts[key] = receipt
	}
	requested.TaskThreads = mergeCommittedOnboardingThreads(requested.TaskThreads, committed.TaskThreads)
	return nil
}
