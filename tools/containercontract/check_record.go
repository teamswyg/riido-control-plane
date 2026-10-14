package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/teamswyg/riido-control-plane/tools/containercontract/dockerfile"
)

type checkRecord struct {
	SchemaVersion           string       `json:"schema_version"`
	ID                      string       `json:"id"`
	Service                 string       `json:"service"`
	Status                  string       `json:"status"`
	Dockerfile              string       `json:"dockerfile"`
	FargateTaskDefinitionIR string       `json:"fargate_task_definition_ir,omitempty"`
	BuildStage              string       `json:"build_stage"`
	FinalBaseImage          string       `json:"final_base_image"`
	FinalUser               string       `json:"final_user"`
	Entrypoint              []string     `json:"entrypoint"`
	ExposedPorts            []int        `json:"exposed_ports"`
	Loop                    evidenceLoop `json:"loop"`
	ChecksTotal             int          `json:"checks_total"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}

func newCheckRecord(contract imageContract, buildStage, finalStage *dockerfile.Stage, checks int) checkRecord {
	return checkRecord{
		SchemaVersion: checkSchemaVersion, ID: contract.ID, Service: contract.Service,
		Status: "verified", Dockerfile: filepath.ToSlash(contract.Dockerfile),
		FargateTaskDefinitionIR: filepath.ToSlash(contract.FargateTaskDefinitionIR),
		BuildStage:              buildStage.Alias, FinalBaseImage: finalStage.Base,
		FinalUser: finalStage.User, Entrypoint: append([]string(nil), finalStage.Entrypoint...),
		ExposedPorts: dockerfile.SortedInts(finalStage.Exposes), Loop: contract.Loop, ChecksTotal: checks,
	}
}

func verifyEvidenceMetadata(id string, assertions []string, loop evidenceLoop) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("id is required")
	}
	if len(assertions) == 0 {
		return errors.New("assertions are required")
	}
	for i, assertion := range assertions {
		if strings.TrimSpace(assertion) == "" {
			return fmt.Errorf("assertions[%d] is required", i)
		}
	}
	for name, value := range map[string]string{
		"observation": loop.Observation, "hypothesis": loop.Hypothesis,
		"execute": loop.Execute, "evaluate": loop.Evaluate, "retrospective": loop.Retrospective,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("loop.%s is required", name)
		}
	}
	return nil
}
