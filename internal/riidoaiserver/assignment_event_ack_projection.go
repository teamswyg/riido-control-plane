package riidoaiserver

import (
	"encoding/hex"
	"encoding/json"
	"errors"
)

const maxAssignmentEventACKBytes = 16384

func assignmentProjectionEventACKField(item map[string]map[string]string, record AssignmentOperationRecord) (map[string]map[string]string, error) {
	encoded, err := encodeAssignmentEventACKs(record)
	if err != nil || encoded == "" {
		return item, err
	}
	item["event_ack_receipts_json"] = map[string]string{"S": encoded}
	return item, nil
}

func assignmentEventACKsFromJSON(value string, lastSeq int64) ([]assignmentEventACKReceipt, error) {
	if value == "" {
		return nil, nil
	}
	if len(value) > maxAssignmentEventACKBytes {
		return nil, errors.New("assignment event acknowledgements exceed budget")
	}
	var receipts []assignmentEventACKReceipt
	if err := json.Unmarshal([]byte(value), &receipts); err != nil || len(receipts) == 0 || len(receipts) > maxAssignmentEventACKReceipts {
		return nil, errors.New("assignment event acknowledgements are invalid")
	}
	var previous int64
	keys := make(map[string]bool)
	for _, r := range receipts {
		key, ke := hex.DecodeString(r.KeyHash)
		payload, pe := hex.DecodeString(r.PayloadHash)
		if ke != nil || pe != nil || len(key) != 32 || len(payload) != 32 || keys[r.KeyHash] || r.Seq <= previous || r.Seq > lastSeq || r.At.IsZero() || r.AssignmentUpdatedAt.IsZero() || (r.Type != EventAssignmentReady && !isTerminal(r.State)) {
			return nil, errors.New("assignment event acknowledgement is invalid")
		}
		keys[r.KeyHash] = true
		previous = r.Seq
	}
	return receipts, nil
}

func encodeAssignmentEventACKs(record AssignmentOperationRecord) (string, error) {
	receipts, err := assignmentEventACKsForRecord(record)
	if err != nil || len(receipts) == 0 {
		return "", err
	}
	data, err := json.Marshal(receipts)
	if err != nil {
		return "", err
	}
	if len(data) > maxAssignmentEventACKBytes {
		return "", errors.New("assignment event acknowledgements exceed budget")
	}
	return string(data), nil
}
