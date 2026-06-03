# AI Agent Generated Endpoint Smoke Matrix

> Riido task: RIID-4897 `[Control Plane] AI Agent generated endpoint smoke matrix 검증`

This file owns the local control-plane rule for proving that every generated AI
Agent client endpoint has HTTP smoke evidence before it is handed off to
`riido-client`.

## Purpose

Frontend developers search and discuss generated calls by facade paths such as
`riido.v2.aiAgent.tasks.threadMessages.create` and
`riido.v2.aiAgent.agents.daemon.stop`. Those names are projected from the AI
Agent client OpenAPI `x-riido-client.generated_path` metadata and then carried
into generated TypeScript comments and facade paths.

The smoke matrix keeps that generated-client surface honest. Whenever OpenAPI
adds, removes, or renames an operation with `x-riido-client.generated_path`, the
matrix must change in the same PR and the HTTP smoke tests must prove that the
local handler still accepts the endpoint.

## Ownership

| Artifact | Owns | Does not own |
| --- | --- | --- |
| `control-plane-ai-agent-client.openapi.json` | HTTP method/path, schemas, `x-riido-client.generated_path` projection | local test coverage evidence |
| `control-plane-ai-agent-client.smoke-matrix.riido.json` | generated-path to smoke-test evidence index | endpoint shape, response schema, UI behavior |
| `ai_agent_client_generated_smoke_test.go` | executable HTTP smoke coverage and matrix drift check | generated TypeScript output, frontend composition policy |

The matrix is intentionally not another API SSOT. It repeats method/path only so
`TestAIAgentGeneratedEndpointSmokeMatrixMatchesOpenAPI` can fail when it drifts
from OpenAPI.

## Gate

The executable gate is:

```bash
go test ./internal/riidoaiserver -run 'GeneratedEndpointSmoke|SmokeMatrix' -count=1
```

The gate enforces:

- every OpenAPI operation with `x-riido-client.generated_path` appears exactly
  once in `control-plane-ai-agent-client.smoke-matrix.riido.json`
- every matrix entry has the same HTTP method and path as OpenAPI
- every matrix entry names a known smoke evidence test
- v1 generated paths cite the v1 HTTP smoke test
- v2 generated paths cite the v2 HTTP smoke test
- matrix entries stay sorted by generated path for reviewable diffs

The v1 smoke test covers the legacy `/v1/client/ai-agent/**` compatibility
surface. The v2 smoke test covers the workspace-scoped
`/v2/client/workspaces/{workspace_id}/ai-agent/**` surface.

## Update Rule

When an AI Agent generated endpoint changes:

1. Update the DSL/IR/OpenAPI projection that owns the endpoint.
2. Update `control-plane-ai-agent-client.smoke-matrix.riido.json` with the
   generated path, method, path, and evidence test.
3. Add or extend HTTP smoke coverage in
   `internal/riidoaiserver/ai_agent_client_generated_smoke_test.go`.
4. Run the gate above and the generated-client generator tests before delivery.

Do not add matrix entries for Figma-only composition facts, UI labels, sorting
rules, or frontend route decisions. Those stay in Figma projection and client
composition SSOTs.
