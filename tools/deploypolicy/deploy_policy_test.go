package deploypolicy

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestDeployAIAgentTestnetPublicRedactionPolicy(t *testing.T) {
	workflow := mustRead(t, "../../.github/workflows/deploy-ai-agent-testnet.yml")
	smokeWorkflow := mustRead(t, "../../.github/workflows/ai-agent-client-testnet-smoke.yml")
	readme := mustRead(t, "../../README.md")
	boundary := mustRead(t, "../../docs/30-architecture/runtime-deployment-boundary.md")
	domain := mustRead(t, "../../docs/20-domain/saas-control-plane.md")
	clientAPI := mustRead(t, "../../docs/20-domain/ai-agent-client-api.md")
	clientDelivery := mustRead(t, "../../docs/30-architecture/api-client-delivery.md")
	migration := mustRead(t, "../../docs/migration/control-plane.md")
	generator := mustRead(t, "../../tools/reactquerygen/main.go")
	generatedClient := mustRead(t, "../../web/generated/aiAgentClient.ts")
	generatedReactClient := mustRead(t, "../../web/generated/aiAgentClient.react.ts")

	requireContains(t, workflow, "echo \"::add-mask::$aws_account_id\"")
	requireContains(t, workflow, "echo \"::add-mask::$registry\"")
	requireContains(t, workflow, "echo \"::add-mask::$image_uri\"")
	requireContains(t, workflow, "echo \"::add-mask::$current_task_definition\"")
	requireContains(t, workflow, "echo \"::add-mask::$next_task_definition\"")
	for _, masked := range []string{
		"AWS_REGION",
		"ECR_REPOSITORY",
		"ECS_CLUSTER",
		"ECS_SERVICE",
		"ECS_CONTAINER_NAME",
		"CODEDEPLOY_APPLICATION",
		"CODEDEPLOY_DEPLOYMENT_GROUP",
		"TESTNET_BASE_URL",
		"TESTNET_WORKSPACE_ID",
	} {
		requireContains(t, workflow, masked)
	}
	requireContains(t, workflow, "redacted=(")
	requireContains(t, workflow, "echo \"::add-mask::${!name}\"")
	requireContains(t, workflow, "if [ \"$image_tag\" = \"latest\" ]")
	requireContains(t, workflow, "workflow_dispatch")
	requireContains(t, workflow, "tags:")
	requireContains(t, workflow, "- \"v*\"")
	requireContains(t, workflow, "TESTNET_BASE_URL: ${{ vars.RIIDO_AI_SERVER_TESTNET_BASE_URL }}")
	requireContains(t, workflow, "CODEDEPLOY_APPLICATION: ${{ vars.RIIDO_AI_SERVER_CODEDEPLOY_APPLICATION }}")
	requireContains(t, workflow, "CODEDEPLOY_DEPLOYMENT_GROUP: ${{ vars.RIIDO_AI_SERVER_CODEDEPLOY_DEPLOYMENT_GROUP }}")
	requireContains(t, workflow, "printf '%s' \"$image_uri\" > \"$RUNNER_TEMP/riido-image-uri\"")
	requireContains(t, workflow, "printf '%s' \"$next_task_definition\" > \"$RUNNER_TEMP/riido-task-definition-arn\"")
	requireContains(t, workflow, "printf '%s' \"$container_port\" > \"$RUNNER_TEMP/riido-container-port\"")
	requireContains(t, workflow, "umask 077")
	requireContains(t, workflow, "chmod 600 \"$current_json\"")
	requireContains(t, workflow, "chmod 600 \"$next_json\"")
	requireContains(t, workflow, "Cleanup live handoff files")
	requireContains(t, workflow, "if: always()")
	requireContains(t, workflow, "rm -f \\")
	requireContains(t, workflow, "\"$RUNNER_TEMP/riido-image-uri\"")
	requireContains(t, workflow, "\"$RUNNER_TEMP/riido-task-definition-arn\"")
	requireContains(t, workflow, "\"$RUNNER_TEMP/riido-container-port\"")
	requireContains(t, workflow, "current_json=\"$RUNNER_TEMP/task-definition.current.json\"")
	requireContains(t, workflow, "next_json=\"$RUNNER_TEMP/task-definition.next.json\"")
	requireContains(t, workflow, "appspec_json=\"$RUNNER_TEMP/codedeploy-appspec.json\"")
	requireContains(t, workflow, "deployment_json=\"$RUNNER_TEMP/codedeploy-deployment.json\"")
	requireContains(t, workflow, "revisionType: \"AppSpecContent\"")
	requireContains(t, workflow, "aws deploy create-deployment")
	requireContains(t, workflow, "wait_deployment_id=\"$(cat \"$deployment_id_file\")\"")
	requireContains(t, workflow, "echo \"::add-mask::$wait_deployment_id\"")
	requireContains(t, workflow, "aws deploy wait deployment-successful")
	requireContains(t, smokeWorkflow, "TESTNET_BASE_URL: ${{ vars.RIIDO_AI_SERVER_TESTNET_BASE_URL }}")
	requireContains(t, smokeWorkflow, "echo \"::add-mask::$TESTNET_BASE_URL\"")
	requireContains(t, smokeWorkflow, "echo \"::add-mask::$TESTNET_TOKEN\"")
	requireContains(t, smokeWorkflow, "umask 077")
	requireContains(t, smokeWorkflow, "trap 'rm -f \"$replay\"' EXIT")
	requireContains(t, clientAPI, "not from a manual")
	requireContains(t, clientAPI, "The workflow masks both values")

	for _, forbidden := range []string{
		"actions/upload-artifact",
		"inputs.base_url",
		"description: \"Optional AI Agent testnet base URL",
		"GITHUB_OUTPUT",
		"latest\" >>",
		"latest' >>",
		"task-definition.next.json\" >>",
		"task-definition.current.json\" >>",
		"appspec.json\" >>",
		"deployment-id\" >>",
	} {
		if strings.Contains(workflow, forbidden) {
			t.Fatalf("deploy workflow must not contain %q", forbidden)
		}
		if strings.Contains(smokeWorkflow, forbidden) {
			t.Fatalf("smoke workflow must not contain %q", forbidden)
		}
	}

	requireContains(t, readme, "live URL, AWS account id, ARN, image digest")
	requireContains(t, boundary, "task-definition ARNs, image digests, live workflow run URLs")
	requireContains(t, boundary, "must not upload")
	requireContains(t, boundary, "deployment artifacts from the live")
	requireContains(t, boundary, "CodeDeploy Handoff")
	requireContains(t, boundary, "runtime artifact CD execution still belongs")
	requireContains(t, domain, "live URLs, task-definition ARNs")
	requireContains(t, domain, "CodeDeploy blue/green")
	requireContains(t, migration, "RIID-4812 tightens that public boundary")
	requireContains(t, migration, "RIID-4814")
	requireContains(t, migration, "RIID-4815")
	requireContains(t, migration, "RIID-4822")
	requireContains(t, migration, "RIID-4825")
	requireContains(t, migration, "RIID-4835")

	for path, body := range map[string]string{
		"README.md":                            readme,
		"runtime-deployment-boundary.md":       boundary,
		"saas-control-plane.md":                domain,
		"ai-agent-client-api.md":               clientAPI,
		"api-client-delivery.md":               clientDelivery,
		"control-plane.md":                     migration,
		"tools/reactquerygen/main.go":          generator,
		"web/generated/aiAgentClient.ts":       generatedClient,
		"web/generated/aiAgentClient.react.ts": generatedReactClient,
	} {
		if strings.Contains(body, "ai-api.riido.io") {
			t.Fatalf("%s must not pin the live testnet host", path)
		}
	}
}

