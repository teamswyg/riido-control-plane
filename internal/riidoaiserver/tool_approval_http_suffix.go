package riidoaiserver

import "strings"

func toolApprovalDecisionSuffixApprovalID(suffix string) (string, bool) {
	return toolApprovalNestedSuffixID(suffix, "tool-approvals/", "/decision")
}

func toolApprovalWaitSuffixApprovalID(suffix string) (string, bool) {
	return toolApprovalNestedSuffixID(suffix, "tool-approvals/", "/wait")
}

func toolApprovalNestedSuffixID(suffix, prefix, tail string) (string, bool) {
	if !strings.HasPrefix(suffix, prefix) || !strings.HasSuffix(suffix, tail) {
		return "", false
	}
	id := strings.TrimSuffix(strings.TrimPrefix(suffix, prefix), tail)
	id = strings.Trim(id, "/")
	return id, id != ""
}
