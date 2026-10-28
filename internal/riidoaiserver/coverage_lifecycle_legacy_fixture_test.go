package riidoaiserver

import "testing"

func TestCoverageLifecycleLegacyFenceFixtureRejectsCoveredAuthority(t *testing.T) {
	key := map[string]map[string]string{"pk": {"S": "DEVICE_POLICY#legacy"}, "sk": {"S": "REVISION"}}
	for _, tc := range []struct {
		name, schema, state string
		want                bool
	}{
		{"missing", "", "", true},
		{"legacy_unknown", "device_policy_revision/v1", "unknown", true},
		{"covered_complete", "device_policy_revision/v2", "complete", false},
		{"covered_unknown", "device_policy_revision/v2", "unknown", false},
		{"invalid_legacy_complete", "device_policy_revision/v1", "complete", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := membershipSharedDynamoFixture{items: map[string]map[string]map[string]string{}}
			if tc.schema != "" {
				fixture.items[membershipDynamoKey(key)] = map[string]map[string]string{"pk": key["pk"], "sk": key["sk"], "policy_schema": {"S": tc.schema}, "coverage_state": {"S": tc.state}}
			}
			action := membershipDynamoAction{Key: key, ConditionExpression: "(attribute_not_exists(pk) AND attribute_not_exists(sk)) OR (policy_schema = :schema AND coverage_state = :unknown)", ExpressionAttributeValues: map[string]map[string]string{":schema": {"S": "device_policy_revision/v1"}, ":unknown": {"S": "unknown"}}}
			if got := fixture.condition(t, &action); got != tc.want {
				t.Fatal("legacy fence fixture accepted the wrong authority")
			}
		})
	}
}
