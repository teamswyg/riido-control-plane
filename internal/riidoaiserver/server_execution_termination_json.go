package riidoaiserver

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func decodeExecutionTermination(w http.ResponseWriter, r *http.Request) (executionTerminationRequest, error) {
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	bad := errors.New("invalid execution termination observation")
	start, err := dec.Token()
	if err != nil {
		return executionTerminationRequest{}, err
	}
	if start != json.Delim('{') {
		return executionTerminationRequest{}, bad
	}
	fields := map[string]json.RawMessage{}
	for dec.More() {
		key, err := dec.Token()
		if err != nil {
			return executionTerminationRequest{}, err
		}
		name, ok := key.(string)
		if !ok || fields[name] != nil || (name != "schema_version" && name != "kind" && name != "started_invocations") {
			return executionTerminationRequest{}, bad
		}
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return executionTerminationRequest{}, err
		}
		fields[name] = raw
	}
	if _, err := dec.Token(); err != nil {
		return executionTerminationRequest{}, err
	}
	if _, err := dec.Token(); err != io.EOF {
		if err != nil {
			return executionTerminationRequest{}, err
		}
		return executionTerminationRequest{}, bad
	}
	var req executionTerminationRequest
	var count *uint64
	if len(fields) != 3 || json.Unmarshal(fields["schema_version"], &req.SchemaVersion) != nil || json.Unmarshal(fields["kind"], &req.Kind) != nil || json.Unmarshal(fields["started_invocations"], &count) != nil || count == nil || req.SchemaVersion != executionTerminationSchema || req.Kind != executionTerminationKind {
		return executionTerminationRequest{}, bad
	}
	req.StartedInvocations = *count
	return req, nil
}
