package main

import "os"

func newRecord(spec workflowSpec) workflowRecord {
	return workflowRecord{
		ID:              spec.ID,
		Path:            spec.Path,
		SummaryArtifact: spec.SummaryArtifact,
		SummaryPath:     spec.SummaryPath,
		SensitiveInputs: spec.SensitiveInputs,
	}
}

func newRunRecord() runRecord {
	return runRecord{
		ID:      os.Getenv("GITHUB_RUN_ID"),
		Attempt: os.Getenv("GITHUB_RUN_ATTEMPT"),
		SHA:     os.Getenv("GITHUB_SHA"),
		RefName: os.Getenv("GITHUB_REF_NAME"),
		Event:   os.Getenv("GITHUB_EVENT_NAME"),
	}
}

func newRedaction(spec workflowSpec) redactionAssertion {
	return redactionAssertion{
		SummaryOnly:        true,
		NoRawSecrets:       true,
		NoRawEndpoints:     true,
		NoAWSResourceIDs:   true,
		AllowedFields:      spec.AllowedFields,
		SensitiveFieldRefs: spec.SensitiveInputs,
	}
}
