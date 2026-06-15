package riidoaiserver

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

type AgentCatalogVisibility string

const (
	AgentCatalogVisibilityPublic  AgentCatalogVisibility = "public"
	AgentCatalogVisibilityPrivate AgentCatalogVisibility = "private"
)

type AgentCatalogAction string

const (
	AgentCatalogActionRead   AgentCatalogAction = "read"
	AgentCatalogActionUpdate AgentCatalogAction = "update"
	AgentCatalogActionDelete AgentCatalogAction = "delete"
)

type AgentCatalogRole string

const (
	AgentCatalogRoleAdmin AgentCatalogRole = "admin"
)

type AgentCatalogRecord struct {
	AgentID          string                 `json:"agent_id"`
	OwnerPrincipalID string                 `json:"owner_principal_id"`
	Visibility       AgentCatalogVisibility `json:"visibility"`
}

type AgentCatalogPrincipal struct {
	PrincipalID string             `json:"principal_id"`
	Roles       []AgentCatalogRole `json:"roles,omitempty"`
}

type AgentCatalogAccessDecisionReason string

const (
	AgentCatalogDecisionAdminRole          AgentCatalogAccessDecisionReason = "admin-role"
	AgentCatalogDecisionOwner              AgentCatalogAccessDecisionReason = "owner"
	AgentCatalogDecisionPublic             AgentCatalogAccessDecisionReason = "public"
	AgentCatalogDecisionInvalidAction      AgentCatalogAccessDecisionReason = "invalid-action"
	AgentCatalogDecisionInvalidAgentRecord AgentCatalogAccessDecisionReason = "invalid-agent-record"
	AgentCatalogDecisionInvalidPrincipal   AgentCatalogAccessDecisionReason = "invalid-principal"
	AgentCatalogDecisionMutationDenied     AgentCatalogAccessDecisionReason = "mutation-requires-admin-or-owner"
	AgentCatalogDecisionPrivateAgentDenied AgentCatalogAccessDecisionReason = "private-agent"
)

type AgentCatalogAccessDecision struct {
	Allowed bool                             `json:"allowed"`
	Reason  AgentCatalogAccessDecisionReason `json:"reason"`
}

var (
	ErrInvalidAgentCatalogRecord    = errors.New("riidoaiserver: invalid agent catalog record")
	ErrInvalidAgentCatalogPrincipal = errors.New("riidoaiserver: invalid agent catalog principal")
)

func AgentCatalogPrincipalFromAuthorization(result AuthorizationResult) AgentCatalogPrincipal {
	return AgentCatalogPrincipal{
		PrincipalID: result.PrincipalID,
		Roles:       append([]AgentCatalogRole(nil), result.Roles...),
	}
}

func EvaluateAgentCatalogAccess(principal AgentCatalogPrincipal, record AgentCatalogRecord, action AgentCatalogAction) AgentCatalogAccessDecision {
	principal = normalizeAgentCatalogPrincipal(principal)
	record = normalizeAgentCatalogRecord(record)
	if err := principal.Validate(); err != nil {
		return denyAgentCatalogAccess(AgentCatalogDecisionInvalidPrincipal)
	}
	if err := record.Validate(); err != nil {
		return denyAgentCatalogAccess(AgentCatalogDecisionInvalidAgentRecord)
	}
	if !isValidAgentCatalogAction(action) {
		return denyAgentCatalogAccess(AgentCatalogDecisionInvalidAction)
	}
	if principal.HasRole(AgentCatalogRoleAdmin) {
		return allowAgentCatalogAccess(AgentCatalogDecisionAdminRole)
	}
	if principal.PrincipalID == record.OwnerPrincipalID {
		return allowAgentCatalogAccess(AgentCatalogDecisionOwner)
	}
	if action == AgentCatalogActionRead && record.Visibility == AgentCatalogVisibilityPublic {
		return allowAgentCatalogAccess(AgentCatalogDecisionPublic)
	}
	if action == AgentCatalogActionRead {
		return denyAgentCatalogAccess(AgentCatalogDecisionPrivateAgentDenied)
	}
	return denyAgentCatalogAccess(AgentCatalogDecisionMutationDenied)
}

