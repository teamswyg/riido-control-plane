package deploypolicy

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestDeployAIAgentTestnetPublicRedactionPolicy(t *testing.T) {
	workflow := mustRead(t, "../../.github/workflows/deploy-ai-agent-testnet.yml")
	readme := mustRead(t, "../../README.md")
	boundary := mustRead(t, "../../docs/30-architecture/runtime-deployment-boundary.md")
	domain := mustRead(t, "../../docs/20-domain/saas-control-plane.md")
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
	requireContains(t, workflow, "current_json=\"$RUNNER_TEMP/task-definition.current.json\"")
	requireContains(t, workflow, "next_json=\"$RUNNER_TEMP/task-definition.next.json\"")
	requireContains(t, workflow, "appspec_json=\"$RUNNER_TEMP/codedeploy-appspec.json\"")
	requireContains(t, workflow, "deployment_json=\"$RUNNER_TEMP/codedeploy-deployment.json\"")
	requireContains(t, workflow, "revisionType: \"AppSpecContent\"")
	requireContains(t, workflow, "aws deploy create-deployment")
	requireContains(t, workflow, "aws deploy wait deployment-successful")

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

	for path, body := range map[string]string{
		"README.md":                      readme,
		"runtime-deployment-boundary.md": boundary,
		"saas-control-plane.md":          domain,
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
		SchemaVersion string `json:"schema_version"`
		ID            string `json:"id"`
		RiidoTask     string `json:"riido_task"`
		Runtime       string `json:"runtime_service"`
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
			ShouldMinimize []string `json:"public_repo_should_minimize"`
			MustNot        []string `json:"public_repo_must_not_commit_or_upload"`
		} `json:"public_redaction_policy"`
		Infra struct {
			Repo       string   `json:"repo"`
			Paths      []string `json:"paths"`
			LocalScope string   `json:"local_scope"`
		} `json:"infra_consumes"`
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
	if parsed.ID != "runtime-cd-ownership" || parsed.RiidoTask != "RIID-4822" || parsed.Runtime != "riido_ai_server" {
		t.Fatalf("manifest identity drifted: %#v", parsed)
	}
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
	requireSliceContains(t, parsed.Redaction.ShouldMinimize, "publish only stable configuration key names that operators must set")
	requireSliceContains(t, parsed.Redaction.ShouldMinimize, "avoid environment-specific examples for domains, clusters, services, applications, deployment groups, ARNs, and URLs")
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

func requireSliceContains(t *testing.T, items []string, want string) {
	t.Helper()
	for _, item := range items {
		if item == want {
			return
		}
	}
	t.Fatalf("missing %q in %#v", want, items)
}
