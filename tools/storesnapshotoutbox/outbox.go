package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func verifyOutboxCase(tc caseSpec) (caseEvidence, error) {
	ctx := context.Background()
	path := filepath.Join(os.TempDir(), "riido-outbox-"+tc.Name+".jsonl")
	_ = os.Remove(path)
	outbox, err := riidoaiserver.NewFileOutbox(path)
	if err != nil {
		return caseEvidence{}, err
	}
	store := riidoaiserver.NewStoreWithConfig(riidoaiserver.StoreConfig{Outbox: outbox})
	if _, err := assignAndPoll(ctx, store); err != nil {
		return caseEvidence{}, err
	}
	store.Close()
	events, err := readOutboxEvents(path)
	if err != nil {
		return caseEvidence{}, err
	}
	result := caseEvidence{Name: tc.Name, Kind: tc.Kind, Records: len(events), EventTypes: events}
	if result.Records != tc.WantRecords {
		return result, fmt.Errorf("%s records=%d want %d", tc.Name, result.Records, tc.WantRecords)
	}
	return result, verifyExpectedEvents(events, tc.WantEvents)
}

func readOutboxEvents(path string) ([]string, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var events []string
	for _, line := range strings.Split(strings.TrimSpace(string(body)), "\n") {
		var record riidoaiserver.OutboxRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return nil, err
		}
		events = append(events, record.Event.Type)
	}
	return events, nil
}
