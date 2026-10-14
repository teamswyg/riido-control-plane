package main

import "strings"

func evidenceUploadCoveragePaths(text string, uploadPaths []string) []string {
	covered := append([]string{}, uploadPaths...)
	for _, copy := range copiedEvidenceOutPaths(text) {
		if stringInSlice(copy.Destination, uploadPaths) {
			covered = append(covered, copy.Source)
		}
	}
	return uniqueStrings(covered)
}

type copiedEvidenceOutPath struct {
	Source      string
	Destination string
}

func copiedEvidenceOutPaths(text string) []copiedEvidenceOutPath {
	lines := strings.Split(text, "\n")
	var out []copiedEvidenceOutPath
	for _, line := range lines {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 3 || fields[0] != "cp" {
			continue
		}
		out = append(out, copiedEvidenceOutPath{
			Source:      cleanWorkflowValue(fields[1]),
			Destination: cleanWorkflowValue(fields[2]),
		})
	}
	return out
}

func stringInSlice(value string, values []string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