func VisibleAgentCatalogRecords(principal AgentCatalogPrincipal, records []AgentCatalogRecord) []AgentCatalogRecord {
	visible := make([]AgentCatalogRecord, 0, len(records))
	for _, record := range records {
		normalized := normalizeAgentCatalogRecord(record)
		decision := EvaluateAgentCatalogAccess(principal, normalized, AgentCatalogActionRead)
		if decision.Allowed {
			visible = append(visible, normalized)
		}
	}
	return visible
}

func (p AgentCatalogPrincipal) Validate() error {
	p = normalizeAgentCatalogPrincipal(p)
	if p.PrincipalID == "" {
		return fmt.Errorf("%w: principal_id is required", ErrInvalidAgentCatalogPrincipal)
	}
	if _, err := normalizeAgentCatalogRoles(p.Roles); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidAgentCatalogPrincipal, err)
	}
	return nil
}

func (p AgentCatalogPrincipal) HasRole(role AgentCatalogRole) bool {
	roles, err := normalizeAgentCatalogRoles(p.Roles)
	if err != nil {
		return false
	}
	return slices.Contains(roles, role)
}

func (r AgentCatalogRecord) Validate() error {
	r = normalizeAgentCatalogRecord(r)
	if r.AgentID == "" {
		return fmt.Errorf("%w: agent_id is required", ErrInvalidAgentCatalogRecord)
	}
	if r.OwnerPrincipalID == "" {
		return fmt.Errorf("%w: owner_principal_id is required", ErrInvalidAgentCatalogRecord)
	}
	if !isValidAgentCatalogVisibility(r.Visibility) {
		return fmt.Errorf("%w: visibility must be public or private", ErrInvalidAgentCatalogRecord)
	}
	return nil
}

func normalizeAgentCatalogPrincipal(principal AgentCatalogPrincipal) AgentCatalogPrincipal {
	principal.PrincipalID = strings.TrimSpace(principal.PrincipalID)
	roles, err := normalizeAgentCatalogRoles(principal.Roles)
	if err == nil {
		principal.Roles = roles
	}
	return principal
}

func normalizeAgentCatalogRecord(record AgentCatalogRecord) AgentCatalogRecord {
	record.AgentID = strings.TrimSpace(record.AgentID)
	record.OwnerPrincipalID = strings.TrimSpace(record.OwnerPrincipalID)
	record.Visibility = AgentCatalogVisibility(strings.TrimSpace(string(record.Visibility)))
	return record
}

func normalizeAgentCatalogRoles(roles []AgentCatalogRole) ([]AgentCatalogRole, error) {
	normalized := make([]AgentCatalogRole, 0, len(roles))
	seen := map[AgentCatalogRole]struct{}{}
	for _, role := range roles {
		role = AgentCatalogRole(strings.TrimSpace(string(role)))
		if role == "" {
			return nil, errors.New("role must not be empty")
		}
		if role != AgentCatalogRoleAdmin {
			return nil, fmt.Errorf("unsupported role %s", role)
		}
		if _, exists := seen[role]; exists {
			return nil, fmt.Errorf("duplicate role %s", role)
		}
		seen[role] = struct{}{}
		normalized = append(normalized, role)
	}
	return normalized, nil
}

func isValidAgentCatalogVisibility(visibility AgentCatalogVisibility) bool {
	return visibility == AgentCatalogVisibilityPublic || visibility == AgentCatalogVisibilityPrivate
}

func isValidAgentCatalogAction(action AgentCatalogAction) bool {
	return action == AgentCatalogActionRead || action == AgentCatalogActionUpdate || action == AgentCatalogActionDelete
}

func allowAgentCatalogAccess(reason AgentCatalogAccessDecisionReason) AgentCatalogAccessDecision {
	return AgentCatalogAccessDecision{Allowed: true, Reason: reason}
}

func denyAgentCatalogAccess(reason AgentCatalogAccessDecisionReason) AgentCatalogAccessDecision {
	return AgentCatalogAccessDecision{Allowed: false, Reason: reason}
}
