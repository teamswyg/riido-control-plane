package riidoaiserver

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/teamswyg/riido-contracts/hostintegration"
)

const ReviewAccountSeedSchemaVersion = "riido-review-account-seed.v1"

//go:embed review_account_seed.riido.json
var reviewAccountSeedFS embed.FS

// ReviewAccountSeed is the store-review-only SaaS account seed. It contains no
// token, password, provider credential, or provider execution grant.
type ReviewAccountSeed struct {
	SchemaVersion           string                          `json:"schema_version"`
	PrincipalID             string                          `json:"principal_id"`
	DisplayName             string                          `json:"display_name"`
	Scopes                  []string                        `json:"scopes"`
	AgentCatalogRecords     []AgentCatalogRecord            `json:"agent_catalog_records"`
	SyntheticProviderStatus ReviewAccountProviderStatusSeed `json:"synthetic_provider_status"`
}

type ReviewAccountProviderStatusSeed struct {
	AgentID string                    `json:"agent_id"`
	Request ProviderStatusSyncRequest `json:"request"`
}

type ReviewAccountProvisionInput struct {
	TokenSHA256 string
}

type ReviewAccountProvisioning struct {
	Principal                      AgentCatalogPrincipal
	Credential                     StaticTokenCredential
	AgentCatalogRecords            []AgentCatalogRecord
	SyntheticProviderStatusAgentID string
	SyntheticProviderStatusRequest ProviderStatusSyncRequest
}

func LoadReviewAccountSeed() (ReviewAccountSeed, error) {
	data, err := reviewAccountSeedFS.ReadFile("review_account_seed.riido.json")
	if err != nil {
		return ReviewAccountSeed{}, err
	}
	var seed ReviewAccountSeed
	if err := json.Unmarshal(data, &seed); err != nil {
		return ReviewAccountSeed{}, fmt.Errorf("decode review account seed: %w", err)
	}
	if err := seed.Validate(); err != nil {
		return ReviewAccountSeed{}, err
	}
	return seed, nil
}

func (s ReviewAccountSeed) Validate() error {
	var errs []error
	s.PrincipalID = strings.TrimSpace(s.PrincipalID)
	s.DisplayName = strings.TrimSpace(s.DisplayName)
	if s.SchemaVersion != ReviewAccountSeedSchemaVersion {
		errs = append(errs, fmt.Errorf("unknown review account seed schema %q", s.SchemaVersion))
	}
	if s.PrincipalID == "" {
		errs = append(errs, errors.New("principal_id is required"))
	}
	if s.DisplayName == "" {
		errs = append(errs, errors.New("display_name is required"))
	}
	if _, err := normalizeAuthorizationScopes(s.Scopes); err != nil {
		errs = append(errs, fmt.Errorf("scopes: %w", err))
	}
	for _, scope := range s.Scopes {
		if forbiddenReviewAccountScope(scope) {
			errs = append(errs, fmt.Errorf("scope %s is forbidden for review account", scope))
		}
	}
	if len(s.AgentCatalogRecords) == 0 {
		errs = append(errs, errors.New("agent_catalog_records is required"))
	}
	hasOwnedPrivate := false
	hasOtherPublic := false
	for i, record := range s.AgentCatalogRecords {
		if err := record.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("agent_catalog_records[%d]: %w", i, err))
			continue
		}
		if record.OwnerPrincipalID == s.PrincipalID && record.Visibility == AgentCatalogVisibilityPrivate {
			hasOwnedPrivate = true
		}
		if record.OwnerPrincipalID != s.PrincipalID && record.Visibility == AgentCatalogVisibilityPublic {
			hasOtherPublic = true
		}
	}
	if !hasOwnedPrivate {
		errs = append(errs, errors.New("review account seed must include an owned private agent"))
	}
	if !hasOtherPublic {
		errs = append(errs, errors.New("review account seed must include another user's public agent"))
	}
	if strings.TrimSpace(s.SyntheticProviderStatus.AgentID) == "" {
		errs = append(errs, errors.New("synthetic_provider_status.agent_id is required"))
	} else if _, err := normalizeProviderStatusSync(s.SyntheticProviderStatus.AgentID, s.SyntheticProviderStatus.Request); err != nil {
		errs = append(errs, fmt.Errorf("synthetic_provider_status.request: %w", err))
	}
	for _, provider := range s.SyntheticProviderStatus.Request.Providers {
		if provider.RoutingStatus == hostintegration.ProviderRoutingAvailable {
			errs = append(errs, fmt.Errorf("synthetic provider %s must not be available", provider.ProviderKind))
		}
	}
	return errors.Join(errs...)
}

