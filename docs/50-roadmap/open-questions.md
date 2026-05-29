# Control Plane Open Questions

> Riido task: RIID-4712 `[Control Plane] Architecture SSOT docs migration`

This file owns unresolved public control-plane decisions. Other docs link here
instead of redefining these questions.

| ID | Area | Question | Current working stance |
| --- | --- | --- | --- |
| Q-CP-001 | AWS adapter dependency | Should public control-plane ever adopt an external AWS SDK? | No. Stdlib-only adapters remain until an ADR accepts the dependency. |
| Q-CP-002 | Agent catalog durability | Should agent catalog persistence use the assignment operation journal, a new catalog journal, or a DynamoDB single-table projection? | Public in-memory store remains the executable behavior; durable production mapping belongs in a future slice. |
| Q-CP-003 | Production identity | Which IdP/JWKS claim mapping becomes the production authorizer contract? | External authorizer port is stable; tenant claim mapping remains outside public defaults. |
| Q-CP-004 | Metrics export | Should stdout EMF remain the only public metrics publisher, or should Prometheus/OpenTelemetry become a contract? | Stdout EMF is the only public publisher. Dashboards and exporters stay infra-owned. |
| Q-CP-005 | Stream relay runtime | Should DynamoDB stream relay run in `cmd/riido_ai_server` or as a separate worker process? | Adapter core is public; runtime topology is infra/deployment-owned until decided. |
| Q-CP-006 | Review/demo lifecycle | How are review/demo accounts rotated, disabled, and audited in production? | Public repo stores only safe seed shape and token-hash provisioning; operations lifecycle is private evidence. |
| Q-CP-007 | Runtime waitlist mutation | Should the Windows app waitlist and marketing-consent action from Figma `node-id=275-22731` be implemented in the AI Agent control-plane client API or delegated to an existing product/user-marketing API? | Unresolved. Current generated client exposes only `GET /v1/client/ai-agent/devices` for runtime empty-state composition; no waitlist/marketing mutation is generated until the owning SSOT is chosen. |
