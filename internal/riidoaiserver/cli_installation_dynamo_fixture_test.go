package riidoaiserver

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"
)

// Independent, deliberately small fixture evaluator; unknown updates fail.
func (f *terminationDynamoFixture) policyUpdateAllowed(update *dynamoDBTransactWriteUpdateAction) bool {
	if !strings.HasPrefix(update.Key["pk"]["S"], "DEVICE_POLICY#") || update.Key["sk"]["S"] != "REVISION" || update.UpdateExpression != "SET policy_schema = if_not_exists(policy_schema, :schema), coverage_state = if_not_exists(coverage_state, :unknown), revision = if_not_exists(revision, :zero) + :one" || update.ConditionExpression != "(attribute_not_exists(pk) AND attribute_not_exists(sk)) OR (policy_schema = :schema AND coverage_state = :unknown AND revision > :zero AND revision < :max)" {
		return false
	}
	v := update.ExpressionAttributeValues
	if v[":schema"]["S"] != "device_policy_revision/v1" || v[":unknown"]["S"] != "unknown" || v[":one"]["N"] != "1" || v[":zero"]["N"] != "0" || v[":max"]["N"] != "18446744073709551615" {
		return false
	}
	stored := f.items[terminationItemKey(update.Key)]
	if len(stored) == 0 {
		return true
	}
	n, err := strconv.ParseUint(stored["revision"]["N"], 10, 64)
	return err == nil && n > 0 && n < math.MaxUint64 && stored["policy_schema"]["S"] == "device_policy_revision/v1" && stored["coverage_state"]["S"] == "unknown"
}

func (f *terminationDynamoFixture) applyPolicyUpdate(update *dynamoDBTransactWriteUpdateAction) {
	key := terminationItemKey(update.Key)
	old := f.items[key]
	n, _ := strconv.ParseUint(old["revision"]["N"], 10, 64)
	f.items[key] = map[string]map[string]string{"pk": update.Key["pk"], "sk": update.Key["sk"], "policy_schema": {"S": "device_policy_revision/v1"}, "coverage_state": {"S": "unknown"}, "revision": {"N": strconv.FormatUint(n+1, 10)}}
}

func (f *terminationDynamoFixture) queryPolicyClaims(w http.ResponseWriter, data []byte) {
	var req struct {
		ConsistentRead            bool
		KeyConditionExpression    string
		ExpressionAttributeValues map[string]map[string]string
		Limit                     int
	}
	if json.Unmarshal(data, &req) != nil || !req.ConsistentRead || req.KeyConditionExpression != "pk = :pk" || req.Limit != 100 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var items []map[string]map[string]string
	for _, item := range f.items {
		if item["pk"]["S"] == req.ExpressionAttributeValues[":pk"]["S"] {
			items = append(items, item)
		}
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"Items": items})
}
