package main

import "strings"

func verifyClassification(document manifest) bool {
	return document.Classification.Code == "baseline_go_ci_native_parity_source_complete" &&
		strings.Contains(document.Classification.Meaning, "legacy workflow") &&
		len(document.Classification.DoesNotClaim) == 5
}
