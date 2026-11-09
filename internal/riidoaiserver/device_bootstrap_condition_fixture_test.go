package riidoaiserver

import (
	"strings"
)

func bootstrapFixtureCondition(p *membershipDynamoAction, previous map[string]map[string]string) bool {
	switch strings.TrimSpace(p.ConditionExpression) {
	case "payload_sha256 = :bootstrap_hash":
		return previous["payload_sha256"]["S"] != "" && previous["payload_sha256"]["S"] == p.ExpressionAttributeValues[":bootstrap_hash"]["S"]
	case "policy_schema = :schema AND coverage_state = :state AND revision = :revision", "policy_schema = :schema AND coverage_state = :state AND transition_id = :transition AND revision = :revision":
		matches := previous["policy_schema"]["S"] == p.ExpressionAttributeValues[":schema"]["S"] && previous["coverage_state"]["S"] == p.ExpressionAttributeValues[":state"]["S"] && previous["revision"]["N"] == p.ExpressionAttributeValues[":revision"]["N"]
		if strings.Contains(p.ConditionExpression, "transition_id") {
			matches = matches && previous["transition_id"]["S"] == p.ExpressionAttributeValues[":transition"]["S"]
		}
		return matches
	default:
		return false
	}
}
