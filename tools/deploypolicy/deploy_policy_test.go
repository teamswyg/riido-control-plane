package deploypolicy

import (
	"encoding/json"
	"os"
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
	migration := mustRead(t, "../../docs/migration/control-plane.md")
	generator := mustRead(t, "../../tools/reactquerygen/main.go")
	generatedClient := mustRead(t, "../../web/generated/aiAgentClient.ts")

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
		"README.md":                      readme,
		"runtime-deployment-boundary.md": boundary,
		"saas-control-plane.md":          domain,
		"ai-agent-client-api.md":         clientAPI,
		"control-plane.md":               migration,
		"tools/reactquerygen/main.go":    generator,
		"web/generated/aiAgentClient.ts": generatedClient,
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

	var parsed struct {
		SchemaVersion string   `json:"schema_version"`
		ID            string   `json:"id"`
		RiidoTask     string   `json:"riido_task"`
		Hardening     []string `json:"hardening_tasks"`
		Runtime       string   `json:"runtime_service"`
		Current       struct {
			ID            string   `json:"id"`
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
			ActivationInputs []string `json:"activation_inputs"`
			Allowed          []string `json:"allowed_actions"`
			MustNotOwn       []string `json:"must_not_own"`
		} `json:"optional_workflow_modes"`
		Future []struct {
			ID                 string   `json:"id"`
			CDOwner            string   `json:"cd_owner"`
			TopologyOwner      string   `json:"topology_owner"`
			ControlPlaneMayOwn []string `json:"control_plane_may_own"`
			InfraMustOwn       []string `json:"infra_must_own"`
		} `json:"future_strategies"`
		Redaction struct {
			Workflows      []string `json:"public_repo_workflows"`
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
	if err := json.Unmarshal([]byte(manifest), &parsed); err != nil {
		t.Fatalf("decode runtime CD ownership manifest: %v", err)
	}

	if parsed.SchemaVersion != "riido-control-plane-runtime-cd-ownership.v1" {
		t.Fatalf("unexpected schema version: %q", parsed.SchemaVersion)
	}
	if parsed.ID != "runtime-cd-ownership" || parsed.RiidoTask != "RIID-4825" || parsed.Runtime != "riido_ai_server" {
		t.Fatalf("manifest identity drifted: %#v", parsed)
	}
	requireSliceContains(t, parsed.Hardening, "RIID-4833")
	requireSliceContains(t, parsed.Hardening, "RIID-4835")
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
	requireSliceContains(t, parsed.Redaction.ShouldMinimize, "publish only stable configuration key names that operators must set")
	requireSliceContains(t, parsed.Redaction.ShouldMinimize, "centralize required deploy key names in the workflow and ownership docs instead of scattering environment-specific examples")
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
	if parsed.InfraVisibility.Repo != "riido-infra" {
		t.Fatalf("infra visibility repo drifted: %q", parsed.InfraVisibility.Repo)
	}
	requireSliceContains(t, parsed.InfraVisibility.MustKnow, "riido-control-plane owns runtime artifact CD execution")
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
	requireContains(t, doc, "Public Export Contract")
	requireContains(t, doc, "RIID-4835")
	requireContains(t, doc, "Image values are deliberately not in that public export set")
	requireContains(t, boundary, "RIID-4835")
	requireContains(t, boundary, "aggregate deploy/smoke pass-fail status without live payload values")
	requireContains(t, boundary, "are not public hand-off artifacts")
	requireNotContains(t, doc+"\n"+boundary+"\n"+integration, "public image digest")
	requireContains(t, parsed.DependencyDirection.TopDown, "control-plane")
	requireContains(t, parsed.DependencyDirection.BottomUp, "Infra")

	for _, body := range []string{doc, boundary, integration} {
		requireContains(t, body, "CodeDeploy")
		requireContains(t, body, "riido-control-plane")
		requireContains(t, body, "riido-infra")
	}
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(body)
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
