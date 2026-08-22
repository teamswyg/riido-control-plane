package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"errors"
)

type controlPlaneOwnerRequest struct {
	Query         string          `json:"query"`
	OperationName string          `json:"operationName"`
	Variables     json.RawMessage `json:"variables"`
}

func decodeControlPlaneOwnerRequest(raw []byte) (controlPlaneOwnerOperation, error) {
	if len(raw) == 0 || rejectControlPlaneOwnerAmbiguousJSON(raw) != nil {
		return controlPlaneOwnerOperation{}, errors.New("invalid JSON")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var input controlPlaneOwnerRequest
	if decoder.Decode(&input) != nil || !emptyControlPlaneOwnerVariables(input.Variables) {
		return controlPlaneOwnerOperation{}, errors.New("invalid request")
	}
	operation, ok := lookupControlPlaneOwnerOperation(input.OperationName)
	if !ok || canonicalControlPlaneOwnerQuery(input.Query) != operation.CanonicalQuery {
		return operation, errors.New("unregistered operation")
	}
	return operation, nil
}

func emptyControlPlaneOwnerVariables(raw json.RawMessage) bool {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return true
	}
	if rejectControlPlaneOwnerAmbiguousJSON(raw) != nil {
		return false
	}
	var variables map[string]json.RawMessage
	return json.Unmarshal(raw, &variables) == nil && variables != nil && len(variables) == 0
}

func canonicalControlPlaneOwnerQuery(value string) string {
	result := make([]byte, 0, len(value))
	for index := 0; index < len(value); index++ {
		char := value[index]
		switch {
		case char == ' ' || char == '\t' || char == '\r' || char == '\n':
			continue
		case char == '{' || char == '}':
			result = append(result, char)
		case char >= 'A' && char <= 'Z', char >= 'a' && char <= 'z', char >= '0' && char <= '9', char == '_':
			result = append(result, char)
		default:
			return ""
		}
	}
	return string(result)
}
