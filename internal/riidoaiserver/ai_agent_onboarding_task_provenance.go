package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

// errOnboardingTaskProvenance rejects a task that the monolith does not record as
// the viewer's own first-task bundle. It maps to the default 400 response.
var errOnboardingTaskProvenance = errors.New("onboarding completion requires the first task bundle created for this user")

// errOnboardingTaskProvenanceUnavailable means the monolith could not answer. The
// completion is refused (fail closed) and the handler responds 502.
var errOnboardingTaskProvenanceUnavailable = errors.New("onboarding task provenance could not be verified")

// onboardingFirstTaskBundleReader asks the monolith with the viewer's Riido Bearer JWT.
type onboardingFirstTaskBundleReader interface {
	IsOnboardingFirstTaskBundleTask(ctx context.Context, bearerToken, workspaceID, taskID string) (bool, error)
}

// IsOnboardingFirstTaskBundleTask resolves the task's team through the monolith and
// then asks whether the viewer's own first-task bundle contains the task. A task
// outside the workspace, an inaccessible task, or a monolith "no" returns false.
// Only transport or protocol failures return an error.
func (c *AIAgentPrivateTaskContextClient) IsOnboardingFirstTaskBundleTask(ctx context.Context, bearerToken, workspaceID, taskID string) (bool, error) {
	if c == nil {
		return false, errOnboardingTaskProvenanceUnavailable
	}
	bearerToken = strings.TrimSpace(bearerToken)
	taskID = strings.TrimSpace(taskID)
	workspaceID = strings.TrimSpace(workspaceID)
	if bearerToken == "" || taskID == "" || workspaceID == "" {
		return false, nil
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	var location struct {
		Component privateComponentWorkspaceResponse `json:"getComponentWorkspace"`
	}
	err := c.dashboardGraphQL(ctx, bearerToken, `query OnboardingTaskLocation($id: String!) { getComponentWorkspace(componentId: $id) { id componentType team { id workspace { id } } } }`, map[string]string{"id": taskID}, &location)
	if errors.Is(err, ErrAuthorizationForbidden) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	component := location.Component
	if component.ID != taskID || component.ComponentType != "task" || component.Team.ID == "" || component.Team.Workspace.ID != workspaceID {
		return false, nil
	}
	var answer struct {
		Matched *bool `json:"isOnboardingFirstTaskBundleTask"`
	}
	err = c.dashboardGraphQL(ctx, bearerToken, `query OnboardingFirstTaskBundleTask($teamId: String!, $taskId: String!) { isOnboardingFirstTaskBundleTask(teamId: $teamId, taskId: $taskId) }`, map[string]string{"teamId": component.Team.ID, "taskId": taskID}, &answer)
	if errors.Is(err, ErrAuthorizationForbidden) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if answer.Matched == nil {
		return false, errOnboardingTaskProvenanceUnavailable
	}
	return *answer.Matched, nil
}

// verifyOnboardingFirstTaskBundleTask runs before the completion store is touched.
// It returns errOnboardingTaskProvenance when the monolith says no and
// errOnboardingTaskProvenanceUnavailable when the answer is missing or failed.
func (s Server) verifyOnboardingFirstTaskBundleTask(ctx context.Context, bearerToken, workspaceID, taskID string) error {
	reader, ok := s.taskContext.(onboardingFirstTaskBundleReader)
	if !ok {
		return errOnboardingTaskProvenanceUnavailable
	}
	matched, err := reader.IsOnboardingFirstTaskBundleTask(ctx, bearerToken, workspaceID, taskID)
	if err != nil {
		return errors.Join(errOnboardingTaskProvenanceUnavailable, err)
	}
	if !matched {
		return errOnboardingTaskProvenance
	}
	return nil
}

// writeOnboardingTaskProvenanceError keeps the rejection code (400) separate from
// the monolith outage (502). Both refuse the completion.
func writeOnboardingTaskProvenanceError(w http.ResponseWriter, err error) {
	if errors.Is(err, errOnboardingTaskProvenance) {
		writeAIAgentClientError(w, errOnboardingTaskProvenance)
		return
	}
	writeError(w, http.StatusBadGateway, errOnboardingTaskProvenanceUnavailable.Error())
}