func TestRuntimeCDOwnershipManifest(t *testing.T) {
	manifest := mustRead(t, "../../docs/30-architecture/runtime-cd-ownership.riido.json")
	doc := mustRead(t, "../../docs/30-architecture/runtime-cd-ownership.md")
	boundary := mustRead(t, "../../docs/30-architecture/runtime-deployment-boundary.md")
	integration := mustRead(t, "../../docs/30-architecture/integration-matrix.md")
	readme := mustRead(t, "../../README.md")
	domain := mustRead(t, "../../docs/20-domain/saas-control-plane.md")
	migration := mustRead(t, "../../docs/migration/control-plane.md")
	deployWorkflow := mustRead(t, "../../.github/workflows/deploy-ai-agent-testnet.yml")
	smokeWorkflow := mustRead(t, "../../.github/workflows/ai-agent-client-testnet-smoke.yml")

	var parsed struct {
		SchemaVersion string   `json:"schema_version"`
		ID            string   `json:"id"`
		RiidoTask     string   `json:"riido_task"`
		Hardening     []string `json:"hardening_tasks"`
		Supersedes    []string `json:"supersedes_tasks"`
		Runtime       string   `json:"runtime_service"`
		Current       struct {
			ID            string   `json:"id"`
			Status        string   `json:"status"`
			CDOwner       string   `json:"cd_owner"`
			TopologyOwner string   `json:"topology_owner"`
			Workflow      string   `json:"workflow"`
			Allowed       []string `json:"allowed_actions"`
		} `json:"current_strategy"`
		OptionalModes []struct {
			ID               string   `json:"id"`
			Status           string   `json:"status"`
			CDOwner          string   `json:"cd_owner"`
			TopologyOwner    string   `json:"topology_owner"`
			Workflow         string   `json:"workflow"`
			ActivationInputs []string `json:"activation_inputs"`
			Allowed          []string `json:"allowed_actions"`
			MustNotOwn       []string `json:"must_not_own"`
		} `json:"optional_workflow_modes"`
		Future []struct {
			ID                 string   `json:"id"`
			Status             string   `json:"status"`
			CDOwner            string   `json:"cd_owner"`
			TopologyOwner      string   `json:"topology_owner"`
			ControlPlaneMayOwn []string `json:"control_plane_may_own"`
			InfraMustOwn       []string `json:"infra_must_own"`
		} `json:"future_strategies"`
		Redaction struct {
			Workflows      []string `json:"public_repo_workflows"`
			MayCommit      []string `json:"public_repo_may_commit"`
			ShouldMinimize []string `json:"public_repo_should_minimize"`
			MustNot        []string `json:"public_repo_must_not_commit_or_upload"`
		} `json:"public_redaction_policy"`
		Handoff struct {
			Scope              string   `json:"scope"`
			AllowedStorage     []string `json:"allowed_storage"`
			RequiredCleanup    []string `json:"required_cleanup"`
			ForbiddenMechanism []string `json:"forbidden_mechanisms"`
		} `json:"live_handoff_policy"`
		Infra struct {
			Repo       string   `json:"repo"`
			Paths      []string `json:"paths"`
			LocalScope string   `json:"local_scope"`
		} `json:"infra_consumes"`
		InfraVisibility struct {
			Repo        string   `json:"repo"`
			MustKnow    []string `json:"must_know"`
			MustNotFrom []string `json:"must_not_receive_from_public_workflow"`
		} `json:"infra_visibility_policy"`
		PublicExport struct {
			RiidoTask              string   `json:"riido_task"`
			CanonicalOwner         string   `json:"canonical_owner"`
			InfraAwarenessOwner    string   `json:"infra_awareness_owner"`
			AllowedPublicExports   []string `json:"allowed_public_exports"`
			ForbiddenPublicExports []string `json:"forbidden_public_exports"`
			InfraMustConsumeOnly   []string `json:"infra_must_consume_only"`
			WorkflowMustNotUse     []string `json:"workflow_must_not_use"`
		} `json:"public_export_contract"`
		PublicSurfaceScan struct {
			RiidoTask                  string   `json:"riido_task"`
			CanonicalOwner             string   `json:"canonical_owner"`
			InfraAwarenessOwner        string   `json:"infra_awareness_owner"`
			ScopePaths                 []string `json:"scope_paths"`
			ForbiddenLiterals          []string `json:"forbidden_literals"`
			ForbiddenRegexes           []string `json:"forbidden_regexes"`
			WorkflowForbiddenMechanism []string `json:"workflow_forbidden_mechanisms"`
			AllowedPublicSurface       []string `json:"allowed_public_surface"`
			InfraMustTreatScanAs       string   `json:"infra_must_treat_scan_as"`
		} `json:"public_surface_scan_contract"`
		PublicConfigKeyMinimization struct {
			RiidoTask              string   `json:"riido_task"`
			CanonicalOwner         string   `json:"canonical_owner"`
			InfraAwarenessOwner    string   `json:"infra_awareness_owner"`
			Rule                   string   `json:"rule"`
			RequiredSecretKeys     []string `json:"required_secret_keys"`
			RequiredVariableKeys   []string `json:"required_variable_keys"`
			OptionalVariableKeys   []string `json:"optional_variable_keys"`
			StableInfraSourceNames []string `json:"stable_infra_source_names"`
			PublicDocsMayReference []string `json:"public_docs_may_reference"`
			PublicDocsMustNotRef   []string `json:"public_docs_must_not_reference"`
			WorkflowKeySource      string   `json:"workflow_key_source"`
		} `json:"public_config_key_minimization"`
		PublicSensitiveSurfaceGuard struct {
			RiidoTask                  string   `json:"riido_task"`
			CanonicalOwner             string   `json:"canonical_owner"`
			InfraAwarenessOwner        string   `json:"infra_awareness_owner"`
			Rule                       string   `json:"rule"`
			PublicKeyNamesAreSensitive bool     `json:"public_key_names_are_sensitive"`
			KeyNameScopePaths          []string `json:"key_name_scope_paths"`
			CanonicalCDKeyListPaths    []string `json:"canonical_cd_key_list_paths"`
			BroadSummaryDocsMustLink   []string `json:"broad_summary_docs_must_link_not_list_cd_keys"`
			AllowedPublicInformation   []string `json:"allowed_public_information"`
			AllowedPublicNonCDRuntime  []struct {
				Name   string `json:"name"`
				Reason string `json:"reason"`
			} `json:"allowed_public_non_cd_runtime_keys"`
			ForbiddenPublicInformation []string `json:"forbidden_public_information"`
			InfraMustKnow              []string `json:"infra_must_know"`
		} `json:"public_sensitive_surface_guard"`
		PublicOperationalDetailMinimization struct {
			RiidoTask             string   `json:"riido_task"`
			CanonicalOwner        string   `json:"canonical_owner"`
			InfraAwarenessOwner   string   `json:"infra_awareness_owner"`
			Rule                  string   `json:"rule"`
			PublicRepoMayKeep     []string `json:"public_repo_may_keep"`
			PublicRepoShouldAvoid []string `json:"public_repo_should_avoid"`
			InfraMustKnow         []string `json:"infra_must_know"`
		} `json:"public_operational_detail_minimization"`
		CodeDeployActivationGate struct {
			RiidoTask              string   `json:"riido_task"`
			CanonicalOwner         string   `json:"canonical_owner"`
			InfraAwarenessOwner    string   `json:"infra_awareness_owner"`
			Status                 string   `json:"status"`
			Rule                   string   `json:"rule"`
			ActivationRequirements []string `json:"activation_requirements"`
			PublicRepoMayKeep      []string `json:"public_repo_may_keep"`
			PublicRepoMustNotKeep  []string `json:"public_repo_must_not_keep"`
			InfraMustKnow          []string `json:"infra_must_know"`
		} `json:"codedeploy_activation_gate"`
		InfraTopology struct {
			RiidoTask      string   `json:"riido_task"`
			Repo           string   `json:"repo"`
			WorkUnit       string   `json:"terraform_work_unit"`
			RequiredOutput []string `json:"required_outputs"`
			MustNotConsume []string `json:"control_plane_must_not_consume"`
		} `json:"infra_topology_contract"`
		DependencyDirection struct {
			TopDown  string `json:"top_down"`
			BottomUp string `json:"bottom_up"`
		} `json:"dependency_direction"`
	}
	decodeStrictJSONDocument(t, "runtime CD ownership manifest", manifest, &parsed)

	if parsed.SchemaVersion != "riido-control-plane-runtime-cd-ownership.v1" {
		t.Fatalf("unexpected schema version: %q", parsed.SchemaVersion)
	}
	if parsed.ID != "runtime-cd-ownership" || parsed.RiidoTask != "RIID-4825" || parsed.Runtime != "riido_ai_server" {
		t.Fatalf("manifest identity drifted: %#v", parsed)
	}
	requireSliceContains(t, parsed.Hardening, "RIID-4833")
	requireSliceContains(t, parsed.Hardening, "RIID-4835")
	requireSliceContains(t, parsed.Hardening, "RIID-4836")
	requireSliceContains(t, parsed.Hardening, "RIID-4837")
	requireSliceContains(t, parsed.Hardening, "RIID-4839")
	requireSliceContains(t, parsed.Hardening, "RIID-4842")
	requireSliceContains(t, parsed.Hardening, "RIID-4844")
	requireSliceContains(t, parsed.Hardening, "RIID-4845")
	requireSliceContains(t, parsed.Hardening, "RIID-4853")
	requireSliceContains(t, parsed.Hardening, "RIID-4855")
	if parsed.Current.CDOwner != "riido-control-plane" || parsed.Current.TopologyOwner != "riido-infra" {
		t.Fatalf("current CD ownership drifted: %#v", parsed.Current)
	}
	if parsed.Current.Workflow != ".github/workflows/deploy-ai-agent-testnet.yml" {
		t.Fatalf("workflow drifted: %q", parsed.Current.Workflow)
	}
	if len(parsed.Current.Allowed) < 5 {
		t.Fatalf("current CD allowed actions are underspecified: %#v", parsed.Current.Allowed)
	}

	var optionalCodeDeployFound bool
	for _, mode := range parsed.OptionalModes {
		if mode.ID == "codedeploy-blue-green" {
			optionalCodeDeployFound = true
			if mode.CDOwner != "riido-control-plane" || mode.TopologyOwner != "riido-infra" {
				t.Fatalf("optional CodeDeploy owner drifted: %#v", mode)
			}
			requireSliceContains(t, mode.ActivationInputs, "RIIDO_AI_SERVER_CODEDEPLOY_APPLICATION")
			requireSliceContains(t, mode.ActivationInputs, "RIIDO_AI_SERVER_CODEDEPLOY_DEPLOYMENT_GROUP")
			requireSliceContains(t, mode.Allowed, "create CodeDeploy AppSpec content in same-job runner temp files")
			requireSliceContains(t, mode.Allowed, "wait for CodeDeploy deployment success")
			requireSliceContains(t, mode.MustNotOwn, "CodeDeploy application or deployment group topology")
		}
	}
	if !optionalCodeDeployFound {
		t.Fatal("optional CodeDeploy workflow mode is missing")
	}

	var codeDeployFound bool
	for _, future := range parsed.Future {
		if future.ID == "codedeploy-blue-green" {
			codeDeployFound = true
			if future.CDOwner != "riido-control-plane" || future.TopologyOwner != "riido-infra" {
				t.Fatalf("CodeDeploy owner drifted: %#v", future)
			}
			requireSliceContains(t, future.ControlPlaneMayOwn, "create CodeDeploy deployment from the same-job immutable image value and infra-provided deployment target")
			requireSliceContains(t, future.ControlPlaneMayOwn, "wait for CodeDeploy deployment completion")
			requireSliceContains(t, future.InfraMustOwn, "CodeDeploy application and deployment group")
			requireSliceContains(t, future.InfraMustOwn, "blue green target groups and listener topology")
		}
	}
	if !codeDeployFound {
		t.Fatal("future CodeDeploy strategy is missing")
	}

	for _, forbidden := range []string{
		"live URL values",
		"AWS account IDs",
		"ARNs",
		"CodeDeploy deployment IDs",
		"image digests or image URIs",
		"workflow_dispatch input values carrying live URLs",
		"GitHub step outputs carrying live deployment values",
		"task definition JSON",
		"CodeDeploy AppSpec JSON",
		"smoke response payloads",
		"Terraform state",
	} {
		requireSliceContains(t, parsed.Redaction.MustNot, forbidden)
	}
	requireSliceContains(t, parsed.Redaction.Workflows, ".github/workflows/deploy-ai-agent-testnet.yml")
	requireSliceContains(t, parsed.Redaction.Workflows, ".github/workflows/ai-agent-client-testnet-smoke.yml")
	requireSliceContains(t, parsed.Redaction.MayCommit, "workflow file")
	requireSliceContains(t, parsed.Redaction.MayCommit, "non-live behavior documentation")
	requireSliceContains(t, parsed.Redaction.ShouldMinimize, "publish only stable configuration key names that operators must set")
	requireSliceContains(t, parsed.Redaction.ShouldMinimize, "centralize required deploy key names in the workflow files and machine-readable ownership manifest instead of scattering environment-specific examples")
	requireSliceContains(t, parsed.Redaction.ShouldMinimize, "keep exact deploy key-name lists out of human-readable public docs except the workflow files that consume them")
	requireSliceContains(t, parsed.Redaction.ShouldMinimize, "avoid environment-specific examples for domains, clusters, services, applications, deployment groups, ARNs, and URLs")
	requireSliceContains(t, parsed.Redaction.ShouldMinimize, "apply restrictive file permissions before writing live handoff, task-definition, CodeDeploy, or smoke replay files")
	if parsed.Handoff.Scope != "same-job-runner-temp-only" {
		t.Fatalf("handoff scope drifted: %#v", parsed.Handoff)
	}
	requireSliceContains(t, parsed.Handoff.AllowedStorage, "$RUNNER_TEMP files created under umask 077 and chmod 600 before reuse")
	requireSliceContains(t, parsed.Handoff.RequiredCleanup, "remove image URI, ECS task-definition ARN, and container-port temp files in an always-running cleanup step")
	requireSliceContains(t, parsed.Handoff.RequiredCleanup, "remove generated CodeDeploy AppSpec, request JSON, and deployment-id files with same-step traps")
	requireSliceContains(t, parsed.Handoff.RequiredCleanup, "remove smoke replay temp files with same-step traps")
	requireSliceContains(t, parsed.Handoff.ForbiddenMechanism, "GitHub step outputs for live deployment values")
	requireSliceContains(t, parsed.Handoff.ForbiddenMechanism, "uploaded workflow artifacts")
	if parsed.Infra.Repo != "riido-infra" {
		t.Fatalf("infra consumer repo drifted: %q", parsed.Infra.Repo)
	}
	if parsed.InfraTopology.RiidoTask != "RIID-4822" || parsed.InfraTopology.Repo != "riido-infra" {
		t.Fatalf("infra topology contract drifted: %#v", parsed.InfraTopology)
	}
	requireSliceContains(t, parsed.InfraTopology.RequiredOutput, "codedeploy_application_name")
	requireSliceContains(t, parsed.InfraTopology.RequiredOutput, "codedeploy_deployment_group_name")
	requireSliceContains(t, parsed.InfraTopology.MustNotConsume, "CodeDeploy service role ARN")
	requireSliceContains(t, parsed.InfraTopology.MustNotConsume, "target group ARN")
	requireSliceContains(t, parsed.Infra.Paths, "docs/architecture/terraform-authoring.md")
	requireSliceContains(t, parsed.Infra.Paths, "deploy/work-units/riid-4825-control-plane-cd-ownership-remodel.riido.json")
	requireSliceContains(t, parsed.Infra.Paths, "deploy/work-units/riid-4833-control-plane-cd-public-redaction-hardening.riido.json")
	requireSliceContains(t, parsed.Infra.Paths, "deploy/work-units/riid-4835-control-plane-cd-public-export-contract.riido.json")
	requireSliceContains(t, parsed.Infra.Paths, "deploy/work-units/riid-4836-control-plane-cd-public-surface-redaction-scan.riido.json")
	requireSliceContains(t, parsed.Infra.Paths, "deploy/work-units/riid-4837-cd-ownership-final-guard-public-surface-minimization.riido.json")
	requireSliceContains(t, parsed.Infra.Paths, "deploy/work-units/riid-4839-cd-public-config-key-minimization.riido.json")
	requireSliceContains(t, parsed.Infra.Paths, "deploy/work-units/riid-4854-control-plane-cd-public-minimization-awareness-no-diff.riido.json")
	requireSliceContains(t, parsed.Infra.Paths, "deploy/work-units/riid-4856-codedeploy-activation-gate-awareness-no-diff.riido.json")
	requireSliceContains(t, parsed.Infra.Paths, "deploy/work-units/riid-4860-control-plane-cd-ownership-awareness-guard.riido.json")
	if parsed.InfraVisibility.Repo != "riido-infra" {
		t.Fatalf("infra visibility repo drifted: %q", parsed.InfraVisibility.Repo)
	}
	requireSliceContains(t, parsed.InfraVisibility.MustKnow, "riido-control-plane owns runtime artifact CD execution")
	requireSliceContains(t, parsed.InfraVisibility.MustKnow, "infra-local awareness guards verify the no-diff boundary without consuming public workflow live payloads")
	requireSliceContains(t, parsed.InfraVisibility.MustNotFrom, "generated CodeDeploy AppSpec JSON")
	requireSliceContains(t, parsed.InfraVisibility.MustNotFrom, "image digests or image URIs")
	requireSliceContains(t, parsed.InfraVisibility.MustNotFrom, "smoke replay temp files")
	if parsed.PublicExport.RiidoTask != "RIID-4835" {
		t.Fatalf("public export work unit drifted: %#v", parsed.PublicExport)
	}
	if parsed.PublicExport.CanonicalOwner != "riido-control-plane" || parsed.PublicExport.InfraAwarenessOwner != "riido-infra" {
		t.Fatalf("public export ownership drifted: %#v", parsed.PublicExport)
	}
	requireSliceContains(t, parsed.PublicExport.AllowedPublicExports, "stable GitHub secret and variable key names")
	requireSliceContains(t, parsed.PublicExport.AllowedPublicExports, "stable infra output key names that operators map into GitHub environment variables")
	requireSliceContains(t, parsed.PublicExport.AllowedPublicExports, "aggregate pass or fail status without live payload values")
	requireSliceContains(t, parsed.PublicExport.ForbiddenPublicExports, "image URIs or digests")
	requireSliceContains(t, parsed.PublicExport.ForbiddenPublicExports, "ECS task-definition JSON")
	requireSliceContains(t, parsed.PublicExport.ForbiddenPublicExports, "CodeDeploy create-deployment request JSON")
	requireSliceContains(t, parsed.PublicExport.ForbiddenPublicExports, "Terraform plan output, state, tfvars, apply logs, or raw operator evidence")
	requireSliceContains(t, parsed.PublicExport.InfraMustConsumeOnly, "stable output names")
	requireSliceContains(t, parsed.PublicExport.InfraMustConsumeOnly, "redaction categories")
	requireSliceContains(t, parsed.PublicExport.InfraMustConsumeOnly, "operator evidence summaries stored outside public repositories")
	requireSliceContains(t, parsed.PublicExport.WorkflowMustNotUse, "actions/upload-artifact for live deployment payloads")
	requireSliceContains(t, parsed.PublicExport.WorkflowMustNotUse, "GITHUB_OUTPUT for live deployment values")
	requireSliceContains(t, parsed.PublicExport.WorkflowMustNotUse, "workflow_dispatch inputs for live URLs")
	if parsed.PublicSurfaceScan.RiidoTask != "RIID-4836" {
		t.Fatalf("public surface scan work unit drifted: %#v", parsed.PublicSurfaceScan)
	}
	if parsed.PublicSurfaceScan.CanonicalOwner != "riido-control-plane" || parsed.PublicSurfaceScan.InfraAwarenessOwner != "riido-infra" {
		t.Fatalf("public surface scan ownership drifted: %#v", parsed.PublicSurfaceScan)
	}
	requireSliceContains(t, parsed.PublicSurfaceScan.ScopePaths, ".github/workflows/deploy-ai-agent-testnet.yml")
	requireSliceContains(t, parsed.PublicSurfaceScan.ScopePaths, "docs/30-architecture/api-client-delivery.md")
	requireSliceContains(t, parsed.PublicSurfaceScan.ScopePaths, "docs/30-architecture/runtime-cd-ownership.md")
	requireSliceContains(t, parsed.PublicSurfaceScan.ScopePaths, "web/generated/aiAgentClient.react.ts")
	requireSliceContains(t, parsed.PublicSurfaceScan.ForbiddenLiterals, "ai-api.riido.io")
	requireSliceContains(t, parsed.PublicSurfaceScan.WorkflowForbiddenMechanism, "GITHUB_OUTPUT")
	requireSliceContains(t, parsed.PublicSurfaceScan.AllowedPublicSurface, "AWS CLI response field names inside the deploy workflow")
	requireContains(t, parsed.PublicSurfaceScan.InfraMustTreatScanAs, "awareness policy only")
	if parsed.PublicConfigKeyMinimization.RiidoTask != "RIID-4839" {
		t.Fatalf("public config key minimization work unit drifted: %#v", parsed.PublicConfigKeyMinimization)
	}
	if parsed.PublicConfigKeyMinimization.CanonicalOwner != "riido-control-plane" ||
		parsed.PublicConfigKeyMinimization.InfraAwarenessOwner != "riido-infra" {
		t.Fatalf("public config key minimization ownership drifted: %#v", parsed.PublicConfigKeyMinimization)
	}
	requireContains(t, parsed.PublicConfigKeyMinimization.Rule, "minimum stable GitHub configuration keys")
	requireContains(t, parsed.PublicConfigKeyMinimization.WorkflowKeySource, "any additional RIIDO_AI_SERVER_*")
	expectedSecrets := []string{
		"RIIDO_AI_SERVER_DEPLOY_ROLE_ARN",
		"RIIDO_AI_SERVER_TESTNET_TOKEN",
	}
	expectedRequiredVars := []string{
		"RIIDO_AI_SERVER_AWS_REGION",
		"RIIDO_AI_SERVER_ECR_REPOSITORY",
		"RIIDO_AI_SERVER_ECS_CLUSTER",
		"RIIDO_AI_SERVER_ECS_SERVICE",
		"RIIDO_AI_SERVER_ECS_CONTAINER_NAME",
		"RIIDO_AI_SERVER_TESTNET_BASE_URL",
	}
	expectedOptionalVars := []string{
		"RIIDO_AI_SERVER_TESTNET_WORKSPACE_ID",
		"RIIDO_AI_SERVER_CODEDEPLOY_APPLICATION",
		"RIIDO_AI_SERVER_CODEDEPLOY_DEPLOYMENT_GROUP",
	}
	expectedCDKeys := append(append([]string{}, expectedSecrets...), append(expectedRequiredVars, expectedOptionalVars...)...)
	requireStringSetExact(t, parsed.PublicConfigKeyMinimization.RequiredSecretKeys, expectedSecrets)
	requireStringSetExact(t, parsed.PublicConfigKeyMinimization.RequiredVariableKeys, expectedRequiredVars)
	requireStringSetExact(t, parsed.PublicConfigKeyMinimization.OptionalVariableKeys, expectedOptionalVars)
	requireSliceContains(t, parsed.PublicConfigKeyMinimization.StableInfraSourceNames, "ecr_repository_name")
	requireSliceContains(t, parsed.PublicConfigKeyMinimization.StableInfraSourceNames, "ecs_cluster_name")
	requireSliceContains(t, parsed.PublicConfigKeyMinimization.StableInfraSourceNames, "service_name")
	requireSliceContains(t, parsed.PublicConfigKeyMinimization.StableInfraSourceNames, "container_name")
	requireSliceContains(t, parsed.PublicConfigKeyMinimization.StableInfraSourceNames, "codedeploy_application_name")
	requireSliceContains(t, parsed.PublicConfigKeyMinimization.StableInfraSourceNames, "codedeploy_deployment_group_name")
	requireSliceContains(t, parsed.PublicConfigKeyMinimization.PublicDocsMayReference, "required or optional key categories")
	requireSliceContains(t, parsed.PublicConfigKeyMinimization.PublicDocsMayReference, "the machine-readable manifest path that contains the exact key list")
	requireSliceContains(t, parsed.PublicConfigKeyMinimization.PublicDocsMustNotRef, "exact deploy or smoke key-name lists outside the machine-readable manifest and workflow files")
	requireSliceContains(t, parsed.PublicConfigKeyMinimization.PublicDocsMustNotRef, "environment-specific example values")
	requireSliceContains(t, parsed.PublicConfigKeyMinimization.PublicDocsMustNotRef, "live hosts")
	requireStringSetExact(t, collectGitHubConfigRefs(deployWorkflow+"\n"+smokeWorkflow, "secrets"), expectedSecrets)
	requireStringSetExact(t, collectGitHubConfigRefs(deployWorkflow+"\n"+smokeWorkflow, "vars"), append(expectedRequiredVars, expectedOptionalVars...))
	if parsed.PublicSensitiveSurfaceGuard.RiidoTask != "RIID-4842" {
		t.Fatalf("public sensitive surface guard work unit drifted: %#v", parsed.PublicSensitiveSurfaceGuard)
	}
	if parsed.PublicSensitiveSurfaceGuard.CanonicalOwner != "riido-control-plane" ||
		parsed.PublicSensitiveSurfaceGuard.InfraAwarenessOwner != "riido-infra" {
		t.Fatalf("public sensitive surface guard ownership drifted: %#v", parsed.PublicSensitiveSurfaceGuard)
	}
	if !parsed.PublicSensitiveSurfaceGuard.PublicKeyNamesAreSensitive {
		t.Fatal("public sensitive surface guard must treat public key names as sensitive")
	}
	requireContains(t, parsed.PublicSensitiveSurfaceGuard.Rule, "sensitivity budget")
	requireSliceContains(t, parsed.PublicSensitiveSurfaceGuard.AllowedPublicInformation, "current stable RIIDO_AI_SERVER_* key names listed in public_config_key_minimization only in canonical machine-readable and workflow paths")
	requireSliceContains(t, parsed.PublicSensitiveSurfaceGuard.AllowedPublicInformation, "explicit non-CD runtime key exceptions listed in this guard")
	requireNonCDRuntimeKey(t, parsed.PublicSensitiveSurfaceGuard.AllowedPublicNonCDRuntime, "RIIDO_AI_SERVER_AI_AGENT_CLIENT_DEVELOPMENT", "not a deploy/smoke GitHub configuration key")
	requireNonCDRuntimeKey(t, parsed.PublicSensitiveSurfaceGuard.AllowedPublicNonCDRuntime, "RIIDO_AI_SERVER_AI_AGENT_CLIENT_DYNAMODB_TABLE", "not a deploy/smoke GitHub configuration key")
	requireNonCDRuntimeKey(t, parsed.PublicSensitiveSurfaceGuard.AllowedPublicNonCDRuntime, "RIIDO_AI_SERVER_DYNAMODB_ENDPOINT", "not a deploy/smoke GitHub configuration key")
	requireNonCDRuntimeKey(t, parsed.PublicSensitiveSurfaceGuard.AllowedPublicNonCDRuntime, "RIIDO_AI_SERVER_ADDR", "not a deploy/smoke GitHub configuration key")
	requireSliceContains(t, parsed.PublicSensitiveSurfaceGuard.KeyNameScopePaths, "README.md")
	requireSliceContains(t, parsed.PublicSensitiveSurfaceGuard.KeyNameScopePaths, "docs/30-architecture/runtime-cd-ownership.md")
	requireSliceContains(t, parsed.PublicSensitiveSurfaceGuard.KeyNameScopePaths, ".github/workflows/deploy-ai-agent-testnet.yml")
	requireSliceContains(t, parsed.PublicSensitiveSurfaceGuard.CanonicalCDKeyListPaths, "docs/30-architecture/runtime-cd-ownership.riido.json")
	if stringSliceContains(parsed.PublicSensitiveSurfaceGuard.CanonicalCDKeyListPaths, "docs/30-architecture/runtime-cd-ownership.md") ||
		stringSliceContains(parsed.PublicSensitiveSurfaceGuard.CanonicalCDKeyListPaths, "docs/30-architecture/runtime-deployment-boundary.md") {
		t.Fatalf("human-readable docs must not be canonical exact CD key-list paths: %#v", parsed.PublicSensitiveSurfaceGuard.CanonicalCDKeyListPaths)
	}
	requireSliceContains(t, parsed.PublicSensitiveSurfaceGuard.CanonicalCDKeyListPaths, ".github/workflows/deploy-ai-agent-testnet.yml")
	requireSliceContains(t, parsed.PublicSensitiveSurfaceGuard.BroadSummaryDocsMustLink, "README.md")
	requireSliceContains(t, parsed.PublicSensitiveSurfaceGuard.BroadSummaryDocsMustLink, "docs/20-domain/ai-agent-client-api.md")
	requireSliceContains(t, parsed.PublicSensitiveSurfaceGuard.BroadSummaryDocsMustLink, "docs/30-architecture/runtime-deployment-boundary.md")
	requireSliceContains(t, parsed.PublicSensitiveSurfaceGuard.BroadSummaryDocsMustLink, "docs/30-architecture/runtime-cd-ownership.md")
	requireSliceContains(t, parsed.PublicSensitiveSurfaceGuard.BroadSummaryDocsMustLink, "docs/migration/control-plane.md")
	requireSliceContains(t, parsed.PublicSensitiveSurfaceGuard.ForbiddenPublicInformation, "new RIIDO_AI_SERVER_* key names that are not listed in public_config_key_minimization")
	requireSliceContains(t, parsed.PublicSensitiveSurfaceGuard.ForbiddenPublicInformation, "exact deploy/smoke key-name lists in human-readable public docs outside canonical machine-readable and workflow paths")
	requireSliceContains(t, parsed.PublicSensitiveSurfaceGuard.ForbiddenPublicInformation, "example values for allowed public key names")
	requireSliceContains(t, parsed.PublicSensitiveSurfaceGuard.InfraMustKnow, "CD execution remains owned by riido-control-plane")
	requireSliceContains(t, parsed.PublicSensitiveSurfaceGuard.InfraMustKnow, "infra consumes the stable key categories and source names only")
	requireSliceContains(t, parsed.PublicSensitiveSurfaceGuard.InfraMustKnow, "human-readable public docs link to the manifest instead of repeating exact deploy/smoke key lists")
	if parsed.PublicOperationalDetailMinimization.RiidoTask != "RIID-4853" {
		t.Fatalf("public operational detail minimization work unit drifted: %#v", parsed.PublicOperationalDetailMinimization)
	}
	if parsed.PublicOperationalDetailMinimization.CanonicalOwner != "riido-control-plane" ||
		parsed.PublicOperationalDetailMinimization.InfraAwarenessOwner != "riido-infra" {
		t.Fatalf("public operational detail minimization ownership drifted: %#v", parsed.PublicOperationalDetailMinimization)
	}
	requireContains(t, parsed.PublicOperationalDetailMinimization.Rule, "smallest useful CD description")
	requireContains(t, parsed.PublicOperationalDetailMinimization.Rule, "Stable non-secret operational details")
	requireSliceContains(t, parsed.PublicOperationalDetailMinimization.PublicRepoMayKeep, "workflow names and trigger policy")
	requireSliceContains(t, parsed.PublicOperationalDetailMinimization.PublicRepoMayKeep, "stable infra source names without values")
	requireSliceContains(t, parsed.PublicOperationalDetailMinimization.PublicRepoShouldAvoid, "exact deploy or smoke key-name lists outside the machine-readable manifest and workflow files")
	requireSliceContains(t, parsed.PublicOperationalDetailMinimization.PublicRepoShouldAvoid, "duplicating operational setup details in broad README, client-facing docs, generated-client docs, or PR prose when a link to the manifest is sufficient")
	requireSliceContains(t, parsed.PublicOperationalDetailMinimization.InfraMustKnow, "the CD ownership remodel is settled: runtime artifact CD remains in riido-control-plane")
	requireSliceContains(t, parsed.PublicOperationalDetailMinimization.InfraMustKnow, "tightening public operational disclosure is Terraform no-diff unless a future SSOT asks for topology, secret, IAM, network, persistence, or evidence tooling changes")
	if parsed.CodeDeployActivationGate.RiidoTask != "RIID-4855" {
		t.Fatalf("CodeDeploy activation gate work unit drifted: %#v", parsed.CodeDeployActivationGate)
	}
	if parsed.CodeDeployActivationGate.CanonicalOwner != "riido-control-plane" ||
		parsed.CodeDeployActivationGate.InfraAwarenessOwner != "riido-infra" {
		t.Fatalf("CodeDeploy activation gate ownership drifted: %#v", parsed.CodeDeployActivationGate)
	}
	if parsed.CodeDeployActivationGate.Status != "topology-ready-operator-environment-gated" {
		t.Fatalf("CodeDeploy activation gate status drifted: %#v", parsed.CodeDeployActivationGate)
	}
	requireContains(t, parsed.CodeDeployActivationGate.Rule, "not an infra-owned deployment action")
	requireContains(t, parsed.CodeDeployActivationGate.Rule, "operators map the infra-provided application/deployment-group names")
	requireSliceContains(t, parsed.CodeDeployActivationGate.ActivationRequirements, "riido-infra CodeDeploy topology work unit has been applied and reviewed outside public repositories")
	requireSliceContains(t, parsed.CodeDeployActivationGate.ActivationRequirements, "both optional CodeDeploy GitHub environment variables are configured together")
	requireSliceContains(t, parsed.CodeDeployActivationGate.ActivationRequirements, "the deploy workflow creates and waits for the CodeDeploy deployment in the same job")
	requireSliceContains(t, parsed.CodeDeployActivationGate.PublicRepoMayKeep, "stable activation key categories")
	requireSliceContains(t, parsed.CodeDeployActivationGate.PublicRepoMayKeep, "aggregate activation readiness or pass-fail status without live payload values")
	requireSliceContains(t, parsed.CodeDeployActivationGate.PublicRepoMustNotKeep, "environment-specific CodeDeploy application or deployment-group values")
	requireSliceContains(t, parsed.CodeDeployActivationGate.PublicRepoMustNotKeep, "generated CodeDeploy AppSpec or request JSON")
	requireSliceContains(t, parsed.CodeDeployActivationGate.PublicRepoMustNotKeep, "workflow run URL as deploy evidence")
	requireSliceContains(t, parsed.CodeDeployActivationGate.PublicRepoMustNotKeep, "Terraform plan, state, tfvars, apply logs, or raw operator evidence")
	requireSliceContains(t, parsed.CodeDeployActivationGate.InfraMustKnow, "activation does not move create/wait/smoke execution out of riido-control-plane")
	requireSliceContains(t, parsed.CodeDeployActivationGate.InfraMustKnow, "public repos should not request convenience handoff payloads from infra or the deploy workflow")
	for _, repoPath := range parsed.PublicSensitiveSurfaceGuard.BroadSummaryDocsMustLink {
		body := mustRead(t, filepath.Join("..", "..", filepath.FromSlash(repoPath)))
		for _, key := range expectedCDKeys {
			if strings.Contains(body, key) {
				t.Fatalf("%s must link to runtime CD ownership instead of listing CD key %q", repoPath, key)
			}
		}
	}
	requireContains(t, doc, "Public Export Contract")
	requireContains(t, doc, "RIID-4835")
	requireContains(t, doc, "Public Surface Scan")
	requireContains(t, doc, "RIID-4836")
	requireContains(t, doc, "Public Config Key Minimization")
	requireContains(t, doc, "RIID-4839")
	requireContains(t, doc, "Public Sensitive Surface Guard")
	requireContains(t, doc, "RIID-4842")
	requireContains(t, doc, "RIID-4844")
	requireContains(t, doc, "RIID-4845")
	requireContains(t, doc, "Public Operational Detail Minimization")
	requireContains(t, doc, "RIID-4853")
	requireContains(t, doc, "RIID-4855")
	requireContains(t, doc, "operator/environment gated")
	requireContains(t, doc, "infra is the same ownership rule")
	requireContains(t, doc, "Image values are deliberately not in that public export set")
	requireContains(t, boundary, "RIID-4835")
	requireContains(t, boundary, "RIID-4839")
	requireContains(t, boundary, "RIID-4842")
	requireContains(t, boundary, "RIID-4845")
	requireContains(t, boundary, "RIID-4853")
	requireContains(t, boundary, "aggregate deploy/smoke pass-fail status without live payload values")
	requireContains(t, boundary, "are not public hand-off artifacts")
	requireContains(t, readme, "RIID-4839")
	requireContains(t, readme, "RIID-4842")
	requireContains(t, readme, "runtime-cd-ownership.md")
	requireContains(t, domain, "RIID-4839")
	requireContains(t, domain, "RIID-4842")
	requireContains(t, migration, "RIID-4839")
	requireContains(t, migration, "RIID-4842")
	requireContains(t, migration, "RIID-4844")
	requireContains(t, migration, "RIID-4845")
	requireContains(t, migration, "RIID-4853")
	requireContains(t, migration, "RIID-4855")
	requireNotContains(t, doc+"\n"+boundary+"\n"+integration, "public image digest")
	requireContains(t, parsed.DependencyDirection.TopDown, "control-plane")
	requireContains(t, parsed.DependencyDirection.BottomUp, "Infra")

	for _, body := range []string{doc, boundary, integration} {
		requireContains(t, body, "CodeDeploy")
		requireContains(t, body, "riido-control-plane")
		requireContains(t, body, "riido-infra")
	}
}

