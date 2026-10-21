package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

// errOnboardingTaskProvenance rejects a task that the monolith does not record as
// the viewer's own first-task bundle task. It maps to the default 400 response.
var errOnboardingTaskProvenance = errors.New("onboarding completion requires the first task bundle created for this user")

// errOnboardingTemplateMismatch rejects a template_key that differs from the bundle
// template recorded by the monolith. It maps to the default 400 response.
var errOnboardingTemplateMismatch = errors.New("onboarding completion template_key must match the first task bundle template")

// errOnboardingViewerMismatch rejects a Bearer JWT that belongs to another Riido
// account than the authorized agent principal. It maps to 403.
var errOnboardingViewerMismatch = errors.New("onboarding completion requires the agent owner's own Riido session")

// errOnboardingTaskProvenanceUnavailable means the monolith did not give a usable
// answer. The completion is refused (fail closed) and the handler responds 502.
var errOnboardingTaskProvenanceUnavailable = errors.New("onboarding task provenance could not be verified")

// onboardingFirstTaskBundleAnswer is one monolith answer. ViewerID is the Riido
// account behind the Bearer JWT. Template is empty when the task is not the viewer's bundle task.
type onboardingFirstTaskBundleAnswer struct {
	ViewerID string
	Template AIAgentOnboardingTaskTemplateKey
}

// onboardingFirstTaskBundleReader asks the monolith with the viewer's Riido Bearer JWT.
type onboardingFirstTaskBundleReader interface {
	OnboardingFirstTaskBundleTemplate(ctx context.Context, bearerToken, workspaceID, taskID string) (onboardingFirstTaskBundleAnswer, error)
}

const onboardingFirstTaskBundleAnswerQuery = `query OnboardingFirstTaskBundleTemplate($teamId: String!, $taskId: String!) { getMyAccountInfo { id } onboardingFirstTaskBundleTemplate(teamId: $teamId, taskId: $taskId) }`

// OnboardingFirstTaskBundleTemplate resolves the task's team through the monolith
// (one request), then makes one request that returns the viewer's account id and
// the bundle template for the task. A task outside the workspace or not visible to
// the viewer returns an empty answer and no error. Transport and protocol failures return an error.
func (c *AIAgentPrivateTaskContextClient) OnboardingFirstTaskBundleTemplate(ctx context.Context, bearerToken, workspaceID, taskID string) (onboardingFirstTaskBundleAnswer, error) {
	if c == nil {
		return onboardingFirstTaskBundleAnswer{}, errOnboardingTaskProvenanceUnavailable
	}
	bearerToken = strings.TrimSpace(bearerToken)
	taskID = strings.TrimSpace(taskID)
	workspaceID = strings.TrimSpace(workspaceID)
	if bearerToken == "" || taskID == "" || workspaceID == "" {
		return onboardingFirstTaskBundleAnswer{}, errOnboardingTaskProvenanceUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	var location struct {
		Component privateComponentWorkspaceResponse `json:"getComponentWorkspace"`
	}
	err := c.dashboardGraphQL(ctx, bearerToken, `query OnboardingTaskLocation($id: String!) { getComponentWorkspace(componentId: $id) { id componentType team { id workspace { id } } } }`, map[string]string{"id": taskID}, &location)
	if errors.Is(err, ErrAuthorizationForbidden) {
		return onboardingFirstTaskBundleAnswer{}, nil
	}
	if err != nil {
		return onboardingFirstTaskBundleAnswer{}, err
	}
	component := location.Component
	if component.ID != taskID || component.ComponentType != "task" || component.Team.ID == "" || component.Team.Workspace.ID != workspaceID {
		return onboardingFirstTaskBundleAnswer{}, nil
	}
	var answer struct {
		Viewer struct {
			ID string `json:"id"`
		} `json:"getMyAccountInfo"`
		Template *string `json:"onboardingFirstTaskBundleTemplate"`
	}
	err = c.dashboardGraphQL(ctx, bearerToken, onboardingFirstTaskBundleAnswerQuery, map[string]string{"teamId": component.Team.ID, "taskId": taskID}, &answer)
	if errors.Is(err, ErrAuthorizationForbidden) {
		return onboardingFirstTaskBundleAnswer{}, nil
	}
	if err != nil {
		return onboardingFirstTaskBundleAnswer{}, err
	}
	result := onboardingFirstTaskBundleAnswer{ViewerID: strings.TrimSpace(answer.Viewer.ID)}
	if answer.Template == nil {
		return result, nil
	}
	template := AIAgentOnboardingTaskTemplateKey(strings.TrimSpace(*answer.Template))
	if _, known := onboardingCompletionCopy[template]; !known {
		return onboardingFirstTaskBundleAnswer{}, errOnboardingTaskProvenanceUnavailable
	}
	result.Template = template
	return result, nil
}

// verifyOnboardingFirstTaskBundleTask runs before the completion store is touched.
// Order: not the viewer's bundle task (400), viewer is another account (403),
// template differs from the request (400). A missing viewer id is unavailable (502).
func (s Server) verifyOnboardingFirstTaskBundleTask(ctx context.Context, bearerToken, workspaceID, principalID, taskID string, requested AIAgentOnboardingTaskTemplateKey) error {
	reader, ok := s.taskContext.(onboardingFirstTaskBundleReader)
	if !ok {
		return errOnboardingTaskProvenanceUnavailable
	}
	answer, err := reader.OnboardingFirstTaskBundleTemplate(ctx, bearerToken, workspaceID, taskID)
	if err != nil {
		return errors.Join(errOnboardingTaskProvenanceUnavailable, err)
	}
	if answer.Template == "" {
		return errOnboardingTaskProvenance
	}
	if answer.ViewerID == "" {
		return errOnboardingTaskProvenanceUnavailable
	}
	if answer.ViewerID != strings.TrimSpace(principalID) {
		return errOnboardingViewerMismatch
	}
	if answer.Template != requested {
		return errOnboardingTemplateMismatch
	}
	return nil
}

// writeOnboardingTaskProvenanceError keeps the rejections apart: 403 for another
// account, 400 for a task or template that does not match, and 502 for a monolith
// outage. Every case refuses the completion.
func writeOnboardingTaskProvenanceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, errOnboardingViewerMismatch):
		writeError(w, http.StatusForbidden, "forbidden")
	case errors.Is(err, errOnboardingTaskProvenance), errors.Is(err, errOnboardingTemplateMismatch):
		writeAIAgentClientError(w, err)
	default:
		writeError(w, http.StatusBadGateway, errOnboardingTaskProvenanceUnavailable.Error())
	}
}
