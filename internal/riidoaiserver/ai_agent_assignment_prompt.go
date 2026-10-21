package riidoaiserver

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

const (
	TaskContextRepositorySourceConnectedPullRequest   = "connected_pull_request"
	TaskContextRepositorySourceWorkspaceConnectedRepo = "workspace_connected_repository"
	taskContextPromptFallbackNotProvided              = "not provided"
	taskContextUnavailableDocumentContent             = "Task context was not available when this assignment was created. " +
		"Use the task_id as the stable work item reference, avoid assuming repository state, " +
		"and act on the explicit user instruction within the authorized scope. " +
		"Ask for missing repository or product context only when it is necessary for the requested action."
)

type AIAgentAssignmentPromptInput struct {
	UserInstruction string
	TaskID          string
	Context         AIAgentTaskContext
}

type AIAgentAssignmentPrompt struct {
	Prompt             string
	SelectedRepository AIAgentTaskContextRepository
	HasRepository      bool
	IntentGateRequired bool
}

func ComposeAIAgentAssignmentPrompt(input AIAgentAssignmentPromptInput) (AIAgentAssignmentPrompt, error) {
	taskID := strings.TrimSpace(input.TaskID)
	component := normalizeAIAgentTaskContextComponent(input.Context.Component)
	document := normalizeAIAgentTaskContextDocument(input.Context.Document)
	hierarchy := normalizeAIAgentTaskContextHierarchy(input.Context.Hierarchy)
	repository, hasRepository := SelectAIAgentTaskContextRepository(input.Context.Repositories)

	if taskID == "" {
		taskID = component.ID
	}
	if taskID == "" {
		return AIAgentAssignmentPrompt{}, errors.New("task_id or component.id is required")
	}
	if component.ID == "" {
		component.ID = taskID
	}
	if component.Title == "" && document.Content == "" {
		return AIAgentAssignmentPrompt{}, errors.New("task context title or document content is required")
	}
	intentClass := classifyTaskContextIntent(component, document, input.UserInstruction)

	var builder strings.Builder
	builder.WriteString("# Riido AI Agent Assignment\n\n")
	builder.WriteString("Use this immutable assignment snapshot as the runtime task context.\n")
	builder.WriteString("Provider-specific instruction placement is owned by riido-daemon.\n\n")

	builder.WriteString("## Task\n")
	writePromptLine(&builder, "title", component.Title)
	writePromptLine(&builder, "task_id", taskID)
	writePromptLine(&builder, "component_id", component.ID)
	writePromptLine(&builder, "component_type", component.ComponentType)
	if component.KeyNumber != "" {
		writePromptLine(&builder, "key_number", component.KeyNumber)
	}
	writePromptLine(&builder, "branch_name", component.BranchName)
	builder.WriteString("\n")

	builder.WriteString("## Repository\n")
	if hasRepository {
		writePromptLine(&builder, "full_name", repository.FullName)
		writePromptLine(&builder, "repository_url", repository.RepositoryURL)
		writePromptLine(&builder, "source", repository.Source)
	} else {
		writePromptLine(&builder, "full_name", "")
	}
	builder.WriteString("\n")

	builder.WriteString("## Hierarchy\n")
	writePromptReference(&builder, "project", hierarchy.Project)
	writePromptReference(&builder, "milestone", hierarchy.Milestone)
	writePromptReference(&builder, "parent_task", hierarchy.ParentTask)
	builder.WriteString("\n")

	builder.WriteString("## Document\n")
	writeTaskContextDocumentSource(&builder, document)
	if document.ContentFormat != "" {
		writePromptLine(&builder, "content_format", document.ContentFormat)
		builder.WriteString("\n")
	}
	writeTaskContextDocumentBody(&builder, document)
	if !strings.HasSuffix(builder.String(), "\n") {
		builder.WriteString("\n")
	}
	writeParentTaskContextDocuments(&builder, hierarchy)
	writeTaskContextResources(&builder, input.Context.Resources, input.Context.ResourceStatus, 2500)
	writePromptTaskInterpretation(&builder, intentClass)
	writePromptInteractionPolicy(&builder)

	if instruction := strings.TrimSpace(input.UserInstruction); instruction != "" {
		builder.WriteString("\n\n[사용자 추가 요청]\n" + instruction)
	}

	return AIAgentAssignmentPrompt{
		Prompt:             builder.String(),
		SelectedRepository: repository,
		HasRepository:      hasRepository,
		// Retained for legacy consumers; heuristic labels never gate new work.
		IntentGateRequired: false,
	}, nil
}

