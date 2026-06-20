package main

import "path/filepath"

func newCheckRecord(contract imageContract, buildStage, finalStage *stage, checks int) checkRecord {
	return checkRecord{
		SchemaVersion:           checkSchemaVersion,
		ID:                      contract.ID,
		Service:                 contract.Service,
		Status:                  "verified",
		Dockerfile:              filepath.ToSlash(contract.Dockerfile),
		FargateTaskDefinitionIR: filepath.ToSlash(contract.FargateTaskDefinitionIR),
		BuildStage:              buildStage.Alias,
		FinalBaseImage:          finalStage.Base,
		FinalUser:               finalStage.User,
		Entrypoint:              append([]string(nil), finalStage.Entrypoint...),
		ExposedPorts:            sortedInts(finalStage.Exposes),
		Loop:                    contract.Loop,
		ChecksTotal:             checks,
	}
}
