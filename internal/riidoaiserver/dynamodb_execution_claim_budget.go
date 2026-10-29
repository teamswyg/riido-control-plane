package riidoaiserver

import (
	"encoding/json"
	"errors"
	"strings"
)

// Preserve the six arbitrary 400 KiB item budget. One strictly bounded private
// claim and one bounded policy revision update may accompany a stale repair.
func validateExecutionClaimTerminalBudget(actions []json.RawMessage) error {
	base, claims, policies := 0, 0, 0
	for _, raw := range actions {
		var action dynamoDBTransactWriteItem
		if err := json.Unmarshal(raw, &action); err != nil {
			return err
		}
		if strings.HasPrefix(devicePolicyActionKey(action), "DEVICE_POLICY#") {
			policies++
			if len(raw) > 4<<10 || (!isDevicePolicyAdvance(action) && !isBoundedCoverageClaimAction(action)) {
				return errors.New("invalid bounded device policy action")
			}
			continue
		}
		put := action.Put
		if put == nil || !strings.HasPrefix(dynamoDBStringValue(put.Item, "pk"), "DEVICE_EXECUTION#") {
			base++
			continue
		}
		claims++
		if len(raw) > executionClaimItemBytes || action.Update != nil || action.Delete != nil || action.ConditionCheck != nil || put.ConditionExpression != "attribute_not_exists(pk) AND attribute_not_exists(sk)" {
			return errors.New("invalid bounded execution claim action")
		}
		if _, err := executionClaimFromItem(put.Item); err != nil {
			return err
		}
	}
	if base > contextTerminalBaseActions || claims > 1 || policies > 1 {
		return errors.New("terminal retention action budget exceeded")
	}
	return nil
}
