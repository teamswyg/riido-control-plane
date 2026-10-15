package riidoaiserver

import (
	"strconv"
	"strings"
	"testing"
)

func (f *membershipSharedDynamoFixture) condition(t *testing.T, p *membershipDynamoAction) bool {
	t.Helper()
	key := p.Key
	if key == nil {
		key = p.Item
	}
	previous := f.items[membershipDynamoKey(key)]
	switch strings.TrimSpace(p.ConditionExpression) {
	case "":
		return true
	case "attribute_not_exists(pk) AND attribute_not_exists(sk)", "attribute_not_exists(pk)":
		return previous == nil
	case "attribute_exists(pk) AND attribute_exists(sk)":
		return previous != nil
	case "attribute_not_exists(publication_committed)":
		_, exists := previous["publication_committed"]
		return !exists
	case "attribute_not_exists(last_event_seq) OR last_event_seq <= :last_event_seq":
		old, _ := strconv.ParseInt(previous["last_event_seq"]["N"], 10, 64)
		next, _ := strconv.ParseInt(p.ExpressionAttributeValues[":last_event_seq"]["N"], 10, 64)
		return old <= next
	case "part_hash = :member_hash":
		value, exists := previous["part_hash"]["S"]
		return exists && value == p.ExpressionAttributeValues[":member_hash"]["S"]
	case "part_hashes_json = :snapshot_hashes":
		return previous["part_hashes_json"]["S"] == p.ExpressionAttributeValues[":snapshot_hashes"]["S"]
	case "generation = :generation":
		return previous["generation"]["S"] == p.ExpressionAttributeValues[":generation"]["S"]
	case "active_assignment_id = :assignment_id":
		return previous["active_assignment_id"]["S"] == p.ExpressionAttributeValues[":assignment_id"]["S"]
	case "assignment_state = :terminal AND last_event_seq = :terminal_seq":
		return previous["assignment_state"]["S"] == p.ExpressionAttributeValues[":terminal"]["S"] && previous["last_event_seq"]["N"] == p.ExpressionAttributeValues[":terminal_seq"]["N"]
	case "assignment_state = :observed_state AND last_event_seq = :observed_seq":
		return previous["assignment_state"]["S"] == p.ExpressionAttributeValues[":observed_state"]["S"] && previous["last_event_seq"]["N"] == p.ExpressionAttributeValues[":observed_seq"]["N"]
	case "assignment_state = :queued AND agent_id = :agent_id AND last_event_seq = :expected_last_event_seq":
		return previous["assignment_state"]["S"] == p.ExpressionAttributeValues[":queued"]["S"] && previous["agent_id"]["S"] == p.ExpressionAttributeValues[":agent_id"]["S"] && previous["last_event_seq"]["N"] == p.ExpressionAttributeValues[":expected_last_event_seq"]["N"]
	case "(attribute_not_exists(pk) AND attribute_not_exists(sk)) OR lease_expires_unix_ms <= :claim_started_unix_ms":
		old, _ := strconv.ParseInt(previous["lease_expires_unix_ms"]["N"], 10, 64)
		next, _ := strconv.ParseInt(p.ExpressionAttributeValues[":claim_started_unix_ms"]["N"], 10, 64)
		return previous == nil || old <= next
	default:
		t.Errorf("unsupported Dynamo condition %q", p.ConditionExpression)
		return false
	}
}
