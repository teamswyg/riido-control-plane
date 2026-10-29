package riidoaiserver

import (
	"encoding/json"
	"errors"
	"strings"
)

// Stored JSON is a conservative upper bound for DynamoDB item bytes. Reserve
// room for TTL, then bound 1 snapshot + 10 comment pages + lifecycle + quota <1.25 MiB.
// Current terminal transactions have at most 6 other items (400 KiB each), so
// their aggregate stays below DynamoDB's 4 MiB limit, including Update items.
const (
	contextSnapshotItemBytes   = 256 << 10
	contextCommentItemBytes    = 96 << 10
	contextLifecycleItemBytes  = 1 << 10
	contextQuotaItemBytes      = 1 << 10
	contextTerminalBaseActions = 6
)

func validateContextPayloadBudget(items []map[string]map[string]string) error {
	if len(items) > assignmentContextMaxCommentPages+1 {
		return errors.New("context capture item limit exceeded")
	}
	snapshots, payloads := 0, 0
	for _, item := range items {
		limit := contextCommentItemBytes
		switch {
		case item["sk"]["S"] == dynamoDBAssignmentContextSnapshotSK:
			snapshots++
			limit = contextSnapshotItemBytes
		case strings.HasPrefix(item["sk"]["S"], dynamoDBAssignmentContextCommentPrefix), strings.HasPrefix(item["sk"]["S"], assignmentContentPrefix):
			payloads++
		default:
			return errors.New("invalid context payload kind")
		}
		encoded, err := json.Marshal(item)
		if err != nil {
			return err
		}
		if len(encoded) > limit-128 {
			return errors.New("context capture byte budget exceeded")
		}
	}
	if snapshots > 1 || payloads > assignmentContextMaxCommentPages {
		return errors.New("context capture shared payload budget exceeded")
	}
	return nil
}
