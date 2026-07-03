package main

import (
	"fmt"

	"github.com/teamswyg/riido-control-plane/tools/snapshotcqrsgate/requirements"
)

type result struct {
	Operations          int
	Signals             int
	DecisionRules       int
	ForbiddenAttributes int
}

func verify(repoRoot string, m manifest, checkDoc bool) (result, error) {
	if err := verifyHeader(m); err != nil {
		return result{}, err
	}
	if checkDoc {
		if err := verifyDoc(repoRoot, m); err != nil {
			return result{}, err
		}
	}
	operations, err := verifyOperations(m)
	if err != nil {
		return result{}, err
	}
	signals, err := verifySignals(m)
	if err != nil {
		return result{}, err
	}
	if err := verifyDecisionRules(m); err != nil {
		return result{}, err
	}
	if err := verifyForbiddenAttributes(m); err != nil {
		return result{}, err
	}
	if err := verifyLoop(m.Loop); err != nil {
		return result{}, err
	}
	return result{operations, signals, len(m.DecisionRules), len(m.ForbiddenTraceAttributes)}, nil
}

func verifyHeader(m manifest) error {
	if m.SchemaVersion != requirements.ManifestSchema || m.ID != requirements.RequiredID || m.RiidoTask != requirements.RequiredTask {
		return fmt.Errorf("unexpected manifest identity")
	}
	if m.HumanDoc != requirements.RequiredHumanDoc {
		return fmt.Errorf("unexpected human_doc %q", m.HumanDoc)
	}
	if m.GeneratedDoc != requirements.RequiredHumanDoc || m.Workflow != requirements.Workflow || m.EvidenceArtifact != requirements.EvidenceArtifact {
		return fmt.Errorf("unexpected snapshot CQRS reader evidence binding")
	}
	if m.Decision.Scope != "ai_agent_client_snapshot_only" || m.Decision.StoreWideCQRS {
		return fmt.Errorf("decision must stay scoped to AI Agent client snapshot")
	}
	return nil
}