func TestGeneratedClientDeliveryTokenBoundary(t *testing.T) {
	workflow := mustRead(t, "../../.github/workflows/generated-client-delivery.yml")
	clientDelivery := mustRead(t, "../../docs/30-architecture/api-client-delivery.md")
	migration := mustRead(t, "../../docs/migration/control-plane.md")

	requireContains(t, workflow, "actions/create-github-app-token@v1")
	requireContains(t, workflow, "RIIDO_CLIENT_DELIVERY_APP_ID")
	requireContains(t, workflow, "RIIDO_CLIENT_DELIVERY_PRIVATE_KEY")
	requireContains(t, workflow, "RIIDO_CLIENT_DELIVERY_TOKEN")
	requireContains(t, workflow, "Generated client delivery needs cross-repository write permission")
	requireContains(t, workflow, "github.event.inputs.create_pr == 'true'")
	requireContains(t, workflow, "target_branch must be the Riido work branchName")
	requireContains(t, workflow, "grep -Eq '^[A-Z][A-Z0-9]*-[0-9]+-.+'")

	requireNotContains(t, workflow, "RIIDO_CLIENT_DELIVERY_TOKEN is required to open or update teamswyg/riido-client PRs.")
	requireNotContains(t, workflow, "react-query-")

	requireContains(t, clientDelivery, "create_pr=false")
	requireContains(t, clientDelivery, "requiring `riido-client` write credentials")
	requireContains(t, clientDelivery, "This failure is intentional only for `create_pr=true`")
	requireContains(t, clientDelivery, "Riido `branchName`")
	requireContains(t, clientDelivery, "secret gates")
	requireContains(t, migration, "RIID-4899")
	requireContains(t, migration, "legacy delivery workflow")
	requireContains(t, migration, "raw `RIIDO_CLIENT_DELIVERY_TOKEN`")
	requireContains(t, migration, "synthesize `react-query-*` branch names")
}

