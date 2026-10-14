package main

import (
	"encoding/json"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

type adapterResult struct {
	AuthorizedStatus   int
	MissingScopeStatus int
	UnconfiguredStatus int
	SchemaVersion      string
	Fields             map[string]bool
	HTTPBreakdownRows  int
	StoreBreakdownRows int
}

func exerciseAdapter(m manifest) (adapterResult, error) {
	okStatus, snapshot, fields, err := callAuthorizedMetrics(m)
	if err != nil {
		return adapterResult{}, err
	}
	missingScopeStatus, err := callMissingScope(m)
	if err != nil {
		return adapterResult{}, err
	}
	unconfiguredStatus, err := callUnconfigured(m)
	if err != nil {
		return adapterResult{}, err
	}
	return adapterResult{
		AuthorizedStatus:   okStatus,
		MissingScopeStatus: missingScopeStatus,
		UnconfiguredStatus: unconfiguredStatus,
		SchemaVersion:      snapshot.SchemaVersion,
		Fields:             fields,
		HTTPBreakdownRows:  len(snapshot.HTTPTransactions),
		StoreBreakdownRows: len(snapshot.StoreOperations),
	}, nil
}

func decodeSnapshot(body []byte) (riidoaiserver.MetricsSnapshot, map[string]bool, error) {
	var snapshot riidoaiserver.MetricsSnapshot
	if err := json.Unmarshal(body, &snapshot); err != nil {
		return snapshot, nil, err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return snapshot, nil, err
	}
	return snapshot, rawKeys(raw), nil
}

func rawKeys(raw map[string]json.RawMessage) map[string]bool {
	out := make(map[string]bool, len(raw))
	for key := range raw {
		out[key] = true
	}
	return out
}
