package main

import "slices"

const (
	syntaxHashLoopID       = "syntax_hash_graph_spike_loop"
	syntaxHashGoldenSource = `go test ./tools/contextmap ./tools/syntaxhash ./awsadapters ./cmd/riido_ai_server ./internal/riidoaiserver ./internal/repoidentity ./internal/contractscompat ./tools/requestauth ./tools/healthreadycmd ./tools/runtimeboundary ./tools/runtimecdownership ./tools/deploypolicy ./tools/containercontract ./tools/containercontract/dockerfile ./tools/configreference ./tools/storesaferouting ./tools/publicpageslive ./tools/gocibaseline ./tools/moduledecomposition ./tools/controlplaneaudit ./tools/controlplaneperf ./tools/controlplanepressure ./tools/aiagentload ./tools/harnesspromotion ./tools/loopclosureaudit ./tools/closedloopcandidateintake ./tools/closedloopcandidatedecision ./tools/looprefreshdispatch ./tools/evidencegraph ./tools/loopregistry ./tools/operationalreadiness ./tools/precommitbaseline ./tools/workflowevidence ./tools/liveworkflowevidence ./tools/aigeneratedsmokematrix ./tools/apiclientdelivery ./tools/figmaprojection ./tools/aiagentclientapi ./tools/generatedclienthandoff ./tools/reactquerygen ./tools/storesnapshotoutbox ./tools/aiagentthreadsnapshot ./tools/webfrontendapi ./tools/aiagentrisk ./tools/migrationledger ./tools/metricshttpadapter ./tools/agentcatalogrbac ./tools/reviewaccountseed ./tools/integrationmatrix ./tools/openquestions ./tools/repositoryreadme ./tools/knowledgecoverage ./tools/dependencyallowlist ./tools/providerstatus ./tools/agentruntimebinding ./tools/assignmentjournal ./tools/cloudwatchemf ./tools/saascontrolplane ./tools/snapshotcqrsgate -run 'Test(ContextMap|SyntaxHashTool|AWSAdaptersFacade|RuntimeConfig|RiidoAIServer|RepoIdentity|ContractsCompat|RequestAuthorization|HealthReadyCommand|RuntimeBoundary|RuntimeCDOwnership|DeployPolicy|ContainerContract|ConfigReference|StoreSafeRouting|PublicPagesLive|GoCIBaseline|ModuleDecomposition|ControlPlaneAudit|ControlPlanePerformance|ControlPlanePressure|AIAgentLoad|HarnessPromotion|LoopClosureAudit|ClosedLoopCandidateIntake|ClosedLoopCandidateDecision|LoopRefreshDispatch|EvidenceGraph|LoopRegistry|OperationalReadiness|PreCommitBaseline|WorkflowEvidence|LiveWorkflowEvidence|RunWrites|APIClientDelivery|FigmaProjection|WebFrontendAPI|AIAgentRisk|MigrationLedger|MetricsHTTPAdapter|AgentCatalogRBAC|ReviewAccountSeed|IntegrationMatrix|OpenQuestions|RepositoryReadme|KnowledgeCoverage|DependencyAllowlist|ProviderStatus|AgentRuntimeBinding|AssignmentJournal|CloudWatchEMF|SaaSControlPlane|SnapshotCQRSGate|AIAgentThreadSnapshot|ReactQueryGen)BehaviorGolden|GeneratedClientHandoff' -count=1`
	syntaxHashGoldenNative = "export RIIDO_LOOP_IDS=" + syntaxHashLoopID + " && " + syntaxHashGoldenSource
	syntaxHashSource       = "mkdir -p out && go test ./tools/syntaxhash -count=1 && go run ./tools/syntaxhash -check-doc -evidence-out out/syntax-hash-evidence.json"
	syntaxHashNative       = "umask 077 && export RIIDO_LOOP_IDS=" + syntaxHashLoopID + " && " + syntaxHashSource
)

func verifySyntaxHashWorkflow(repoRoot string, child boundedChild) bool {
	value := child.Baseline
	raw, err := readRootFile(repoRoot, value.Workflow)
	required := []string{
		"name: Syntax Hash", "RIIDO_LOOP_IDS: syntax_hash_graph_spike_loop",
		"actions/checkout@v7", "actions/setup-go@v6", "go-version-file: go.mod", "cache: false",
		"Run golden locks before syntax hash", "go test ./tools/contextmap ./tools/syntaxhash ./awsadapters",
		"GeneratedClientHandoff'", "Verify syntax hash evidence", "go test ./tools/syntaxhash -count=1",
		"go run ./tools/syntaxhash", "-check-doc", "-evidence-out out/syntax-hash-evidence.json",
		"actions/upload-artifact@v7", "if: always()", "name: syntax-hash-evidence",
		"path: out/syntax-hash-evidence.json", "if-no-files-found: error",
	}
	return err == nil && value.SourceRevision == "45f08aa2f91bb4ae0b6e4d7d75a841833dac0435" &&
		value.Workflow == ".github/workflows/syntax-hash.yml" &&
		value.WorkflowSHA256 == "470d347daa1b35680fb4c899ec234e3ead152356c8f38ab07d521fbcdf194ec7" &&
		digest(raw) == value.WorkflowSHA256 && value.WorkflowName == "Syntax Hash" &&
		value.Job == "syntax-hash" && value.TrackedWorkflowCount == 56 && containsAll(string(raw), required)
}

func verifySyntaxHashMapping(child boundedChild) bool {
	mapping, claim := child.NativeMapping, child.ParityClaim
	return mapping.Checkout.Source == "actions/checkout@v7" && mapping.Checkout.NativeKind == "checkout" &&
		!mapping.Checkout.AdapterRequired && mapping.GoToolchain.Source == "actions/setup-go@v6" &&
		mapping.GoToolchain.NativeKind == "worker_toolchain" && mapping.GoToolchain.VersionSource == "go.mod" &&
		mapping.GoToolchain.Version == "1.26.5" && !mapping.GoToolchain.Cache && !mapping.GoToolchain.AdapterRequired &&
		verifyCommand(mapping.SyntaxHashGolden, syntaxHashGoldenSource, syntaxHashGoldenNative, "") &&
		verifyCommand(mapping.SyntaxHash, syntaxHashSource, syntaxHashNative, "out/syntax-hash-evidence.json") &&
		claim.AllSourceStepsMapped && claim.RequiredAdapterCount == 0 && claim.SyntaxHashLoopEnvironmentExact &&
		claim.SyntaxHashGoldenCommandExact && claim.SyntaxHashCommandExact && claim.SecureEvidencePermissions &&
		!claim.SourceWorkflowEdited && !claim.SourceWorkflowExecutedByThisSlice
}

func verifySyntaxHashArtifact(child boundedChild) bool {
	value := child.NativeMapping.EvidenceArtifact
	want := []string{"out/syntax-hash-evidence.json"}
	return value.Source == "actions/upload-artifact@v7" && value.NativeKind == "artifact" &&
		slices.Equal(value.Paths, want) && value.Redaction == "required" && value.RunWhen == "always" &&
		value.IfNoFilesFound == "error" && value.ContentAddressed && !value.AdapterRequired
}
