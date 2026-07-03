package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"sort"
	"strings"
)

func readOperations(path string) ([]operationRow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var spec openAPISpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, err
	}
	rows := operationRows(spec)
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].GeneratedPath == rows[j].GeneratedPath {
			return rows[i].OperationID < rows[j].OperationID
		}
		return rows[i].GeneratedPath < rows[j].GeneratedPath
	})
	return rows, nil
}

func operationRows(spec openAPISpec) []operationRow {
	var rows []operationRow
	for path, methods := range spec.Paths {
		for method, op := range methods {
			rows = append(rows, operationRow{
				Method:         strings.ToUpper(method),
				Path:           path,
				OperationID:    op.OperationID,
				Summary:        op.Summary,
				GeneratedPath:  op.Client.GeneratedPath,
				Deprecated:     op.Deprecated,
				Lifecycle:      op.Lifecycle,
				Replacement:    op.Replacement,
				RemovalHorizon: op.RemovalHorizon,
			})
		}
	}
	return rows
}

func fileHashes(paths map[string]string) (map[string]string, error) {
	out := map[string]string{}
	for name, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(data)
		out[name] = hex.EncodeToString(sum[:])
	}
	return out, nil
}
