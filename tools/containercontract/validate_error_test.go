package main

import "testing"

func TestValidateContractRejectsIdentityAndMetadataDrift(t *testing.T) {
	for name, mutate := range map[string]func(*imageContract){
		"schema": func(c *imageContract) { c.SchemaVersion = "other" },
		"id":     func(c *imageContract) { c.ID = "" },
		"assert": func(c *imageContract) { c.Assertions = nil },
		"loop":   func(c *imageContract) { c.Loop.Execute = "" },
		"field":  func(c *imageContract) { c.Build.GoBuild.Output = "" },
	} {
		t.Run(name, func(t *testing.T) {
			contract, _ := writeContainerContractFixture(t, "65532:65532", false)
			mutate(&contract)
			if err := validateContract(contract); err == nil {
				t.Fatalf("expected validateContract error")
			}
		})
	}
}

func TestRequireFinalContractRejectsMissingRequiredValues(t *testing.T) {
	for name, mutate := range map[string]func(*finalContract){
		"ports":      func(f *finalContract) { f.ExposedPorts = nil },
		"entrypoint": func(f *finalContract) { f.Entrypoint = nil },
		"env":        func(f *finalContract) { f.Env = nil },
		"copy":       func(f *finalContract) { f.RequiredCopies[0].Source = "" },
	} {
		t.Run(name, func(t *testing.T) {
			final := fixtureFinalContract("65532:65532")
			mutate(&final)
			if err := requireFinalContract(final); err == nil {
				t.Fatalf("expected final contract error")
			}
		})
	}
}

func TestVerifyEvidenceMetadataRejectsBlankAssertion(t *testing.T) {
	loop := evidenceLoop{
		Observation:   "observe",
		Hypothesis:    "hypothesis",
		Execute:       "execute",
		Evaluate:      "evaluate",
		Retrospective: "retro",
	}
	if err := verifyEvidenceMetadata("id", []string{""}, loop); err == nil {
		t.Fatalf("expected blank assertion error")
	}
}
