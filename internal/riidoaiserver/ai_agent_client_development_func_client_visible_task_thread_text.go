package riidoaiserver

import (
	"strings"
)

func clientVisibleTaskThreadText(message string) string {
	message = stripRiidoLogBlocks(message)
	message = clientVisibleMarkdownLocalLinkPattern.ReplaceAllString(message, "$1")
	message = clientVisibleAngleLocalPathPattern.ReplaceAllString(message, "로컬 파일")
	message = clientVisibleApplicationSupportLocalPathPattern.ReplaceAllString(message, "로컬 파일")
	message = clientVisibleLocalPathPattern.ReplaceAllString(message, "로컬 파일")
	message = restoreClientVisibleInlineCode(message)
	message = clientVisibleLocalizedTaskThreadText(message)
	if confirmationMessage, ok := clientVisibleLocalApprovalMessage(message); ok {
		message = confirmationMessage
	}
	if limitMessage, ok := clientVisibleProviderLimitMessage(message); ok {
		message = limitMessage
	}
	return strings.TrimSpace(message)
}

func clientVisibleLocalizedTaskThreadText(message string) string {
	switch strings.TrimSpace(message) {
	case "agent assignment is queued",
		"agent is busy; task assignment was queued",
		"agent is busy; task comment was queued",
		"agent is busy; task thread message was queued":
		return clientMessageAgentBusyQueued
	case "agent work was stopped",
		"agent work was stopped by user request",
		"agent work was stopped by task participant removal",
		"agent assignment was replaced by a participant change",
		"context canceled",
		"context cancelled",
		"supervisor: stopped":
		return clientMessageTaskStopped
	case "agent work was stopped by agent delete":
		return clientMessageAgentDeleted
	case "agent work failed":
		return clientMessageTaskFailed
	case "agent work completed":
		return clientMessageTaskCompleted
	case "agent assignment was accepted by runtime",
		"agent assignment started from task participant",
		"agent progress updated",
		"agent work continued from task thread message",
		"agent work is running",
		"agent work started from task comment":
		return clientMessageTaskRunning
	case "agent assignment timed out after runtime went silent",
		"agent stop timed out after runtime went silent":
		return clientMessageTaskTimeout
	case "recovery requires provider session id; refusing fresh start":
		return clientMessageRecoveryBlocked
	default:
		return message
	}
}