func TestRuntimeCDPublicSurfaceScanContract(t *testing.T) {
	manifest := mustRead(t, "../../docs/30-architecture/runtime-cd-ownership.riido.json")
	var parsed struct {
		PublicSurfaceScan struct {
			RiidoTask                  string   `json:"riido_task"`
			ScopePaths                 []string `json:"scope_paths"`
			ForbiddenLiterals          []string `json:"forbidden_literals"`
			ForbiddenRegexes           []string `json:"forbidden_regexes"`
			WorkflowForbiddenMechanism []string `json:"workflow_forbidden_mechanisms"`
		} `json:"public_surface_scan_contract"`
		PublicConfigKeyMinimization struct {
			RequiredSecretKeys   []string `json:"required_secret_keys"`
			RequiredVariableKeys []string `json:"required_variable_keys"`
			OptionalVariableKeys []string `json:"optional_variable_keys"`
		} `json:"public_config_key_minimization"`
		PublicSensitiveSurfaceGuard struct {
			RiidoTask                  string   `json:"riido_task"`
			PublicKeyNamesAreSensitive bool     `json:"public_key_names_are_sensitive"`
			KeyNameScopePaths          []string `json:"key_name_scope_paths"`
			BroadSummaryDocsMustLink   []string `json:"broad_summary_docs_must_link_not_list_cd_keys"`
			AllowedPublicNonCDRuntime  []struct {
				Name string `json:"name"`
			} `json:"allowed_public_non_cd_runtime_keys"`
		} `json:"public_sensitive_surface_guard"`
	}
	if err := json.Unmarshal([]byte(manifest), &parsed); err != nil {
		t.Fatalf("decode runtime CD ownership manifest: %v", err)
	}
	if parsed.PublicSurfaceScan.RiidoTask != "RIID-4836" {
		t.Fatalf("public surface scan task drifted: %#v", parsed.PublicSurfaceScan)
	}
	if parsed.PublicSensitiveSurfaceGuard.RiidoTask != "RIID-4842" || !parsed.PublicSensitiveSurfaceGuard.PublicKeyNamesAreSensitive {
		t.Fatalf("public sensitive surface guard drifted: %#v", parsed.PublicSensitiveSurfaceGuard)
	}
	allowedPublicKeys := append(
		append([]string{}, parsed.PublicConfigKeyMinimization.RequiredSecretKeys...),
		append(parsed.PublicConfigKeyMinimization.RequiredVariableKeys, parsed.PublicConfigKeyMinimization.OptionalVariableKeys...)...,
	)
	for _, key := range parsed.PublicSensitiveSurfaceGuard.AllowedPublicNonCDRuntime {
		allowedPublicKeys = append(allowedPublicKeys, key.Name)
	}

	for _, repoPath := range parsed.PublicSurfaceScan.ScopePaths {
		body := mustRead(t, filepath.Join("..", "..", filepath.FromSlash(repoPath)))
		for _, forbidden := range parsed.PublicSurfaceScan.ForbiddenLiterals {
			if strings.Contains(body, forbidden) {
				t.Fatalf("%s contains forbidden public CD literal %q", repoPath, forbidden)
			}
		}
		for _, pattern := range parsed.PublicSurfaceScan.ForbiddenRegexes {
			re, err := regexp.Compile(pattern)
			if err != nil {
				t.Fatalf("compile public CD forbidden regex %q: %v", pattern, err)
			}
			if match := re.FindString(body); match != "" {
				t.Fatalf("%s contains forbidden public CD pattern %q via %q", repoPath, pattern, match)
			}
		}
	}

	for _, repoPath := range parsed.PublicSensitiveSurfaceGuard.KeyNameScopePaths {
		body := mustRead(t, filepath.Join("..", "..", filepath.FromSlash(repoPath)))
		for _, key := range collectRiidoAIServerKeyLiterals(body) {
			if !stringSliceContains(allowedPublicKeys, key) {
				t.Fatalf("%s contains unregistered public CD configuration key %q", repoPath, key)
			}
		}
	}
	for _, repoPath := range parsed.PublicSensitiveSurfaceGuard.BroadSummaryDocsMustLink {
		body := mustRead(t, filepath.Join("..", "..", filepath.FromSlash(repoPath)))
		for _, key := range allowedPublicKeys {
			if stringSliceContains(collectNonCDRuntimeKeyNames(parsed.PublicSensitiveSurfaceGuard.AllowedPublicNonCDRuntime), key) {
				continue
			}
			if strings.Contains(body, key) {
				t.Fatalf("%s must not repeat public CD configuration key %q", repoPath, key)
			}
		}
	}

	for _, workflowPath := range []string{
		".github/workflows/deploy-ai-agent-testnet.yml",
		".github/workflows/ai-agent-client-testnet-smoke.yml",
	} {
		body := mustRead(t, filepath.Join("..", "..", filepath.FromSlash(workflowPath)))
		for _, forbidden := range parsed.PublicSurfaceScan.WorkflowForbiddenMechanism {
			if strings.Contains(body, forbidden) {
				t.Fatalf("%s contains forbidden public CD handoff mechanism %q", workflowPath, forbidden)
			}
		}
	}
}