func ProvisionReviewAccount(seed ReviewAccountSeed, input ReviewAccountProvisionInput) (ReviewAccountProvisioning, error) {
	if err := seed.Validate(); err != nil {
		return ReviewAccountProvisioning{}, err
	}
	credential := StaticTokenCredential{
		PrincipalID: strings.TrimSpace(seed.PrincipalID),
		TokenSHA256: strings.ToLower(strings.TrimSpace(input.TokenSHA256)),
		Scopes:      append([]string(nil), seed.Scopes...),
	}
	if _, err := NewStaticTokenAuthorizer([]StaticTokenCredential{credential}); err != nil {
		return ReviewAccountProvisioning{}, err
	}
	status, err := normalizeProviderStatusSync(seed.SyntheticProviderStatus.AgentID, seed.SyntheticProviderStatus.Request)
	if err != nil {
		return ReviewAccountProvisioning{}, err
	}
	return ReviewAccountProvisioning{
		Principal: AgentCatalogPrincipal{
			PrincipalID: seed.PrincipalID,
		},
		Credential:                     credential,
		AgentCatalogRecords:            append([]AgentCatalogRecord(nil), seed.AgentCatalogRecords...),
		SyntheticProviderStatusAgentID: strings.TrimSpace(seed.SyntheticProviderStatus.AgentID),
		SyntheticProviderStatusRequest: status,
	}, nil
}

func (s *Store) ApplyReviewAccountProvisioning(ctx context.Context, provisioning ReviewAccountProvisioning) error {
	reply := make(chan applyReviewAccountProvisioningResult, 1)
	if err := s.send(ctx, applyReviewAccountProvisioningCmd{provisioning: provisioning, reply: reply}); err != nil {
		return err
	}
	select {
	case res := <-reply:
		return res.err
	case <-ctx.Done():
		return ctx.Err()
	}
}

type applyReviewAccountProvisioningCmd struct {
	provisioning ReviewAccountProvisioning
	reply        chan applyReviewAccountProvisioningResult
}

type applyReviewAccountProvisioningResult struct {
	err error
}

func (s *Store) handleApplyReviewAccountProvisioning(state *storeState, provisioning ReviewAccountProvisioning) error {
	if err := provisioning.Principal.Validate(); err != nil {
		return err
	}
	if _, err := NewStaticTokenAuthorizer([]StaticTokenCredential{provisioning.Credential}); err != nil {
		return err
	}
	for _, record := range provisioning.AgentCatalogRecords {
		record = normalizeAgentCatalogRecord(record)
		if err := record.Validate(); err != nil {
			return err
		}
		state.agentCatalog[record.AgentID] = record
	}
	req, err := normalizeProviderStatusSync(provisioning.SyntheticProviderStatusAgentID, provisioning.SyntheticProviderStatusRequest)
	if err != nil {
		return err
	}
	state.providerStatuses[provisioning.SyntheticProviderStatusAgentID] = ProviderStatusSyncResponse{
		SchemaVersion:       SchemaVersion,
		AgentID:             provisioning.SyntheticProviderStatusAgentID,
		DaemonID:            req.DaemonID,
		DeviceID:            req.DeviceID,
		RuntimeID:           req.RuntimeID,
		DistributionChannel: req.DistributionChannel,
		AppVersion:          req.AppVersion,
		Providers:           append([]ProviderStatusRecord(nil), req.Providers...),
		SyncedAt:            s.now(),
	}
	return nil
}

func forbiddenReviewAccountScope(scope string) bool {
	scope = strings.TrimSpace(scope)
	if scope == "riido:*" || scope == "agent:*" {
		return true
	}
	return strings.Contains(scope, ":poll") ||
		strings.Contains(scope, ":heartbeat") ||
		strings.Contains(scope, ":events:write") ||
		strings.Contains(scope, ":provider-status:write")
}
