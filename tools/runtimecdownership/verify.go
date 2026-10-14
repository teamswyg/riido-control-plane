package main

import "fmt"

func verifyAll(root string, m manifest) (verifyResult, error) {
	if err := verifyIdentity(m); err != nil {
		return verifyResult{}, err
	}
	strategies, err := verifyStrategies(root, m)
	if err != nil {
		return verifyResult{}, err
	}
	policies, guards, forbidden, keys, err := verifyPublicBoundary(m)
	if err != nil {
		return verifyResult{}, err
	}
	infraLinks, err := verifyInfraBoundary(m)
	if err != nil {
		return verifyResult{}, err
	}
	loopFields, err := verifyLoop(m.Loop)
	if err != nil {
		return verifyResult{}, err
	}
	return verifyResult{
		Strategies: strategies, PublicPolicies: policies, PublicGuards: guards,
		ForbiddenItems: forbidden, InfraLinks: infraLinks, LoopFields: loopFields,
		StableKeyCount: keys, WorkflowCount: len(m.Redaction.Workflows),
		HardeningCount: len(m.Hardening), SupersedesCount: len(m.Supersedes),
	}, nil
}

func verifyIdentity(m manifest) error {
	if m.SchemaVersion != manifestSchema || m.ID != requiredID {
		return fmt.Errorf("unexpected manifest identity")
	}
	if m.RiidoTask != requiredTask || m.Runtime != requiredRuntime {
		return fmt.Errorf("unexpected runtime ownership target")
	}
	if m.GeneratedDoc != generatedDoc || m.Workflow != workflow || m.EvidenceArtifact != evidenceArtifact {
		return fmt.Errorf("unexpected runtime ownership reader evidence binding")
	}
	if len(m.Hardening) == 0 || len(m.Supersedes) == 0 {
		return fmt.Errorf("runtime ownership manifest must keep task lineage")
	}
	return nil
}

func verifyDoc(root, expected string) error {
	actual, err := readText(repoPath(root, generatedDoc))
	if err != nil {
		return err
	}
	if actual != expected {
		return fmt.Errorf("%s is stale; run go run ./tools/runtimecdownership -write-doc", generatedDoc)
	}
	return nil
}

func verifyLoop(loop evidenceLoop) (int, error) {
	if !nonEmpty(loop.Observation, loop.Hypothesis, loop.Execute, loop.Evaluate, loop.Retrospective) {
		return 0, fmt.Errorf("loop evidence must include observe, hypothesis, execute, evaluate, retrospective")
	}
	return 5, nil
}