func collectGitHubConfigRefs(body, namespace string) []string {
	re := regexp.MustCompile(`\$\{\{\s*` + regexp.QuoteMeta(namespace) + `\.([A-Z0-9_]+)\s*\}\}`)
	matches := re.FindAllStringSubmatch(body, -1)
	seen := map[string]bool{}
	var refs []string
	for _, match := range matches {
		if len(match) < 2 || seen[match[1]] {
			continue
		}
		seen[match[1]] = true
		refs = append(refs, match[1])
	}
	return refs
}

func collectRiidoAIServerKeyLiterals(body string) []string {
	re := regexp.MustCompile(`RIIDO_AI_SERVER_[A-Z0-9_]+`)
	matches := re.FindAllString(body, -1)
	seen := map[string]bool{}
	var keys []string
	for _, match := range matches {
		if seen[match] {
			continue
		}
		seen[match] = true
		keys = append(keys, match)
	}
	return keys
}

func collectNonCDRuntimeKeyNames(keys []struct {
	Name string `json:"name"`
}) []string {
	names := make([]string, 0, len(keys))
	for _, key := range keys {
		names = append(names, key.Name)
	}
	return names
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(body)
}

func decodeStrictJSONDocument(t *testing.T, name, body string, dst any) {
	t.Helper()

	decoder := json.NewDecoder(bytes.NewReader([]byte(body)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		t.Fatalf("decode %s: %v", name, err)
	}

	var trailing struct{}
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		t.Fatalf("%s must contain exactly one JSON document", name)
	}
}

