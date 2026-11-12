package riidoaiserver

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestCheckoutLatestDateUsesInstantsAndNewestDuplicate(t *testing.T) {
	records := []json.RawMessage{
		json.RawMessage(`{"number":19,"updatedAt":"2026-11-11T01:00:00Z","htmlUrl":"https://github.com/example/repo/pull/19"}`),
		json.RawMessage(`{"number":21,"updatedAt":"2026-11-12T01:00:00Z","htmlUrl":"https://github.com/example/repo/pull/21"}`),
		json.RawMessage(`{"number":19,"updatedAt":"2026-11-12T10:00:01+09:00","htmlUrl":"https://github.com/example/repo/pull/19"}`),
	}
	for _, records := range [][]json.RawMessage{records, {records[2], records[1], records[0]}} {
		pulls := checkoutPullsForRepository(taskCheckoutPullRequests(records), &AssignmentWorktree{RepositoryFullName: "EXAMPLE/repo"})
		pull, ok := latestCheckoutPull(pulls)
		if !ok || len(pulls) != 2 || pull.number != "19" {
			t.Fatalf("latest duplicate/timezone was lost: %+v", pulls)
		}
	}
}

func TestCheckoutEqualOrMissingDatesHaveStableTieBreak(t *testing.T) {
	for _, date := range []string{"", "invalid", "2026-11-12T01:00:00Z"} {
		pulls := taskCheckoutPullRequests([]json.RawMessage{
			json.RawMessage(`{"number":9,"updatedAt":"` + date + `","htmlUrl":"https://github.com/example/repo/pull/9"}`),
			json.RawMessage(`{"number":11,"updatedAt":"` + date + `","htmlUrl":"https://github.com/example/repo/pull/11"}`),
		})
		pull, ok := latestCheckoutPull(pulls)
		if !ok || pull.number != "11" {
			t.Fatalf("equal dates must select the greater PR number: %+v", pull)
		}
	}
}

func TestCheckoutExplicitSelectionRequiresRepositoryContext(t *testing.T) {
	_, err := composeAssignRequestWithoutTaskContext("task", "task", AssignRequest{Prompt: "branch: staging"})
	if !errors.Is(err, errTaskCheckoutSelection) {
		t.Fatalf("explicit selection silently fell back without repository context: %v", err)
	}
}
