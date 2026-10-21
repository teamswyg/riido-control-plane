package riidoaiserver

import (
	"reflect"
	"sort"
)

// Concurrent changes to different keys commute; opposing changes to the same key
// must reload instead of resurrecting a deleted record or replacing a receipt.
func mergeOnboardingMap[T any](baseline, requested, committed map[string]T) (map[string]T, error) {
	result := make(map[string]T, len(committed))
	for key, value := range committed {
		result[key] = value
	}
	keys := map[string]bool{}
	for key := range baseline {
		keys[key] = true
	}
	for key := range requested {
		keys[key] = true
	}
	for key := range keys {
		before, hadBefore := baseline[key]
		after, hasAfter := requested[key]
		if hadBefore == hasAfter && reflect.DeepEqual(before, after) {
			continue
		}
		current, hasCurrent := committed[key]
		if hasCurrent == hasAfter && reflect.DeepEqual(current, after) {
			continue
		}
		if hasCurrent != hadBefore || !reflect.DeepEqual(current, before) {
			return nil, errTaskAgentSnapshotChanged
		}
		if hasAfter {
			result[key] = after
		} else {
			delete(result, key)
		}
	}
	return result, nil
}

func mergeCommittedOnboardingThreads(requested, committed map[string][]AIAgentTaskThreadRecord) map[string][]AIAgentTaskThreadRecord {
	result := make(map[string][]AIAgentTaskThreadRecord, len(requested))
	for taskID, threads := range requested {
		result[taskID] = append([]AIAgentTaskThreadRecord(nil), threads...)
	}
	for taskID, threads := range committed {
		known := map[string]bool{}
		for _, thread := range result[taskID] {
			known[thread.AssignmentID] = true
		}
		for _, thread := range threads {
			if isOnboardingSeedAssignmentID(thread.AssignmentID) && !known[thread.AssignmentID] {
				result[taskID] = append(result[taskID], copyTaskThread(thread))
			}
		}
		if len(result[taskID]) > aiAgentClientThreadsPerTaskLimit {
			sort.SliceStable(result[taskID], func(i, j int) bool { return result[taskID][i].StartedAt.Before(result[taskID][j].StartedAt) })
			retained := result[taskID][:0]
			for i, thread := range result[taskID] {
				if i >= len(result[taskID])-aiAgentClientThreadsPerTaskLimit || taskThreadNeedsRetention(thread) {
					retained = append(retained, thread)
				}
			}
			result[taskID] = retained
		}
	}
	return result
}
