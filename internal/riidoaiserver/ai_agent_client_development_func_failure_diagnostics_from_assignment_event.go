package riidoaiserver

import (
	"strings"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func failureDiagnosticsFromAssignmentEvent(metadata map[string]string, message string) *AIAgentTaskThreadFailureDiagnostics {
	diagnostics := AIAgentTaskThreadFailureDiagnostics{
		ResultStatus:    strings.TrimSpace(metadata[metadatakeys.AssignmentResultStatus.String()]),
		FailureCategory: clientVisibleFailureCategory(metadata, message),
		Message:         clientVisibleFailureMessage(metadata, message),
	}
	if diagnostics.ResultStatus == "" && diagnostics.FailureCategory == "" && diagnostics.Message == "" {
		return nil
	}
	return &diagnostics
}

func clientVisibleFailureCategory(metadata map[string]string, message string) string {
	if _, ok := clientVisibleProviderAuthMessage(message); ok {
		return providerAuthFailureCategory
	}
	if _, ok := clientVisibleProviderLimitMessage(message); ok {
		return providerLimitFailureCategory
	}
	return strings.TrimSpace(metadata[metadatakeys.AssignmentFailureCategory.String()])
}
