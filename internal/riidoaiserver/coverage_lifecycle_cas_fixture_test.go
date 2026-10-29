package riidoaiserver

import (
	"maps"
	"testing"
)

func TestCoverageLifecycleFixtureRequiresEveryCoverageCASField(t *testing.T) {
	key := map[string]map[string]string{"pk": {"S": "DEVICE_POLICY#dcov_fixture"}, "sk": {"S": "REVISION"}}
	item := map[string]map[string]string{"pk": key["pk"], "sk": key["sk"], "policy_schema": {"S": "device_policy_revision/v2"}, "generation": {"S": "cov_fixture"}, "credential_generation": {"S": "cred_fixture"}, "owner_principal_id": {"S": "owner"}, "coverage_state": {"S": "complete"}, "revision": {"N": "2"}, "unresolved": {"N": "1"}}
	values := map[string]map[string]string{":schema": {"S": "device_policy_revision/v2"}, ":generation": {"S": "cov_fixture"}, ":credential": {"S": "cred_fixture"}, ":owner": {"S": "owner"}, ":state": {"S": "complete"}, ":revision": {"N": "2"}, ":unresolved": {"N": "1"}}
	check := func(stored, parameters map[string]map[string]string) bool {
		fixture := membershipSharedDynamoFixture{items: map[string]map[string]map[string]string{membershipDynamoKey(key): stored}}
		return fixture.condition(t, &membershipDynamoAction{Key: key, ConditionExpression: executionCoverageCAS, ExpressionAttributeValues: parameters})
	}
	if !check(item, values) {
		t.Fatal("complete coverage CAS was rejected")
	}
	for field, parameter := range map[string]string{"policy_schema": ":schema", "generation": ":generation", "credential_generation": ":credential", "owner_principal_id": ":owner", "coverage_state": ":state", "revision": ":revision", "unresolved": ":unresolved"} {
		t.Run(field, func(t *testing.T) {
			missingValue := maps.Clone(values)
			delete(missingValue, parameter)
			missingField := maps.Clone(item)
			delete(missingField, field)
			wrongCurrent := maps.Clone(item)
			wrongCurrent[field] = map[string]string{"B": "invalid"}
			wrongType := maps.Clone(values)
			wrongType[parameter] = map[string]string{"B": "invalid"}
			extraType := maps.Clone(values)
			extraType[parameter] = maps.Clone(values[parameter])
			extraType[parameter]["B"] = "invalid"
			if check(item, missingValue) || check(missingField, values) || check(item, wrongType) || check(item, extraType) || check(wrongCurrent, values) {
				t.Fatal("coverage fixture accepted missing or malformed CAS authority")
			}
		})
	}
}
