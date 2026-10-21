package riidoaiserver

import (
	"context"
	"errors"

	"github.com/teamswyg/riido-contracts/assignment"
)

// A stored opt-in is not proof that a provider is still safe to dispatch now.
func (s *Store) recheckRecoveryCandidate(ctx context.Context, current Assignment, provider string) error {
	client, ok := s.agentRegistry.(AIAgentClientStore)
	if !ok {
		return errors.New("current provider recovery inventory is unavailable")
	}
	principal := AuthorizationResult{PrincipalID: current.CreatedBy, WorkspaceID: current.WorkspaceID}
	agents, err := client.ListAIAgentTaskAssignableAgents(ctx, principal, current.TaskID)
	if err != nil {
		return assignmentDependencyFailure(err)
	}
	selected, err := selectAssignableAgent(agents.Agents, current.AgentID)
	if err != nil {
		return err
	}
	binding, ok := lookupDaemonBinding(s.agentRegistry, current.AgentID)
	if !ok {
		return errors.New("current provider recovery binding is unavailable")
	}
	policy := assignment.CloneProviderRecoveryPolicy(current.RecoveryPolicy)
	if policy == nil {
		return errors.New("provider recovery candidate was not approved")
	}
	policy.Fallbacks = nil
	for _, candidate := range current.RecoveryPolicy.Fallbacks {
		if candidate.Provider == provider {
			policy.Fallbacks = append(policy.Fallbacks, candidate)
		}
	}
	if len(policy.Fallbacks) != 1 {
		return errors.New("provider recovery candidate was not approved")
	}
	_, err = (Server{aiAgent: client}).authorizeRecoveryPolicy(ctx, principal, selected, binding, current.AllowExperimentalRuntime, policy)
	return err
}
