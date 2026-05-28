// Package riidoaiserver owns the public Riido SaaS control-plane domain and
// adapter boundary.
//
// The current migration slices contain stdlib-only assignment contracts and
// DTOs, assignment operation journal ports, agent catalog RBAC, agent catalog
// API ports and HTTP adapter, request authorization adapters, agent/runtime
// binding guards, provider status contracts, the in-memory assignment store
// actor, review account seed provisioning, store snapshot/file outbox adapters,
// HTTP/SSE/metrics/health adapters, browser frontend CORS transport
// allowlisting, the stdout CloudWatch EMF metrics publisher, and stdlib-only
// DynamoDB/EventBridge adapter request construction that can be verified with
// fake endpoints. This package does not own daemon provider process execution,
// cloud deployment wiring, CloudWatch API calls, dashboards, live AWS evidence,
// Terraform, or production secret values.
package riidoaiserver