func SelectAIAgentTaskContextRepository(repositories []AIAgentTaskContextRepository) (AIAgentTaskContextRepository, bool) {
	normalized := make([]AIAgentTaskContextRepository, 0, len(repositories))
	for _, repository := range repositories {
		repository.ID = strings.TrimSpace(repository.ID)
		repository.FullName = safeAIAgentRepositoryFullName(repository.FullName)
		repository.RepositoryURL = safeAIAgentRepositoryURL(repository.RepositoryURL)
		repository.Source = strings.TrimSpace(repository.Source)
		if repository.FullName == "" && repository.RepositoryURL == "" {
			continue
		}
		normalized = append(normalized, repository)
	}
	if len(normalized) == 0 {
		return AIAgentTaskContextRepository{}, false
	}
	sort.SliceStable(normalized, func(i, j int) bool {
		leftRank := taskContextRepositorySourceRank(normalized[i].Source)
		rightRank := taskContextRepositorySourceRank(normalized[j].Source)
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		if normalized[i].FullName != normalized[j].FullName {
			return normalized[i].FullName < normalized[j].FullName
		}
		return normalized[i].ID < normalized[j].ID
	})
	return normalized[0], true
}

func taskContextRepositorySourceRank(source string) int {
	switch source {
	case TaskContextRepositorySourceConnectedPullRequest:
		return 0
	case TaskContextRepositorySourceWorkspaceConnectedRepo:
		return 1
	default:
		return 2
	}
}

func safeAIAgentRepositoryFullName(rawFullName string) string {
	return assignmentcontract.NormalizePublicGitHubRepositoryFullName(rawFullName)
}

func safeAIAgentRepositoryURL(rawURL string) string {
	return assignmentcontract.NormalizePublicGitHubRepositoryURL(rawURL)
}

func normalizeAIAgentTaskContextComponent(component AIAgentTaskContextComponent) AIAgentTaskContextComponent {
	component.ID = strings.TrimSpace(component.ID)
	component.ComponentType = strings.TrimSpace(component.ComponentType)
	component.Title = strings.TrimSpace(component.Title)
	component.KeyNumber = strings.TrimSpace(component.KeyNumber)
	component.BranchName = strings.TrimSpace(component.BranchName)
	return component
}

func normalizeAIAgentTaskContextDocument(document AIAgentTaskContextDocument) AIAgentTaskContextDocument {
	document.ID = strings.TrimSpace(document.ID)
	document.TiptapDocumentID = strings.TrimSpace(document.TiptapDocumentID)
	document = normalizeTaskContextAlternate(document, taskDocumentContextLimit)
	document.ContentFormat = strings.TrimSpace(document.ContentFormat)
	if strings.EqualFold(document.ContentFormat, "html") {
		document.ContentFormat = "markdown"
	}
	return document
}

func normalizeAIAgentTaskContextHierarchy(hierarchy AIAgentTaskContextHierarchy) AIAgentTaskContextHierarchy {
	hierarchy.Project = normalizeAIAgentTaskContextReference(hierarchy.Project)
	hierarchy.Milestone = normalizeAIAgentTaskContextReference(hierarchy.Milestone)
	hierarchy.ParentTask = normalizeAIAgentTaskContextReference(hierarchy.ParentTask)
	return hierarchy
}

func normalizeAIAgentTaskContextReference(reference AIAgentTaskContextReference) AIAgentTaskContextReference {
	reference.ID = strings.TrimSpace(reference.ID)
	reference.Title = strings.TrimSpace(reference.Title)
	reference.KeyNumber = strings.TrimSpace(reference.KeyNumber)
	if reference.Document != nil {
		document := *reference.Document
		document = normalizeTaskContextAlternate(document, parentDocumentContextLimit)
		document.ContentFormat = "markdown"
		reference.Document = &document
	}
	return reference
}

func writePromptLine(builder *strings.Builder, key, value string) {
	value = strings.Join(strings.Fields(value), " ")
	if value == "" {
		value = taskContextPromptFallbackNotProvided
	}
	builder.WriteString("- ")
	builder.WriteString(key)
	builder.WriteString(": ")
	builder.WriteString(value)
	builder.WriteString("\n")
}

func writePromptReference(builder *strings.Builder, key string, reference AIAgentTaskContextReference) {
	value := strings.TrimSpace(reference.Title)
	if reference.KeyNumber != "" {
		value = fmt.Sprintf("%s %s", formatTaskContextKeyNumber(reference.KeyNumber), value)
	}
	if value == "" {
		value = reference.ID
	}
	writePromptLine(builder, key, value)
}

func formatTaskContextKeyNumber(keyNumber string) string {
	keyNumber = strings.TrimSpace(keyNumber)
	if keyNumber == "" || strings.HasPrefix(keyNumber, "RIID-") {
		return keyNumber
	}
	return "RIID-" + keyNumber
}
