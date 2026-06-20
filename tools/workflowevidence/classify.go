package main

import "strings"

func hasExecutableStep(text string) bool {
	needles := []string{"go test", "go run", "pnpm", "npm", "terraform", "aws ", "curl", "docker"}
	for _, needle := range needles {
		if strings.Contains(text, needle) {
			return true
		}
	}
	return false
}

func classify(record workflowRecord, accepted map[string]acceptedGap, used map[string]bool) workflowRecord {
	if record.HasEvidenceOut && record.UploadsArtifact {
		record.Status = "covered"
		return record
	}
	if gap, ok := accepted[record.Path]; ok {
		used[record.Path] = true
		record.Status = "accepted_gap"
		record.Reason = gap.Reason
		record.Next = gap.Next
		return record
	}
	if !record.HasExecutable {
		record.Status = "metadata_only"
		return record
	}
	record.Status = "unregistered_gap"
	return record
}

func addRecord(result *auditResult, record workflowRecord) {
	switch record.Status {
	case "covered":
		result.Covered++
	case "accepted_gap":
		result.Accepted++
	case "unregistered_gap":
		result.Unregistered = append(result.Unregistered, record.Path)
	}
}
