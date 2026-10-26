package riidoaiserver

import (
	"encoding/json"
	"errors"
	"strings"
)

// A lifecycle conflict must never masquerade as the legacy projection's
// recovery-watermark condition failure and acknowledge an uncommitted terminal.
func projectionConditionFailed(err error) bool {
	var api dynamoDBAPIError
	if !errors.As(err, &api) {
		return false
	}
	var failure struct {
		Type                string `json:"__type"`
		CancellationReasons []struct{ Code string }
	}
	if json.Unmarshal(api.body, &failure) != nil {
		return false
	}
	if strings.Contains(failure.Type, "ConditionalCheckFailedException") {
		return true
	}
	return strings.Contains(failure.Type, "TransactionCanceledException") && len(failure.CancellationReasons) > 0 && failure.CancellationReasons[0].Code == "ConditionalCheckFailed"
}