func requireContains(t *testing.T, body, want string) {
	t.Helper()
	if !strings.Contains(body, want) {
		t.Fatalf("missing %q", want)
	}
}

func requireNotContains(t *testing.T, body, want string) {
	t.Helper()
	if strings.Contains(body, want) {
		t.Fatalf("unexpected %q", want)
	}
}

func requireSliceContains(t *testing.T, items []string, want string) {
	t.Helper()
	for _, item := range items {
		if item == want {
			return
		}
	}
	t.Fatalf("missing %q in %#v", want, items)
}

func requireStringSetExact(t *testing.T, got, want []string) {
	t.Helper()
	gotSet := map[string]int{}
	for _, item := range got {
		gotSet[item]++
		if gotSet[item] > 1 {
			t.Fatalf("duplicate item %q in %#v", item, got)
		}
	}
	wantSet := map[string]bool{}
	for _, item := range want {
		wantSet[item] = true
		if gotSet[item] == 0 {
			t.Fatalf("missing %q in %#v", item, got)
		}
	}
	for item := range gotSet {
		if !wantSet[item] {
			t.Fatalf("unexpected %q in %#v, expected %#v", item, got, want)
		}
	}
}

func requireNonCDRuntimeKey(t *testing.T, keys []struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}, wantName, wantReason string) {
	t.Helper()
	for _, key := range keys {
		if key.Name == wantName {
			requireContains(t, key.Reason, wantReason)
			return
		}
	}
	t.Fatalf("missing non-CD runtime key %q in %#v", wantName, keys)
}

func stringSliceContains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
