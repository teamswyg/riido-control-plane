package riidoaiserver

import (
	"strings"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func failureDiagnosticsFromAssignmentEvent(metadata map[string]string, message string) *AIAgentTaskThreadFailureDiagnostics {
	diagnostics := AIAgentTaskThreadFailureDiagnostics{
		ResultStatus:    strings.TrimSpace(metadata[metadatakeys.AssignmentResultStatus.String()]),
		FailureCategory: strings.TrimSpace(metadata[metadatakeys.AssignmentFailureCategory.String()]),
		Message:         clientVisibleFailureMessage(metadata, message),
	}
	if diagnostics.ResultStatus == "" && diagnostics.FailureCategory == "" && diagnostics.Message == "" {
		return nil
	}
	return &diagnostics
}
