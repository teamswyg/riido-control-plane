package deploypolicy

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"
)

const deployPolicyGoldenSHA256 = "b1d71f6b66a11396de97e2f0313b9c78f4fe737968aeeae1f6cc74033a138e7d"

func TestDeployPolicyBehaviorGolden(t *testing.T) {
	parsed := loadRuntimeCDOwnership(t)
	body, err := json.Marshal(deployPolicyBehavior{
		Manifest:                  parsed,
		RequiredHardeningTasks:    requiredHardeningTasks(),
		RequiredInfraPaths:        requiredInfraAwarenessPaths(),
		ExpectedSecretKeys:        expectedSecretKeys(),
		ExpectedRequiredVars:      expectedRequiredVars(),
		ExpectedOptionalVars:      expectedOptionalVars(),
		ExpectedNonCDRuntimeKeys:  expectedNonCDRuntimeKeys(),
		DeployRequiredPhrases:     deployWorkflowRequiredPhrases(),
		DeployRuntimePhrases:      deployRuntimeRequiredPhrases(),
		SmokeRequiredPhrases:      smokeWorkflowRequiredPhrases(),
		WorkflowForbiddenPhrases:  workflowForbiddenPhrases(),
		GeneratedWorkflowPhrases:  generatedClientWorkflowPhrases(),
		GeneratedDocPhrases:       generatedClientDocPhrases(),
		GeneratedMigrationPhrases: generatedClientMigrationPhrases(),
		LiveWorkflowPaths:         liveWorkflowPaths(),
	})
	if err != nil {
		t.Fatalf("marshal deploy policy behavior: %v", err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(body)); got != deployPolicyGoldenSHA256 {
		t.Fatalf("deploy policy SHA mismatch: got %s want %s\n%s", got, deployPolicyGoldenSHA256, body)
	}
}

type deployPolicyBehavior struct {
	Manifest                  runtimeCDOwnership `json:"manifest"`
	RequiredHardeningTasks    []string           `json:"required_hardening_tasks"`
	RequiredInfraPaths        []string           `json:"required_infra_paths"`
	ExpectedSecretKeys        []string           `json:"expected_secret_keys"`
	ExpectedRequiredVars      []string           `json:"expected_required_vars"`
	ExpectedOptionalVars      []string           `json:"expected_optional_vars"`
	ExpectedNonCDRuntimeKeys  []string           `json:"expected_non_cd_runtime_keys"`
	DeployRequiredPhrases     []string           `json:"deploy_required_phrases"`
	DeployRuntimePhrases      []string           `json:"deploy_runtime_phrases"`
	SmokeRequiredPhrases      []string           `json:"smoke_required_phrases"`
	WorkflowForbiddenPhrases  []string           `json:"workflow_forbidden_phrases"`
	GeneratedWorkflowPhrases  []string           `json:"generated_workflow_phrases"`
	GeneratedDocPhrases       []string           `json:"generated_doc_phrases"`
	GeneratedMigrationPhrases []string           `json:"generated_migration_phrases"`
	LiveWorkflowPaths         []string           `json:"live_workflow_paths"`
}
