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
	if len(items) > assignmentContextMaxCommentPages {
		return errors.New("context capture item limit exceeded")
	}
	for _, item := range items {
		limit := contextCommentItemBytes
		if item["sk"]["S"] == dynamoDBAssignmentContextSnapshotSK {
			limit = contextSnapshotItemBytes
		} else if !strings.HasPrefix(item["sk"]["S"], dynamoDBAssignmentContextCommentPrefix) {
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
	return nil
}
